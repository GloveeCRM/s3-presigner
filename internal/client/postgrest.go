package client

import (
	"fmt"
	"net/http"
	"net/url"
)

type PostgrestClient struct {
	baseURL string
	apiKey  string
	schema  string
	client  *http.Client
}

func NewPostgrestClient(baseURL, apiKey, schema string) *PostgrestClient {
	return &PostgrestClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		schema:  schema,
		client:  http.DefaultClient,
	}
}

func (pc *PostgrestClient) Request(method, path string, params url.Values) (*http.Response, error) {
	requestURL := fmt.Sprintf("%s%s", pc.baseURL, path)
	if params != nil {
		requestURL = fmt.Sprintf("%s?%s", requestURL, params.Encode())
	}

	req, err := http.NewRequest(method, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Accept-Profile", pc.schema)
	req.Header.Set("Content-Profile", pc.schema)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", pc.apiKey))

	return pc.client.Do(req)
}
