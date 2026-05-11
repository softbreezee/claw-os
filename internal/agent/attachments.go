package agent

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/fastclaw-ai/fastclaw/internal/bus"
	"github.com/fastclaw-ai/fastclaw/internal/provider"
)

// isVisionModel reports whether the named model can directly consume
// image content parts. We use a conservative allow-list rather than
// a deny-list because mistakes here are expensive: sending an image
// to a non-vision model returns a hard 400 and breaks the whole turn.
//
// When in doubt, NEW models default to non-vision (text breadcrumb path),
// which is the safer failure mode — the agent still knows files were
// attached and can choose to forward them to a vision sub-agent via
// spawn_subagent(forward_attachments=true).
func isVisionModel(model string) bool {
	m := strings.ToLower(provider.StripProviderPrefix(model))
	switch {
	case strings.HasPrefix(m, "claude-"):
		return true
	case strings.HasPrefix(m, "gpt-4o"), strings.HasPrefix(m, "gpt-4-turbo"), strings.HasPrefix(m, "gpt-5"):
		return true
	case strings.HasPrefix(m, "o1"):
		return true
	case strings.HasPrefix(m, "kimi-k2"), strings.HasPrefix(m, "kimi-vl"):
		return true
	case strings.HasPrefix(m, "qwen-vl"), strings.HasPrefix(m, "qwen2-vl"), strings.HasPrefix(m, "qwen2.5-vl"):
		return true
	case strings.HasPrefix(m, "gemini-"):
		return true
	}
	return false
}

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
// based on which attachment fields are populated AND whether the
// receiving model can actually see images.
//
// Resolution order:
//   1. Attachments + vision model → multimodal ContentParts (inline base64)
//   2. Attachments + non-vision model → text breadcrumb listing the files,
//      so the LLM knows they exist and can decide to delegate to a
//      vision sub-agent via spawn_subagent(forward_attachments=true).
//      The actual files stay reachable via AttachmentsFromContext.
//   3. Telegram-style PhotoURL → single image_url ContentPart (legacy)
//   4. Plain text → simple Content string
func buildUserMessage(msg bus.InboundMessage, model string) provider.Message {
	if len(msg.Attachments) > 0 {
		if isVisionModel(model) {
			parts, err := buildContentParts(msg.Text, msg.Attachments)
			if err != nil {
				slog.Warn("multimodal: build parts failed, falling back to text breadcrumb",
					"error", err, "model", model)
			}
			if len(parts) > 0 {
				return provider.Message{Role: "user", ContentParts: parts}
			}
		}
		// Non-vision model (or part-build failure): describe attachments
		// in text. The agent SOUL prompt is expected to teach delegation
		// to a vision-capable sub-agent when this path fires.
		return provider.Message{
			Role:    "user",
			Content: textWithAttachmentBreadcrumb(msg.Text, msg.Attachments),
		}
	}
	if msg.PhotoURL != "" {
		// Legacy single-photo path (Telegram). We don't currently know
		// the receiving agent's model on this code path; assume vision
		// — Telegram bots that wire up photos are typically configured
		// against a vision model anyway.
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

// textWithAttachmentBreadcrumb appends a structured listing of the
// attachments to the user's text so a non-vision LLM is at least aware
// they exist. Format is intentionally machine-friendly so we can
// reference indices later in tool calls (e.g. forward_attachments).
//
// Example output:
//
//   What's in this picture?
//
//   [Attached files]
//   [0] image/png — diagram.png (124 KB)
//   [1] image/jpeg — photo.jpg (480 KB)
//
//   Note: you cannot see images directly. To analyse them, use
//   spawn_subagent with a vision-capable agent and forward_attachments=true.
func textWithAttachmentBreadcrumb(text string, atts []bus.Attachment) string {
	var b strings.Builder
	if text != "" {
		b.WriteString(text)
		b.WriteString("\n\n")
	}
	b.WriteString("[Attached files]\n")
	for i, a := range atts {
		mime := a.MimeType
		if mime == "" {
			mime = "application/octet-stream"
		}
		fmt.Fprintf(&b, "[%d] %s — %s (%d KB)\n", i, mime, a.Name, a.Size/1024)
	}
	b.WriteString("\nNote: this model cannot see images directly. To analyse them, ")
	b.WriteString("call spawn_subagent with a vision-capable agent ")
	b.WriteString("and forward_attachments=true.")
	return b.String()
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
