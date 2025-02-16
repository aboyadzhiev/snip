package urlhaus

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type MaliciousURL struct {
	DateAdded   string   `json:"dateadded"`
	URL         string   `json:"url"`
	URLStatus   string   `json:"url_status"`
	LastOnline  string   `json:"last_online"`
	Threat      string   `json:"threat"`
	Tags        []string `json:"tags"`
	UrlhausLink string   `json:"urlhaus_link"`
	Reporter    string   `json:"reporter"`
}

type Client interface {
	FetchAll(ctx context.Context) ([]MaliciousURL, error)
}

type urlhausClient struct {
	apiEndpoint string
	httpClient  *http.Client
	logger      *slog.Logger
}

func (c *urlhausClient) FetchAll(ctx context.Context) ([]MaliciousURL, error) {
	req, err := http.NewRequest(http.MethodGet, c.apiEndpoint, nil)
	if err != nil {
		c.logger.ErrorContext(ctx, fmt.Sprintf("Error creating request: %v", err))
		return nil, err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.ErrorContext(ctx, fmt.Sprintf("Error executing request: %v", err))
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		c.logger.ErrorContext(ctx, fmt.Sprintf("Error fetching users: %v", res.StatusCode))
		return nil, fmt.Errorf("error fetching users: %v", res.StatusCode)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		c.logger.ErrorContext(ctx, fmt.Sprintf("Error reading body: %v", err))
		return nil, err
	}

	var rawURLs map[string][]MaliciousURL
	err = json.Unmarshal(body, &rawURLs)
	if err != nil {
		c.logger.ErrorContext(ctx, fmt.Sprintf("Error unmarshalling body: %v", err))
		return nil, err
	}

	urls := make([]MaliciousURL, 0, len(rawURLs))
	for _, u := range rawURLs {
		urls = append(urls, u[0])
	}

	c.logger.InfoContext(ctx, fmt.Sprintf("Fetched %d malicious urls", len(urls)))

	return urls, nil
}

func NewClient(apiEndpoint string, httpClient *http.Client, logger *slog.Logger) Client {
	return &urlhausClient{
		apiEndpoint: apiEndpoint,
		httpClient:  httpClient,
		logger:      logger,
	}
}
