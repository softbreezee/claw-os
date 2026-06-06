package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type webSearchArgs struct {
	Query string `json:"query"`
	Count int    `json:"count,omitempty"`
}

// braveSearchResponse is the Brave Search API response shape.
type braveSearchResponse struct {
	Web struct {
		Results []struct {
			Title       string `json:"title"`
			URL         string `json:"url"`
			Description string `json:"description"`
		} `json:"results"`
	} `json:"web"`
}

// bochaSearchResponse is the Bocha AI Search API response shape.
type bochaSearchResponse struct {
	WebPages struct {
		TotalEstimatedMatches int `json:"totalEstimatedMatches"`
		Value                 []struct {
			Name          string `json:"name"`
			URL           string `json:"url"`
			SiteName      string `json:"siteName"`
			Snippet       string `json:"snippet"`
			Summary       string `json:"summary"`
			DatePublished string `json:"datePublished"`
		} `json:"value"`
	} `json:"webPages"`
}

const webSearchDesc = "Search the web and return results with titles, URLs, and snippets. Supports Chinese search queries."

// RegisterWebSearch registers the web_search tool.
// provider is either "brave" or "bocha"; apiKey is the respective API key.
func RegisterWebSearch(r *Registry, provider, apiKey string) {
	r.Register("web_search", webSearchDesc, map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "The search query (supports Chinese)",
			},
			"count": map[string]interface{}{
				"type":        "integer",
				"description": "Number of results to return (default 5, max 20)",
			},
		},
		"required": []string{"query"},
	}, makeWebSearchTool(provider, apiKey))
}

func makeWebSearchTool(provider, apiKey string) ToolFunc {
	switch provider {
	case "bocha":
		return makeBochaSearch(apiKey)
	default:
		return makeBraveSearch(apiKey)
	}
}

// ─── Brave Search ───────────────────────────────────────────────────────────

func makeBraveSearch(apiKey string) ToolFunc {
	return func(ctx context.Context, rawArgs json.RawMessage) (string, error) {
		var args webSearchArgs
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return "", fmt.Errorf("parse args: %w", err)
		}

		if args.Query == "" {
			return "", fmt.Errorf("query is required")
		}

		count := args.Count
		if count <= 0 {
			count = 5
		}
		if count > 20 {
			count = 20
		}

		searchCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(searchCtx, http.MethodGet, "https://api.search.brave.com/res/v1/web/search", nil)
		if err != nil {
			return "", fmt.Errorf("create request: %w", err)
		}

		q := req.URL.Query()
		q.Set("q", args.Query)
		q.Set("count", fmt.Sprintf("%d", count))
		req.URL.RawQuery = q.Encode()

		req.Header.Set("Accept", "application/json")
		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Set("X-Subscription-Token", apiKey)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", fmt.Errorf("search request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
			return "", fmt.Errorf("Brave Search API returned HTTP %d: %s", resp.StatusCode, string(body))
		}

		var searchResp braveSearchResponse
		if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
			return "", fmt.Errorf("parse search response: %w", err)
		}

		if len(searchResp.Web.Results) == 0 {
			return "No results found for: " + args.Query, nil
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Search results for: %s\n\n", args.Query))
		for i, r := range searchResp.Web.Results {
			sb.WriteString(fmt.Sprintf("%d. %s\n   URL: %s\n   %s\n\n", i+1, r.Title, r.URL, r.Description))
		}

		return sb.String(), nil
	}
}

// ─── Bocha AI Search ────────────────────────────────────────────────────────

func makeBochaSearch(apiKey string) ToolFunc {
	return func(ctx context.Context, rawArgs json.RawMessage) (string, error) {
		var args webSearchArgs
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return "", fmt.Errorf("parse args: %w", err)
		}

		if args.Query == "" {
			return "", fmt.Errorf("query is required")
		}

		count := args.Count
		if count <= 0 {
			count = 8
		}
		if count > 20 {
			count = 20
		}

		body := map[string]interface{}{
			"query":     args.Query,
			"freshness": "noLimit",
			"summary":   true,
			"count":     count,
		}
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return "", fmt.Errorf("marshal request: %w", err)
		}

		searchCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(searchCtx, http.MethodPost, "https://api.bochaai.com/v1/web-search", bytes.NewReader(bodyBytes))
		if err != nil {
			return "", fmt.Errorf("create request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", fmt.Errorf("search request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
			return "", fmt.Errorf("Bocha Search API returned HTTP %d: %s", resp.StatusCode, string(respBody))
		}

		var searchResp bochaSearchResponse
		if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
			return "", fmt.Errorf("parse search response: %w", err)
		}

		if len(searchResp.WebPages.Value) == 0 {
			return "No results found for: " + args.Query, nil
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Search results for: %s (%d total matches)\n\n", args.Query, searchResp.WebPages.TotalEstimatedMatches))
		for i, r := range searchResp.WebPages.Value {
			desc := r.Summary
			if desc == "" {
				desc = r.Snippet
			}
			dateInfo := ""
			if r.DatePublished != "" {
				dateInfo = fmt.Sprintf(" (%s)", r.DatePublished)
			}
			sb.WriteString(fmt.Sprintf("%d. %s%s\n   URL: %s\n   Source: %s\n   %s\n\n", i+1, r.Name, dateInfo, r.URL, r.SiteName, desc))
		}

		return sb.String(), nil
	}
}
