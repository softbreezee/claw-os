package agent

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"

	"github.com/fastclaw-ai/fastclaw/internal/bus"
	"github.com/fastclaw-ai/fastclaw/internal/provider"
)

// buildContentParts turns a user-typed text plus a list of attachments
// into the ContentParts slice that gets sent to the LLM.
//
// Today this only handles images — the design doc keeps the door open
// for documents/audio later, but Stage 1 of the multimodal feature
// limited the scope deliberately. Non-image attachments are mentioned
// to the model as text ("[attached: filename.pdf]") so the agent at
// least knows they were sent and can decide whether to call a tool.
//
// Images are inlined as base64 data: URLs rather than fastclaw-hosted
// HTTP URLs because most upstream LLM APIs need to fetch attachments
// from their own servers, and a localhost gateway URL won't be
// reachable. Inlining sidesteps that whole class of problem.
func buildContentParts(text string, attachments []bus.Attachment) ([]provider.ContentPart, error) {
	if len(attachments) == 0 {
		return nil, nil
	}

	parts := make([]provider.ContentPart, 0, 1+len(attachments))

	// Text first (some providers require text-then-images ordering for
	// best results; OpenAI and Anthropic both accept either, but this
	// matches their docs' examples).
	if text != "" {
		parts = append(parts, provider.ContentPart{Type: "text", Text: text})
	}

	// One ContentPart per image; non-image attachments get a synthetic
	// text breadcrumb so the model can at least talk about them.
	var nonImageNote string
	for _, att := range attachments {
		if att.IsImage() {
			dataURL, err := imageToDataURL(att)
			if err != nil {
				slog.Warn("multimodal: skipping image attachment",
					"path", att.Path, "name", att.Name, "error", err)
				continue
			}
			parts = append(parts, provider.ContentPart{
				Type: "image_url",
				ImageURL: &provider.ImageURL{
					URL:    dataURL,
					Detail: "auto",
				},
			})
		} else {
			nonImageNote += fmt.Sprintf("\n[attached file: %s (%s)]", att.Name, att.MimeType)
		}
	}

	if nonImageNote != "" {
		// Append the breadcrumb to the (possibly empty) text part. We
		// either extend the existing text part or insert a new one at
		// the front; either way, the model sees a single textual frame
		// before any images.
		if len(parts) > 0 && parts[0].Type == "text" {
			parts[0].Text += nonImageNote
		} else {
			parts = append([]provider.ContentPart{{Type: "text", Text: nonImageNote}}, parts...)
		}
	}

	// If we somehow stripped every attachment (e.g. all reads failed)
	// and there's no text either, return nil so the caller falls back
	// to the plain Content path.
	if len(parts) == 0 {
		return nil, nil
	}

	return parts, nil
}

// buildUserMessage constructs the provider.Message for an inbound
// user turn, choosing the right shape (plain Content vs. ContentParts)
// based on which attachment fields are populated. Used by both the
// non-streaming and streaming agent loops to keep the conversion
// logic in one place.
//
// Resolution order:
//   1. Web-UI Attachments      → multimodal ContentParts via buildContentParts
//   2. Telegram-style PhotoURL → single image_url ContentPart (legacy)
//   3. Plain text              → simple Content string
func buildUserMessage(msg bus.InboundMessage) provider.Message {
	if len(msg.Attachments) > 0 {
		parts, err := buildContentParts(msg.Text, msg.Attachments)
		if err != nil {
			slog.Warn("multimodal: build parts failed, falling back to text-only",
				"error", err)
		}
		if len(parts) > 0 {
			return provider.Message{Role: "user", ContentParts: parts}
		}
	}
	if msg.PhotoURL != "" {
		return provider.Message{
			Role: "user",
			ContentParts: []provider.ContentPart{
				{Type: "text", Text: msg.Text},
				{Type: "image_url", ImageURL: &provider.ImageURL{URL: msg.PhotoURL, Detail: "auto"}},
			},
		}
	}
	return provider.Message{Role: "user", Content: msg.Text}
}

// flattenUserContent splits a user provider.Message back into a text
// string and a list of attachment descriptors suitable for JSON
// marshalling to the chat history API. The reverse of buildContentParts.
//
// Each returned attachment is a map with these keys:
//   - "type":  "image" today (room for "document" / "audio" later)
//   - "url":   ready-to-render URL (data:… for inlined images, or
//              http(s):// for hosted)
//
// If the message used the legacy plain `Content` string with no parts,
// we return that string as text and no attachments.
func flattenUserContent(m provider.Message) (string, []map[string]any) {
	if len(m.ContentParts) == 0 {
		return m.Content, nil
	}
	var text string
	var attachments []map[string]any
	for _, p := range m.ContentParts {
		switch p.Type {
		case "text":
			if text != "" {
				text += "\n"
			}
			text += p.Text
		case "image_url":
			if p.ImageURL == nil || p.ImageURL.URL == "" {
				continue
			}
			attachments = append(attachments, map[string]any{
				"type": "image",
				"url":  p.ImageURL.URL,
			})
		}
	}
	return text, attachments
}

// imageToDataURL reads the image off disk and produces a base64
// "data:<mime>;base64,<...>" URL suitable for the OpenAI/Anthropic
// image_url field.
//
// We don't bound the file size here on purpose — that's a config
// concern handled at the upload boundary. By the time we're here, the
// file is already on disk, so refusing it just to refuse it would be
// odd. The LLM API's own size limits will reject anything truly silly.
func imageToDataURL(att bus.Attachment) (string, error) {
	data, err := os.ReadFile(att.Path)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", att.Path, err)
	}
	mime := att.MimeType
	if mime == "" {
		mime = "application/octet-stream"
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	return "data:" + mime + ";base64," + encoded, nil
}
