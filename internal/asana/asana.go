package asana

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/jeremy-miller/standup-reporter/internal/config"
)

type response struct {
	Data []entry `json:"data"`
}

type entry struct {
	Gid string `json:"gid"`
}

type taskResponse struct {
	Data []task `json:"data"`
}

type task struct {
	Completed   bool      `json:"completed"`
	CompletedAt time.Time `json:"completed_at"`
	Name        string    `json:"name"`
}

func Report(config *config.Config) {
	workspaceGID := workspaceGID(config)
	if workspaceGID == "" {
		panic("No workspace gid")
	}
	projectGIDs := projectGIDs(workspaceGID, config)
	if len(projectGIDs) == 0 {
		panic("No projects in workspace")
	}
	tasks := allTasks(projectGIDs, config)
	printCompletedTasks(tasks, config)
}

func workspaceGID(config *config.Config) string {
	url := "https://app.asana.com/api/1.0/workspaces"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Authorization", config.AuthHeader)
	res, err := config.Client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	var resp response
	if err = json.NewDecoder(res.Body).Decode(&resp); err != nil {
		log.Fatal(err)
	}
	return resp.Data[0].Gid
}

func projectGIDs(workspaceGID string, config *config.Config) []string {
	url := fmt.Sprintf("https://app.asana.com/api/1.0/workspaces/%s/projects", workspaceGID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Authorization", config.AuthHeader)
	res, err := config.Client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	var response response
	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		log.Fatal(err)
	}
	var projectGIDs []string
	for _, project := range response.Data {
		projectGIDs = append(projectGIDs, project.Gid)
	}
	return projectGIDs
}

func allTasks(projectGIDs []string, config *config.Config) []task {
	var tasks []task
	for _, projectGID := range projectGIDs {
		url := fmt.Sprintf("https://app.asana.com/api/1.0/projects/%s/tasks?opt_fields=name,completed,completed_at&completed_since=%s", projectGID, config.EarliestDate)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Set("Authorization", config.AuthHeader)
		res, err := config.Client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer res.Body.Close()
		var taskResponse taskResponse
		if err = json.NewDecoder(res.Body).Decode(&taskResponse); err != nil {
			log.Fatal(err)
		}
		tasks = append(tasks, taskResponse.Data...)
	}
	return tasks
}

func printCompletedTasks(tasks []task, config *config.Config) {
	var completedTasks []task
	for _, task := range tasks {
		if task.Completed && task.CompletedAt.Before(config.TodayMidnight) {
			completedTasks = append(completedTasks, task)
		}
	}
	sort.Slice(completedTasks, func(i, j int) bool { return completedTasks[i].CompletedAt.Before(completedTasks[j].CompletedAt) })
	fmt.Println("--- Completed Tasks ---")
	for _, task := range completedTasks {
		fmt.Println("- ", task.Name)
	}
}
