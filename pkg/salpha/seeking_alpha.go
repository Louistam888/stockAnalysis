package salpha

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/louistam888/stockAnalysis/internal/news"
)

const (
	urlPath      = "/news/v2/list-by-symbol"
	apiKeyHeader = "x-rapidapi-key"
	pageSize     = 5
)

type client struct {
	baseURL string
	apiKey  string
}

// Fetch retrieves articles for a given ticker
func (c *client) Fetch(ticker string) ([]news.Article, error) {
	fullURL, err := c.buildURL(ticker)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add(apiKeyHeader, c.apiKey)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("unsuccessful status code: %d", resp.StatusCode)
	}

	return c.parse(resp)
}

// parse converts the HTTP response into news.Article slices
func (c *client) parse(resp *http.Response) ([]news.Article, error) {
	res := &SeekingAlphaResponse{}
	if err := json.NewDecoder(resp.Body).Decode(res); err != nil {
		return nil, err
	}

	var articles []news.Article
	for _, item := range res.Data {
		// Convert PublishOn string to time.Time
		publishTime, err := time.Parse(time.RFC3339, item.Attributes.PublishOn)
		if err != nil {
			// skip invalid dates
			continue
		}

		art := news.Article{
			PublishOn: publishTime,
			Headline:  item.Attributes.Title,
		}
		articles = append(articles, art)
	}
	return articles, nil
}

// buildURL constructs the full API URL for a ticker
func (c *client) buildURL(ticker string) (string, error) {
	parsedURL, err := url.Parse(c.baseURL)
	if err != nil {
		return "", err
	}

	parsedURL.Path += urlPath

	params := url.Values{}
	params.Add("size", fmt.Sprint(pageSize))
	params.Add("id", ticker)
	parsedURL.RawQuery = params.Encode()

	return parsedURL.String(), nil
}

// NewClient returns a news.Fetcher implementation
func NewClient(baseURL, apiKey string) news.Fetcher {
	return &client{baseURL: baseURL, apiKey: apiKey}
}

// ---- Helper types for JSON parsing ----
type attributes struct {
	PublishOn string `json:"publishOn"`
	Title     string `json:"title"`
}

type seekingAlphaNews struct {
	Attributes attributes `json:"attributes"`
}

type SeekingAlphaResponse struct {
	Data []seekingAlphaNews `json:"data"`
}
