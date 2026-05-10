package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

// OpenAIProvider implements the Provider interface for OpenAI-compatible APIs.
type OpenAIProvider struct {
	apiKey  string
	apiBase string
	client  *http.Client
}

// NewOpenAI creates a new OpenAI-compatible provider.
func NewOpenAI(apiKey, apiBase string) *OpenAIProvider {
	if apiBase == "" {
		apiBase = "https://api.openai.com/v1"
	}
	apiBase = strings.TrimRight(apiBase, "/")
	return &OpenAIProvider{
		apiKey:  apiKey,
		apiBase: apiBase,
		client:  &http.Client{},
	}
}

// apiMessage is the wire format for a message sent to the OpenAI API.
// It uses json.RawMessage for Content to support both string and array formats.
type apiMessage struct {
	Role             string          `json:"role"`
	Content          json.RawMessage `json:"content,omitempty"`
	ToolCalls        []ToolCall      `json:"tool_calls,omitempty"`
	ToolCallID       string          `json:"tool_call_id,omitempty"`
	Name             string          `json:"name,omitempty"`
	// ReasoningContent is required by DeepSeek thinking models: the chain-of-thought
	// returned in a previous response must be echoed back in history messages.
	ReasoningContent string          `json:"reasoning_content,omitempty"`
}

type chatRequest struct {
	Model    string       `json:"model"`
	Messages []apiMessage `json:"messages"`
	Tools    []Tool       `json:"tools,omitempty"`
	MaxTokens int         `json:"max_tokens,omitempty"`
	// Temperature is a pointer so we can distinguish "use default" (nil,
	// field omitted from JSON) from "explicitly send 0.0". Some upstreams
	// — Kimi-k2, OpenAI reasoning models (o1/o3), GPT-5 family — reject
	// any temperature other than 1 with a hard 400, so we have to be
	// able to omit the field entirely.
	Temperature *float64 `json:"temperature,omitempty"`
	Stream      bool     `json:"stream"`
}

// modelLocksTemperature reports whether the named model rejects all
// temperature values except its mandated default (usually 1.0). For
// these models we omit the temperature field on the wire.
//
// The list is conservative — we'd rather a request succeed with the
// model's default than fail with a confusing 400. New entries should
// be added when a user reports the symptom.
func modelLocksTemperature(model string) bool {
	m := strings.ToLower(model)
	switch {
	case strings.HasPrefix(m, "kimi-k2"):
		return true
	case strings.HasPrefix(m, "o1"), strings.HasPrefix(m, "o3"), strings.HasPrefix(m, "o4"):
		return true
	case strings.HasPrefix(m, "gpt-5"):
		return true
	}
	return false
}

// toAPIMessages converts provider Messages to wire-format apiMessages,
// handling ContentParts for multimodal messages and preserving ReasoningContent
// for DeepSeek thinking models.
func toAPIMessages(msgs []Message) []apiMessage {
	out := make([]apiMessage, len(msgs))
	for i, m := range msgs {
		am := apiMessage{
			Role:             m.Role,
			ToolCalls:        m.ToolCalls,
			ToolCallID:       m.ToolCallID,
			Name:             m.Name,
			ReasoningContent: m.ReasoningContent,
		}
		if len(m.ContentParts) > 0 {
			am.Content, _ = json.Marshal(m.ContentParts)
		} else {
			// Even empty strings must be sent as content when tool_calls are present.
			am.Content, _ = json.Marshal(m.Content)
		}
		out[i] = am
	}
	return out
}

// sseDelta mirrors the OpenAI streaming delta structure including tool call index.
type sseToolCallDelta struct {
	Index    int          `json:"index"`
	ID       string       `json:"id,omitempty"`
	Type     string       `json:"type,omitempty"`
	Function FunctionCall `json:"function"`
}

type sseDelta struct {
	Role             string             `json:"role,omitempty"`
	Content          string             `json:"content,omitempty"`
	ReasoningContent string             `json:"reasoning_content,omitempty"`
	ToolCalls        []sseToolCallDelta `json:"tool_calls,omitempty"`
}

type sseChoice struct {
	Delta        sseDelta `json:"delta"`
	FinishReason string   `json:"finish_reason"`
}

type sseResponse struct {
	Choices []sseChoice `json:"choices"`
}

