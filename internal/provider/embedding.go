package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

// DefaultEmbedModel is the model used when the caller passes "".
// text-embedding-3-small is a good cost/quality default — 1536 dims,
// fast, and matches the pgvector schema deployed by Pawnix (memories
// table uses vector(1536)). If you switch to a different-dim model
// you'll also need to ALTER the embedding column accordingly.
const DefaultEmbedModel = "text-embedding-3-small"

// embeddingRequest is the OpenAI /v1/embeddings request body.
type embeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

// embeddingResponse mirrors the OpenAI /v1/embeddings response.
// We only need the first vector — Embed accepts a single string and
// returns its corresponding vector.
type embeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

// Embed converts text to a vector using the OpenAI-compatible
// /embeddings endpoint. The model defaults to DefaultEmbedModel when
// blank. Returns the raw float32 slice ready to be passed to
// pg.MemoryStore.Insert / SearchSemantic.
//
// Errors are returned verbatim — the memory layer is responsible for
// degrading to nil-embedding writes when this fails (so an embedding
// outage doesn't block memory persistence).
func (p *OpenAIProvider) Embed(ctx context.Context, text string, model string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("embed: empty input")
	}
	if model == "" {
		model = DefaultEmbedModel
	}
	model = StripProviderPrefix(model)

	body, err := json.Marshal(embeddingRequest{Model: model, Input: text})
	if err != nil {
		return nil, fmt.Errorf("embed: marshal: %w", err)
	}

	url := p.apiBase + p.embedPath
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("embed: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("embed: http: %w", err)
	}
	defer resp.Body.Close()

	rawBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embed: API %d: %s", resp.StatusCode, string(rawBody))
	}

	var er embeddingResponse
	if err := json.Unmarshal(rawBody, &er); err != nil {
		return nil, fmt.Errorf("embed: decode: %w", err)
	}
	if er.Error != nil {
		return nil, fmt.Errorf("embed: API error: %s", er.Error.Message)
	}
	if len(er.Data) == 0 || len(er.Data[0].Embedding) == 0 {
		return nil, fmt.Errorf("embed: empty response")
	}

	slog.Debug("embed: ok",
		"model", model,
		"dims", len(er.Data[0].Embedding),
		"input_len", len(text),
	)
	return er.Data[0].Embedding, nil
}

// Embed is not supported by the Anthropic Messages API. Returns
// ErrEmbeddingNotSupported so callers can route around it (typically
// by skipping embedding altogether and storing memories with
// embedding=nil).
func (p *AnthropicProvider) Embed(ctx context.Context, text string, model string) ([]float32, error) {
	return nil, ErrEmbeddingNotSupported
}
