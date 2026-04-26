package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"todoist-cli/internal/models"
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

// doRequest performs the HTTP request, handles errors, and unmarshals the response into target.
func (c *TodoistClient) doRequest(method, endpoint string, reqBody any, target any) error {
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
		var bodyReader io.Reader
		if reqBytes != nil {
			bodyReader = bytes.NewBuffer(reqBytes)
		}

		req, err := http.NewRequest(method, fullURL, bodyReader)
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
				resp.Body.Close()
				secs := 1
				if retryAfter != "" {
					if parsedSecs, err := strconv.Atoi(retryAfter); err == nil {
						secs = parsedSecs
					}
				} else if resp.StatusCode >= 500 {
					secs = 2
				}
				time.Sleep(time.Duration(secs) * time.Second)
				continue
			}
		}
		break
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (%d): %s", resp.StatusCode, string(bodyBytes))
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
	if err := c.doRequest("GET", "/projects", nil, &projectsResp); err != nil {
		return nil, err
	}
	return projectsResp.Results, nil
}

func (c *TodoistClient) CreateTask(task models.TaskRequest) (*models.TaskResponse, error) {
	var taskRes models.TaskResponse
	if err := c.doRequest("POST", "/tasks", task, &taskRes); err != nil {
		return nil, err
	}
	return &taskRes, nil
}

func (c *TodoistClient) FilterTasks(queryFinal, cursor string) (*models.FilterResponse, error) {
	params := url.Values{}
	params.Add("query", queryFinal)
	if cursor != "" {
		params.Add("cursor", cursor)
	}

	endpoint := "/tasks/filter?" + params.Encode()
	var apiResp models.FilterResponse
	if err := c.doRequest("GET", endpoint, nil, &apiResp); err != nil {
		return nil, err
	}
	return &apiResp, nil
}