// buildRequest constructs the HTTP request for /chat/completions.
//
// `omitTemperature` lets callers request a retry that strips the field
// when a previous attempt failed with the model's "temperature must be
// 1" 400. Normal callers pass false and rely on modelLocksTemperature
// to do the right thing automatically.
func (p *OpenAIProvider) buildRequest(ctx context.Context, messages []Message, tools []Tool, model string, maxTokens int, temperature float64, stream bool, omitTemperature bool) (*http.Request, error) {
	bareModel := StripProviderPrefix(model)
	req := chatRequest{
		Model:     bareModel,
		Messages:  toAPIMessages(messages),
		MaxTokens: maxTokens,
		Stream:    stream,
	}
	if !omitTemperature && !modelLocksTemperature(bareModel) {
		t := temperature
		req.Temperature = &t
	}
	if len(tools) > 0 {
		req.Tools = tools
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := p.apiBase + "/chat/completions"
	slog.Info("openai request",
		"url", url,
		"model", req.Model,
		"temperature_sent", req.Temperature != nil,
	)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	return httpReq, nil
}

// isTemperatureLockError sniffs an upstream 400 body to detect the
// "this model only accepts a fixed temperature" rejection. Used to
// retry without the offending field, which both fixes the request
// and tells us we should add this model to modelLocksTemperature.
func isTemperatureLockError(body string) bool {
	b := strings.ToLower(body)
	return strings.Contains(b, "temperature") &&
		(strings.Contains(b, "only 1 is allowed") ||
			strings.Contains(b, "must be 1") ||
			strings.Contains(b, "does not support") ||
			strings.Contains(b, "not supported"))
}

func (p *OpenAIProvider) Chat(ctx context.Context, messages []Message, tools []Tool, model string, maxTokens int, temperature float64) (*Response, error) {
	resp, body, err := p.doChat(ctx, messages, tools, model, maxTokens, temperature, true, false)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		// Self-healing retry: if upstream said "this model only takes
		// temperature=1", drop the field and try once more. We log it
		// loudly so the bare-model name can be added to the static
		// lock list and avoid the wasted round-trip next time.
		if resp.StatusCode == http.StatusBadRequest && isTemperatureLockError(body) {
			slog.Warn("upstream rejected temperature, retrying without it",
				"model", StripProviderPrefix(model),
				"hint", "consider adding this model prefix to modelLocksTemperature",
			)
			resp.Body.Close()
			resp2, body2, err2 := p.doChat(ctx, messages, tools, model, maxTokens, temperature, true, true)
			if err2 != nil {
				return nil, err2
			}
			defer resp2.Body.Close()
			if resp2.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("API error %d: %s", resp2.StatusCode, body2)
			}
			return p.parseSSE(resp2.Body)
		}
		resp.Body.Close()
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, body)
	}
	defer resp.Body.Close()
	return p.parseSSE(resp.Body)
}

// doChat is a thin helper that issues the request and reads the body
// when the status isn't 200 (so callers can inspect it for retry
// decisions). On a 200 the body is left to the caller (parseSSE
// consumes it as a stream).
func (p *OpenAIProvider) doChat(ctx context.Context, messages []Message, tools []Tool, model string, maxTokens int, temperature float64, stream, omitTemperature bool) (*http.Response, string, error) {
	httpReq, err := p.buildRequest(ctx, messages, tools, model, maxTokens, temperature, stream, omitTemperature)
	if err != nil {
		return nil, "", err
	}
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, "", fmt.Errorf("send request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return resp, string(respBody), nil
	}
	return resp, "", nil
}

// openStream is the streaming counterpart to doChat. It returns the
// raw response so the caller can either drain the body (on error) or
// hand it off to the SSE consumer (on success).
func (p *OpenAIProvider) openStream(ctx context.Context, messages []Message, tools []Tool, model string, maxTokens int, temperature float64, omitTemperature bool) (*http.Response, error) {
	httpReq, err := p.buildRequest(ctx, messages, tools, model, maxTokens, temperature, true, omitTemperature)
	if err != nil {
		return nil, err
	}
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	return resp, nil
}

