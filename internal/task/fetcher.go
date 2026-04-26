package task

import (
	"fmt"
	"time"

	"todoist-cli/internal/cache"
	"todoist-cli/internal/client"
	"todoist-cli/internal/models"
)

// --- CONFIGURACIÓN DE QUERIES ---

const exclusionGlobal = "& !(#Study & /Horario)"
const maxPages = 20

var queries = map[string]string{
	"foco":  "today & p1 & !@reuniones",
	"radar": "7 days & @importante",
}

type Fetcher struct {
	Client *client.TodoistClient
}

func NewFetcher(apiClient *client.TodoistClient) *Fetcher {
	return &Fetcher{Client: apiClient}
}

func (f *Fetcher) Fetch(queryName string) error {
	queryBase, exists := queries[queryName]
	if !exists {
		queryBase = queryName
	}

	var queryFinal string
	if exists {
		queryFinal = fmt.Sprintf("(%s) %s", queryBase, exclusionGlobal)
		fmt.Printf("\n🔍 Executing preset: [%s]\n", queryName)
	} else {
		queryFinal = queryBase
		fmt.Printf("\n🔍 Executing custom filter\n")
	}
	fmt.Printf("💻 Sent query: %s\n", queryFinal)

	var allTasks []models.FilteredTask
	cursor := ""
	pageCount := 0

	for {
		if pageCount >= maxPages {
			fmt.Printf("⚠️ Warning: Reached maximum pagination limit (%d pages). Some tasks might be missing.\n", maxPages)
			break
		}

		apiResp, err := f.Client.FilterTasks(queryFinal, cursor)
		if err != nil {
			return err
		}

		allTasks = append(allTasks, apiResp.Results...)

		if apiResp.NextCursor == "" {
			break
		}
		cursor = apiResp.NextCursor
		pageCount++
	}

	if len(allTasks) == 0 {
		fmt.Println("   🤷‍♂️ Inbox zero. No tasks found for this filter.")
		return nil
	}

	fmt.Printf("   🎯 Found %d tasks:\n", len(allTasks))

	idToName := cache.GetAllCachedProjects()

	now := time.Now()
	for _, t := range allTasks {
		fmt.Printf("      %s\n", FormatTask(t, now, idToName))
	}

	return nil
}
