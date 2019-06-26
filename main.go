package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type AsanaResponse struct {
	Data []Response `json:"data"`
}

type Response struct {
	Gid string `json:"gid"`
}

type TaskResponse struct {
	Data []Task `json:"data"`
}

type Task struct {
	Completed   bool      `json:"completed"`
	CompletedAt time.Time `json:"completed_at"`
	Name        string    `json:"name"`
}

func main() {
	accessToken := os.Getenv("ASANA_PERSONAL_ACCESS_TOKEN")
	authHeader := fmt.Sprintf("Bearer %s", accessToken)
	client := http.Client{}

	workspaceGID := getWorkspaceGID(&client, authHeader)
	if workspaceGID == "" {
		panic("No workspace gid")
	}
	projectGIDs := getProjectGIDs(&client, authHeader, workspaceGID)
	if len(projectGIDs) == 0 {
		panic("No projects in workspace")
	}
	tasks := getAllTasks(&client, authHeader, projectGIDs)
	for _, task := range tasks {
		fmt.Println(task.Name)
	}
}

func getWorkspaceGID(client *http.Client, authHeader string) string {
	url := "https://app.asana.com/api/1.0/workspaces"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Authorization", authHeader)
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	var asanaResponse AsanaResponse
	if err = json.NewDecoder(res.Body).Decode(&asanaResponse); err != nil {
		log.Fatal(err)
	}
	return asanaResponse.Data[0].Gid
}

func getProjectGIDs(client *http.Client, authHeader string, workspaceGID string) []string {
	url := fmt.Sprintf("https://app.asana.com/api/1.0/workspaces/%s/projects", workspaceGID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Authorization", authHeader)
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	var asanaResponse AsanaResponse
	if err = json.NewDecoder(res.Body).Decode(&asanaResponse); err != nil {
		log.Fatal(err)
	}
	var projectGIDs []string
	for _, project := range asanaResponse.Data {
		projectGIDs = append(projectGIDs, project.Gid)
	}
	return projectGIDs
}

func getAllTasks(client *http.Client, authHeader string, projectGIDs []string) []Task {
	var tasks []Task
	for _, projectGID := range projectGIDs {
		completedSince := time.Now().UTC().AddDate(0, 0, -1).Format(time.RFC3339)
		url := fmt.Sprintf("https://app.asana.com/api/1.0/projects/%s/tasks?opt_fields=name,completed,completed_at&completed_since=%s", projectGID, completedSince)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Set("Authorization", authHeader)
		res, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer res.Body.Close()
		var taskResponse TaskResponse
		if err = json.NewDecoder(res.Body).Decode(&taskResponse); err != nil {
			log.Fatal(err)
		}
		tasks = append(tasks, taskResponse.Data...)
	}
	return tasks
}
