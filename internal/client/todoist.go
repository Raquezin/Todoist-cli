package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"todoist-cli/internal/models"
)

var BaseURL = "https://api.todoist.com/api/v1"

type TodoistClient struct {
	Token  string
	Client *http.Client
}

func New(token string) *TodoistClient {
	return &TodoistClient{
		Token:  token,
		Client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *TodoistClient) doRequest(method, endpoint string, body []byte) (*http.Response, error) {
	var req *http.Request
	var err error
	if body != nil {
		req, err = http.NewRequest(method, BaseURL+endpoint, bytes.NewBuffer(body))
	} else {
		req, err = http.NewRequest(method, BaseURL+endpoint, nil)
	}

	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.Client.Do(req)
}

func (c *TodoistClient) GetProjects() ([]models.Project, error) {
	resp, err := c.doRequest("GET", "/projects", nil)
	if err != nil {
		return nil, fmt.Errorf("error connecting to Todoist: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	var projectsResp models.ProjectsResponse
	if err := json.NewDecoder(resp.Body).Decode(&projectsResp); err != nil {
		return nil, fmt.Errorf("error decoding projects: %w", err)
	}

	return projectsResp.Results, nil
}

func (c *TodoistClient) CreateTask(task models.TaskRequest) (*models.TaskResponse, error) {
	payload, err := json.Marshal(task)
	if err != nil {
		return nil, fmt.Errorf("error marshaling task: %w", err)
	}

	resp, err := c.doRequest("POST", "/tasks", payload)
	if err != nil {
		return nil, fmt.Errorf("error creating task: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	var taskRes models.TaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&taskRes); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
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
	resp, err := c.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	var apiResp models.FilterResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("error decoding JSON: %w", err)
	}

	return &apiResp, nil
}
