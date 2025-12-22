package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Result represents a single search result
type Result struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"`
}

// SearchResults contains the search results and metadata
type SearchResults struct {
	Query   string   `json:"query"`
	Results []Result `json:"results"`
}

// Client defines the interface for web search providers
type Client interface {
	Search(ctx context.Context, query string, maxResults int) (*SearchResults, error)
}

// Config holds the configuration for search clients
type Config struct {
	Provider   string // "tavily", "google", or "duckduckgo"
	APIKey     string
	MaxResults int
	Timeout    time.Duration
}

// NewClient creates a new search client based on the configuration
func NewClient(cfg Config) (Client, error) {
	if cfg.MaxResults == 0 {
		cfg.MaxResults = 5
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	switch cfg.Provider {
	case "tavily":
		if cfg.APIKey == "" {
			return nil, fmt.Errorf("TAVILY_API_KEY is required for tavily provider")
		}
		return &tavilyClient{
			apiKey:     cfg.APIKey,
			maxResults: cfg.MaxResults,
			httpClient: &http.Client{Timeout: cfg.Timeout},
		}, nil
	case "google":
		if cfg.APIKey == "" {
			return nil, fmt.Errorf("GOOGLE_SEARCH_API_KEY is required for google provider")
		}
		return &googleClient{
			apiKey:     cfg.APIKey,
			maxResults: cfg.MaxResults,
			httpClient: &http.Client{Timeout: cfg.Timeout},
		}, nil
	case "duckduckgo":
		return &duckDuckGoClient{
			maxResults: cfg.MaxResults,
			httpClient: &http.Client{Timeout: cfg.Timeout},
		}, nil
	case "":
		// No search provider configured, return nil client
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupported search provider: %s", cfg.Provider)
	}
}

// tavilyClient implements Client for Tavily Search API
type tavilyClient struct {
	apiKey     string
	maxResults int
	httpClient *http.Client
}

type tavilyRequest struct {
	APIKey           string `json:"api_key"`
	Query            string `json:"query"`
	MaxResults       int    `json:"max_results"`
	SearchDepth      string `json:"search_depth"`
	IncludeAnswer    bool   `json:"include_answer"`
	IncludeRawContent bool  `json:"include_raw_content"`
}

type tavilyResponse struct {
	Results []struct {
		Title   string `json:"title"`
		URL     string `json:"url"`
		Content string `json:"content"`
	} `json:"results"`
	Answer string `json:"answer,omitempty"`
}

func (c *tavilyClient) Search(ctx context.Context, query string, maxResults int) (*SearchResults, error) {
	if maxResults == 0 {
		maxResults = c.maxResults
	}

	reqBody := tavilyRequest{
		APIKey:           c.apiKey,
		Query:            query,
		MaxResults:       maxResults,
		SearchDepth:      "advanced",
		IncludeAnswer:    false,
		IncludeRawContent: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.tavily.com/search", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Tavily API error (status %d): %s", resp.StatusCode, string(body))
	}

	var tavilyResp tavilyResponse
	if err := json.Unmarshal(body, &tavilyResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	results := &SearchResults{
		Query:   query,
		Results: make([]Result, 0, len(tavilyResp.Results)),
	}

	for _, r := range tavilyResp.Results {
		results.Results = append(results.Results, Result{
			Title:   r.Title,
			URL:     r.URL,
			Content: r.Content,
		})
	}

	return results, nil
}

// googleClient implements Client for Google Custom Search API
type googleClient struct {
	apiKey     string
	cseID      string // Custom Search Engine ID
	maxResults int
	httpClient *http.Client
}

type googleResponse struct {
	Items []struct {
		Title   string `json:"title"`
		Link    string `json:"link"`
		Snippet string `json:"snippet"`
	} `json:"items"`
}

func (c *googleClient) Search(ctx context.Context, query string, maxResults int) (*SearchResults, error) {
	if maxResults == 0 {
		maxResults = c.maxResults
	}

	// Build URL with query parameters
	baseURL := "https://www.googleapis.com/customsearch/v1"
	params := url.Values{}
	params.Set("key", c.apiKey)
	params.Set("q", query)
	params.Set("num", fmt.Sprintf("%d", maxResults))
	if c.cseID != "" {
		params.Set("cx", c.cseID)
	}

	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Google API error (status %d): %s", resp.StatusCode, string(body))
	}

	var googleResp googleResponse
	if err := json.Unmarshal(body, &googleResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	results := &SearchResults{
		Query:   query,
		Results: make([]Result, 0, len(googleResp.Items)),
	}

	for _, item := range googleResp.Items {
		results.Results = append(results.Results, Result{
			Title:   item.Title,
			URL:     item.Link,
			Content: item.Snippet,
		})
	}

	return results, nil
}

// duckDuckGoClient implements Client for DuckDuckGo Instant Answer API
type duckDuckGoClient struct {
	maxResults int
	httpClient *http.Client
}

type duckDuckGoResponse struct {
	AbstractText   string `json:"AbstractText"`
	AbstractSource string `json:"AbstractSource"`
	AbstractURL    string `json:"AbstractURL"`
	RelatedTopics  []struct {
		Text     string `json:"Text"`
		FirstURL string `json:"FirstURL"`
	} `json:"RelatedTopics"`
}

func (c *duckDuckGoClient) Search(ctx context.Context, query string, maxResults int) (*SearchResults, error) {
	if maxResults == 0 {
		maxResults = c.maxResults
	}

	// DuckDuckGo Instant Answer API
	baseURL := "https://api.duckduckgo.com/"
	params := url.Values{}
	params.Set("q", query)
	params.Set("format", "json")
	params.Set("no_html", "1")
	params.Set("skip_disambig", "1")

	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "VerdictAgent/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("DuckDuckGo API error (status %d): %s", resp.StatusCode, string(body))
	}

	var ddgResp duckDuckGoResponse
	if err := json.Unmarshal(body, &ddgResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	results := &SearchResults{
		Query:   query,
		Results: make([]Result, 0, maxResults),
	}

	// Add abstract if available
	if ddgResp.AbstractText != "" {
		results.Results = append(results.Results, Result{
			Title:   ddgResp.AbstractSource,
			URL:     ddgResp.AbstractURL,
			Content: ddgResp.AbstractText,
		})
	}

	// Add related topics
	for _, topic := range ddgResp.RelatedTopics {
		if len(results.Results) >= maxResults {
			break
		}
		if topic.Text != "" && topic.FirstURL != "" {
			results.Results = append(results.Results, Result{
				Title:   extractTitle(topic.Text),
				URL:     topic.FirstURL,
				Content: topic.Text,
			})
		}
	}

	return results, nil
}

// extractTitle extracts a title from DuckDuckGo topic text
func extractTitle(text string) string {
	// DuckDuckGo format: "Title - Description"
	parts := strings.SplitN(text, " - ", 2)
	if len(parts) > 0 {
		return strings.TrimSpace(parts[0])
	}
	return text
}

// FormatForPrompt formats search results for inclusion in an LLM prompt
func (sr *SearchResults) FormatForPrompt() string {
	if sr == nil || len(sr.Results) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Web Search Results for: %s\n\n", sr.Query))

	for i, r := range sr.Results {
		sb.WriteString(fmt.Sprintf("### [%d] %s\n", i+1, r.Title))
		sb.WriteString(fmt.Sprintf("URL: %s\n", r.URL))
		sb.WriteString(fmt.Sprintf("Content: %s\n\n", r.Content))
	}

	sb.WriteString("---\n")
	sb.WriteString("Use the above search results to provide accurate, up-to-date information in your response.\n")

	return sb.String()
}
