package asana

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"golang.org/x/xerrors"

	"github.com/jeremy-miller/standup-reporter/internal/configuration"
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

func Report(config *configuration.Configuration) error {
	fmt.Println("Gathering Asana data")
	workspaceGID, err := workspaceGID(config)
	if err != nil {
		return xerrors.Errorf("error retrieving workspace: %w", err)
	}
	if workspaceGID == "" {
		return xerrors.New("no workspace")
	}
	projectGIDs, err := projectGIDs(workspaceGID, config)
	if err != nil {
		return xerrors.Errorf("error retrieving projects: %w", err)
	}
	if len(projectGIDs) == 0 {
		return xerrors.New("no projects in workspace")
	}
	tasks, err := allTasks(projectGIDs, config)
	if err != nil {
		return xerrors.Errorf("error retrieving tasks: %w", err)
	}
	printCompletedTasks(tasks, config)
	printIncompleteTasks(tasks)
	return nil
}

func workspaceGID(config *configuration.Configuration) (string, error) {
	fmt.Println("Getting Asana workspace")
	url := "https://app.asana.com/api/1.0/workspaces"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", xerrors.Errorf("error creating workspace request: %w", err)
	}
	req.Header.Set("Authorization", config.AuthHeader)
	res, err := config.Client.Do(req)
	if err != nil {
		return "", xerrors.Errorf("error requesting workspace: %w", err)
	}
	defer res.Body.Close()
	var resp response
	if err = json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return "", xerrors.Errorf("error decoding workspace response: %w", err)
	}
	return resp.Data[0].Gid, nil
}

func projectGIDs(workspaceGID string, config *configuration.Configuration) ([]string, error) {
	fmt.Println("Getting projects in workspace")
	url := fmt.Sprintf("https://app.asana.com/api/1.0/workspaces/%s/projects", workspaceGID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, xerrors.Errorf("error creating projects request: %w", err)
	}
	req.Header.Set("Authorization", config.AuthHeader)
	res, err := config.Client.Do(req)
	if err != nil {
		return nil, xerrors.Errorf("error requesting projects: %w", err)
	}
	defer res.Body.Close()
	var response response
	if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, xerrors.Errorf("error decoding projects response: %w", err)
	}
	var projectGIDs []string
	for _, project := range response.Data {
		projectGIDs = append(projectGIDs, project.Gid)
	}
	return projectGIDs, nil
}

func allTasks(projectGIDs []string, config *configuration.Configuration) ([]task, error) {
	fmt.Println("Getting tasks for all Asana projects")
	var tasks []task
	for _, projectGID := range projectGIDs {
		url := fmt.Sprintf("https://app.asana.com/api/1.0/projects/%s/tasks?opt_fields=name,completed,completed_at&completed_since=%s", projectGID, config.EarliestDate)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Printf("error creating task request for project %s: %+v", projectGID, err)
			continue
		}
		req.Header.Set("Authorization", config.AuthHeader)
		res, err := config.Client.Do(req)
		if err != nil {
			fmt.Printf("error requesting tasks for project %s: %+v", projectGID, err)
			continue
		}
		defer res.Body.Close()
		var taskResponse taskResponse
		if err = json.NewDecoder(res.Body).Decode(&taskResponse); err != nil {
			fmt.Printf("error decoding tasks response for project %s: %+v", projectGID, err)
			continue
		}
		tasks = append(tasks, taskResponse.Data...)
	}
	return tasks, nil
}

func printCompletedTasks(tasks []task, config *configuration.Configuration) {
	var completedTasks []task
	for _, task := range tasks {
		if task.Completed && task.CompletedAt.Before(config.TodayMidnight) {
			completedTasks = append(completedTasks, task)
		}
	}
	sort.Slice(completedTasks, func(i, j int) bool { return completedTasks[i].CompletedAt.Before(completedTasks[j].CompletedAt) })
	fmt.Println("\nYesterday's Activity:")
	for _, task := range completedTasks {
		fmt.Println("-", task.Name)
	}
}

func printIncompleteTasks(tasks []task) {
	var incompleteTasks []task
	for _, task := range tasks {
		if !task.Completed {
			incompleteTasks = append(incompleteTasks, task)
		}
	}
	fmt.Println("\nToday's Planned Activity:")
	for _, task := range incompleteTasks {
		fmt.Println("-", task.Name)
	}
}