// ChatStream returns a StreamReader that yields chunks as they arrive from the LLM.
func (p *OpenAIProvider) ChatStream(ctx context.Context, messages []Message, tools []Tool, model string, maxTokens int, temperature float64) (*StreamReader, error) {
	resp, err := p.openStream(ctx, messages, tools, model, maxTokens, temperature, false)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		body := string(respBody)
		// Same self-healing as the non-streaming path.
		if resp.StatusCode == http.StatusBadRequest && isTemperatureLockError(body) {
			slog.Warn("upstream rejected temperature on stream, retrying without it",
				"model", StripProviderPrefix(model),
			)
			resp2, err2 := p.openStream(ctx, messages, tools, model, maxTokens, temperature, true)
			if err2 != nil {
				return nil, err2
			}
			if resp2.StatusCode != http.StatusOK {
				body2, _ := io.ReadAll(resp2.Body)
				resp2.Body.Close()
				return nil, fmt.Errorf("API error %d: %s", resp2.StatusCode, string(body2))
			}
			resp = resp2
		} else {
			return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, body)
		}
	}

	ch := make(chan StreamChunk, 64)
	reader := NewStreamReader(ch)

	go func() {
		defer resp.Body.Close()
		defer close(ch)

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

		toolCalls := make(map[int]*ToolCall)

		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				// Send final chunk with accumulated tool calls
				var tcs []ToolCall
				for i := 0; i < len(toolCalls); i++ {
					if tc, ok := toolCalls[i]; ok {
						tcs = append(tcs, *tc)
					}
				}
				select {
				case ch <- StreamChunk{ToolCalls: tcs, Done: true}:
				case <-ctx.Done():
				}
				return
			}

			var chunk sseResponse
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				slog.Warn("parse SSE chunk", "error", err, "data", data)
				continue
			}

			if len(chunk.Choices) == 0 {
				continue
			}

			delta := chunk.Choices[0].Delta

			// Accumulate tool calls
			for _, tc := range delta.ToolCalls {
				existing, ok := toolCalls[tc.Index]
				if !ok {
					toolCalls[tc.Index] = &ToolCall{
						ID:   tc.ID,
						Type: tc.Type,
						Function: FunctionCall{
							Name:      tc.Function.Name,
							Arguments: tc.Function.Arguments,
						},
					}
				} else {
					if tc.ID != "" {
						existing.ID = tc.ID
					}
					if tc.Type != "" {
						existing.Type = tc.Type
					}
					if tc.Function.Name != "" {
						existing.Function.Name += tc.Function.Name
					}
					existing.Function.Arguments += tc.Function.Arguments
				}
			}

			// Yield content chunks
			if delta.Content != "" {
				select {
				case ch <- StreamChunk{Content: delta.Content}:
				case <-ctx.Done():
					return
				}
			}
		}

		if err := scanner.Err(); err != nil {
			reader.SetErr(fmt.Errorf("read stream: %w", err))
		}
	}()

	return reader, nil
}

func (p *OpenAIProvider) parseSSE(reader io.Reader) (*Response, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var contentBuilder strings.Builder
	var reasoningBuilder strings.Builder
	toolCalls := make(map[int]*ToolCall)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk sseResponse
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			slog.Warn("parse SSE chunk", "error", err, "data", data)
			continue
		}

		if len(chunk.Choices) == 0 {
			continue
		}

		delta := chunk.Choices[0].Delta

		if delta.Content != "" {
			contentBuilder.WriteString(delta.Content)
		}
		// Accumulate DeepSeek reasoning_content so it can be echoed back.
		if delta.ReasoningContent != "" {
			reasoningBuilder.WriteString(delta.ReasoningContent)
		}

		for _, tc := range delta.ToolCalls {
			existing, ok := toolCalls[tc.Index]
			if !ok {
				toolCalls[tc.Index] = &ToolCall{
					ID:   tc.ID,
					Type: tc.Type,
					Function: FunctionCall{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				}
			} else {
				if tc.ID != "" {
					existing.ID = tc.ID
				}
				if tc.Type != "" {
					existing.Type = tc.Type
				}
				if tc.Function.Name != "" {
					existing.Function.Name += tc.Function.Name
				}
				existing.Function.Arguments += tc.Function.Arguments
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read stream: %w", err)
	}

	result := &Response{
		Content:          contentBuilder.String(),
		ReasoningContent: reasoningBuilder.String(),
	}
	for i := 0; i < len(toolCalls); i++ {
		if tc, ok := toolCalls[i]; ok {
			result.ToolCalls = append(result.ToolCalls, *tc)
		}
	}

	return result, nil
}
