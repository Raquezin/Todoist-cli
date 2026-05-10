package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"todoist-cli/internal/limits"
	"todoist-cli/internal/models"
	"todoist-cli/internal/sanitize"
)

type TodoistClient struct {
	Token   string
	BaseURL string
	Client  *http.Client
}

func New(token string) *TodoistClient {
	baseURL := os.Getenv("TODOIST_API_URL")
	if baseURL == "" {
		baseURL = "https://api.todoist.com/api/v1"
	}
	return &TodoistClient{
		Token:   token,
		BaseURL: baseURL,
		Client:  &http.Client{Timeout: 10 * time.Second},
	}
}

const maxRetryAfter = 5 * time.Second

func isLoopback(host string) bool {
	if host == "localhost" {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback()
	}
	return false
}

func validateBaseURL(baseURL string) error {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("invalid API URL: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return fmt.Errorf("invalid API URL scheme %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return fmt.Errorf("invalid API URL: missing host")
	}
	if parsed.Scheme == "http" {
		host := parsed.Hostname()
		if !isLoopback(host) {
			return fmt.Errorf("refusing to send API token over insecure HTTP to %s", host)
		}
	}
	return nil
}

// doRequest performs the HTTP request, handles errors, and unmarshals the response into target.
func (c *TodoistClient) doRequest(ctx context.Context, method, endpoint string, reqBody any, target any) error {
	if err := validateBaseURL(c.BaseURL); err != nil {
		return err
	}

	var reqBytes []byte
	var err error
	if reqBody != nil {
		reqBytes, err = json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("error marshaling request: %w", err)
		}
	}

	endpoint = strings.TrimLeft(endpoint, "/")
	fullURL := strings.TrimRight(c.BaseURL, "/") + "/" + endpoint

	var resp *http.Response
	for retries := 0; retries < 3; retries++ {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("request cancelled: %w", err)
		}
		var bodyReader io.Reader
		if reqBytes != nil {
			bodyReader = bytes.NewBuffer(reqBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
		if err != nil {
			return err
		}

		req.Header.Set("Authorization", "Bearer "+c.Token)
		if reqBytes != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err = c.Client.Do(req)
		if err != nil {
			return fmt.Errorf("network error: %w", err)
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			if retries < 2 {
				retryAfter := resp.Header.Get("Retry-After")
				_ = resp.Body.Close()
				secs := 1
				if retryAfter != "" {
					if parsedSecs, err := strconv.Atoi(retryAfter); err == nil {
						secs = parsedSecs
					}
				} else if resp.StatusCode >= 500 {
					secs = 2
				}
				delay := time.Duration(secs) * time.Second
				if delay < 0 {
					delay = 0
				}
				if delay > maxRetryAfter {
					delay = maxRetryAfter
				}
				time.Sleep(delay)
				continue
			}
		}
		break
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, limits.MaxErrorBodyBytes+1))
		body := string(bodyBytes)
		if len(bodyBytes) > limits.MaxErrorBodyBytes {
			body = sanitize.TerminalLimit(body, limits.MaxErrorBodyBytes)
		} else {
			body = sanitize.Terminal(body)
		}
		return fmt.Errorf("API error (%d): %s", resp.StatusCode, body)
	}

	if target != nil {
		if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
			return fmt.Errorf("error decoding response: %w", err)
		}
	}
	return nil
}

func (c *TodoistClient) GetProjects() ([]models.Project, error) {
	var projectsResp models.ProjectsResponse
	if err := c.doRequest(context.Background(), "GET", "/projects", nil, &projectsResp); err != nil {
		return nil, err
	}
	return projectsResp.Results, nil
}

func (c *TodoistClient) GetSections() ([]models.Section, error) {
	var sectionsResp models.SectionsResponse
	if err := c.doRequest(context.Background(), "GET", "/sections", nil, &sectionsResp); err != nil {
		return nil, err
	}
	return sectionsResp.Results, nil
}

func (c *TodoistClient) CreateTask(task models.TaskRequest) (*models.TaskResponse, error) {
	var taskRes models.TaskResponse
	if err := c.doRequest(context.Background(), "POST", "/tasks", task, &taskRes); err != nil {
		return nil, err
	}
	return &taskRes, nil
}

func (c *TodoistClient) FilterTasks(ctx context.Context, queryFinal, cursor string) (*models.FilterResponse, error) {
	params := url.Values{}
	params.Add("query", queryFinal)
	if cursor != "" {
		params.Add("cursor", cursor)
	}

	endpoint := "/tasks/filter?" + params.Encode()
	var apiResp models.FilterResponse
	if err := c.doRequest(ctx, "GET", endpoint, nil, &apiResp); err != nil {
		return nil, err
	}
	return &apiResp, nil
}
