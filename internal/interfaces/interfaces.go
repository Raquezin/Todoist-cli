package interfaces

import (
	"context"

	"todoist-cli/internal/models"
)

type TaskCreator interface {
	CreateTask(task models.TaskRequest) (*models.TaskResponse, error)
}

type TaskFilterer interface {
	FilterTasks(ctx context.Context, query, cursor string) (*models.FilterResponse, error)
}

type ProjectProvider interface {
	GetProjects() ([]models.Project, error)
}

type SectionProvider interface {
	GetSections() ([]models.Section, error)
}
