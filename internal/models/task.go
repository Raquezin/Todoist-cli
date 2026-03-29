package models

type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ProjectsResponse struct {
	Results    []Project `json:"results"`
	NextCursor string    `json:"next_cursor,omitempty"`
}

type TaskRequest struct {
	Content      string   `json:"content"`
	DueString    string   `json:"due_string,omitempty"`
	DueDatetime  string   `json:"due_datetime,omitempty"`
	DueDate      string   `json:"due_date,omitempty"`
	ProjectID    string   `json:"project_id,omitempty"`
	Labels       []string `json:"labels,omitempty"`
	Description  string   `json:"description,omitempty"`
	Priority     int      `json:"priority,omitempty"`
	Duration     int      `json:"duration,omitempty"`
	DurationUnit string   `json:"duration_unit,omitempty"`
}

type TaskResponse struct {
	Content     string    `json:"content"`
	Description string    `json:"description,omitempty"`
	ProjectID   string    `json:"project_id,omitempty"`
	Labels      []string  `json:"labels,omitempty"`
	Priority    int       `json:"priority,omitempty"`
	DueDatetime string    `json:"due_datetime,omitempty"`
	Duration    *Duration `json:"duration,omitempty"`
}

type Due struct {
	Date     string `json:"date"`
	Datetime string `json:"datetime,omitempty"`
	String   string `json:"string,omitempty"`
	Timezone string `json:"timezone,omitempty"`
}

type Duration struct {
	Amount int    `json:"amount"`
	Unit   string `json:"unit"`
}

type FilteredTask struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Content   string    `json:"content"`
	Priority  int       `json:"priority"`
	Labels    []string  `json:"labels"`
	Due       *Due      `json:"due,omitempty"`
	Duration  *Duration `json:"duration,omitempty"`
}

type FilterResponse struct {
	Results    []FilteredTask `json:"results"`
	NextCursor string         `json:"next_cursor,omitempty"`
}

// ToAPIPriority converts UI priority (1=Urgent, 4=Normal) to Todoist API priority (4=Urgent, 1=Normal).
func ToAPIPriority(ui int) int {
	p := 5 - ui
	if p < 1 {
		return 1
	}
	if p > 4 {
		return 4
	}
	return p
}

// ToUIPriority converts Todoist API priority (4=Urgent, 1=Normal) to UI priority (1=Urgent, 4=Normal).
func ToUIPriority(api int) int {
	p := 5 - api
	if p < 1 {
		return 1
	}
	if p > 4 {
		return 4
	}
	return p
}
