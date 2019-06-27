package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app        = kingpin.New("standup-reporter", "Command-line application to gather daily standup reports.")
	days       = app.Flag("days", "Number of days to go back to collect completed tasks. Default 1 day (or 3 days on Monday).").Short('d').PlaceHolder("N").Int()
	asanaToken = app.Flag("asana", "Asana Personal Access Token").Short('a').Required().PlaceHolder("TOKEN").String()
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
	app.HelpFlag.Short('h')
	kingpin.MustParse(app.Parse(os.Args[1:]))
	if *days == 0 {
		*days = calculateDays()
	}
	authHeader := fmt.Sprintf("Bearer %s", *asanaToken)
	client := http.Client{}
	t := time.Now().Local()
	todayMidnight := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
	earliestDate := todayMidnight.AddDate(0, 0, -*days).Format(time.RFC3339)

	workspaceGID := getWorkspaceGID(&client, authHeader)
	if workspaceGID == "" {
		panic("No workspace gid")
	}
	projectGIDs := getProjectGIDs(&client, authHeader, workspaceGID)
	if len(projectGIDs) == 0 {
		panic("No projects in workspace")
	}
	tasks := getAllTasks(&client, authHeader, projectGIDs, earliestDate)
	printCompletedTasks(tasks, todayMidnight)
}

func calculateDays() int {
	if time.Now().Weekday() == time.Monday { // account for weekend
		return 3
	}
	return 1
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

func getAllTasks(client *http.Client, authHeader string, projectGIDs []string, earliestDate string) []Task {
	var tasks []Task
	for _, projectGID := range projectGIDs {
		url := fmt.Sprintf("https://app.asana.com/api/1.0/projects/%s/tasks?opt_fields=name,completed,completed_at&completed_since=%s", projectGID, earliestDate)
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

func printCompletedTasks(tasks []Task, todayMidnight time.Time) {
	var completedTasks []Task
	for _, task := range tasks {
		if task.Completed && task.CompletedAt.Before(todayMidnight) {
			completedTasks = append(completedTasks, task)
		}
	}
	sort.Slice(completedTasks, func(i, j int) bool { return completedTasks[i].CompletedAt.Before(completedTasks[j].CompletedAt) })
	fmt.Println("--- Completed Tasks ---")
	for _, task := range completedTasks {
		fmt.Println("- ", task.Name)
	}
}
