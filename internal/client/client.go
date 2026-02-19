package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultTimeout    = 30 * time.Second
	maxRetryAttempts  = 3
	retryBackoffBase  = 200 * time.Millisecond
	contentTypeJSON   = "application/json"
	authorizationHead = "Authorization"
)

type Client struct {
	baseURL    *url.URL
	apiKey     string
	httpClient *http.Client
}

type apiErrorEnvelope struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Details string `json:"details"`
}

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("lago api error (status %d): %s", e.StatusCode, e.Message)
}

func NewClient(apiEndpoint, apiKey string) (*Client, error) {
	if strings.TrimSpace(apiEndpoint) == "" {
		return nil, errors.New("api endpoint cannot be empty")
	}
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("api key cannot be empty")
	}

	cleanEndpoint := strings.TrimSpace(apiEndpoint)
	cleanAPIKey := strings.TrimSpace(apiKey)

	parsed, err := url.Parse(cleanEndpoint)
	if err != nil {
		return nil, fmt.Errorf("parse api endpoint: %w", err)
	}
	if !parsed.IsAbs() || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return nil, errors.New("api endpoint must be an absolute http(s) URL")
	}

	return &Client{
		baseURL: parsed,
		apiKey:  cleanAPIKey,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}, nil
}

func IsNotFound(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == http.StatusNotFound
	}
	return false
}

func (c *Client) doRequest(ctx context.Context, method, path string, requestBody any, responseBody any) error {
	endpoint, err := c.baseURL.Parse(path)
	if err != nil {
		return fmt.Errorf("build request url: %w", err)
	}

	var bodyBytes []byte
	if requestBody != nil {
		bodyBytes, err = json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
	}

	for attempt := 1; attempt <= maxRetryAttempts; attempt++ {
		var bodyReader io.Reader
		if bodyBytes != nil {
			bodyReader = bytes.NewReader(bodyBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), bodyReader)
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}

		req.Header.Set(authorizationHead, fmt.Sprintf("Bearer %s", c.apiKey))
		if requestBody != nil {
			req.Header.Set("Content-Type", contentTypeJSON)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			if attempt == maxRetryAttempts {
				return fmt.Errorf("perform request: %w", err)
			}
			time.Sleep(retryBackoffBase * time.Duration(attempt))
			continue
		}

		respBytes, readErr := io.ReadAll(resp.Body)
		closeErr := resp.Body.Close()
		if readErr != nil {
			return fmt.Errorf("read response body: %w", readErr)
		}
		if closeErr != nil {
			return fmt.Errorf("close response body: %w", closeErr)
		}

		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			if (resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500) && attempt < maxRetryAttempts {
				time.Sleep(retryBackoffBase * time.Duration(attempt))
				continue
			}

			apiErr := &APIError{StatusCode: resp.StatusCode}
			if len(respBytes) > 0 {
				var envelope apiErrorEnvelope
				if err := json.Unmarshal(respBytes, &envelope); err == nil {
					switch {
					case envelope.Error != "":
						apiErr.Message = envelope.Error
					case envelope.Message != "":
						apiErr.Message = envelope.Message
					case envelope.Details != "":
						apiErr.Message = envelope.Details
					}
				}
				if apiErr.Message == "" {
					apiErr.Message = string(respBytes)
				}
			}
			if apiErr.Message == "" {
				apiErr.Message = "request failed"
			}
			return apiErr
		}

		if responseBody != nil && len(respBytes) > 0 {
			if err := json.Unmarshal(respBytes, responseBody); err != nil {
				return fmt.Errorf("decode response body: %w", err)
			}
		}

		return nil
	}

	return errors.New("request failed after retries")
}
