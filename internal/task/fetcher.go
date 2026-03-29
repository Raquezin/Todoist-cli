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

var queries = map[string]string{
	"foco":  "today & p1 & !@reuniones",
	"radar": "7 days & @importante",
}

type Fetcher struct {
	Token string
}

func NewFetcher(token string) *Fetcher {
	return &Fetcher{Token: token}
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

	todoistClient := client.New(f.Token)

	for {
		apiResp, err := todoistClient.FilterTasks(queryFinal, cursor)
		if err != nil {
			return err
		}

		allTasks = append(allTasks, apiResp.Results...)

		if apiResp.NextCursor == "" {
			break
		}
		cursor = apiResp.NextCursor
	}

	if len(allTasks) == 0 {
		fmt.Println("   🤷‍♂️ Inbox zero. No tasks found for this filter.")
		return nil
	}

	fmt.Printf("   🎯 Found %d tasks:\n", len(allTasks))

	nameToID := cache.GetAllCachedProjects()
	projectMap := make(map[string]string, len(nameToID))
	for name, id := range nameToID {
		projectMap[id] = name
	}

	now := time.Now()
	for _, t := range allTasks {
		fmt.Printf("      %s\n", FormatTask(t, now, projectMap))
	}

	return nil
}
