/*
Package asana contains all functionality for retrieving tasks from Asana and printing them to the screen.

Tasks from all Asana projects are used in the standup-reporter.  Unless specified, the default number of days to go back
and get tasks for is 1, except if the script is run on a Monday, in which case it will go back 3 days (to account for
the weekend).

Completed tasks are sorted oldest to most recently completed.  Only tasks which were completed between midnight of the
requested day and midnight of the current day (both in local time) are shown.

Regarding incomplete tasks, all non-complete tasks are shown and are not sorted.
*/
package asana

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"time"

	"golang.org/x/xerrors"

	"github.com/jeremy-miller/standup-reporter/internal/configuration"
)

type client struct {
	authToken string
	baseURL   *url.URL
	client    http.Client
}

type response struct {
	Data interface{} `json:"data"`
}

type entry struct {
	Gid string `json:"gid"`
}

type task struct {
	Completed   bool      `json:"completed"`
	CompletedAt time.Time `json:"completed_at"`
	Name        string    `json:"name"`
}

type taskResult struct {
	Tasks []task
	Err   error
}

/*
Report coordinates gathering of Asana task data and prints completed and incomplete tasks to the screen.
*/
func Report(authToken string, config *configuration.Configuration) error {
	fmt.Println("\nGathering Asana data...")
	client := getClient(authToken)
	workspaceGID, err := client.workspaceGID()
	if err != nil {
		return xerrors.Errorf("error retrieving workspace: %w", err)
	}
	projectGIDs, err := client.projectGIDs(workspaceGID)
	if err != nil {
		return xerrors.Errorf("error retrieving projects: %w", err)
	}
	if len(projectGIDs) == 0 {
		return xerrors.New("no projects in workspace")
	}
	tasks := client.allTasks(projectGIDs, config)
	if len(tasks) == 0 {
		return xerrors.New("no tasks available")
	}
	printCompletedTasks(tasks, config)
	printIncompleteTasks(tasks)
	return nil
}

func getClient(authToken string) *client {
	const defaultBaseURL = "https://app.asana.com/api/1.0/"
	baseURL, _ := url.Parse(defaultBaseURL) //nolint:errcheck
	return &client{
		authToken: authToken,
		baseURL:   baseURL,
		client: http.Client{
			Timeout: time.Second * 10,
		},
	}
}

func (c *client) request(ctx context.Context, path string, responseObj interface{}) error {
	relPath, err := url.Parse(path)
	if err != nil {
		return xerrors.Errorf("error parsing relative path \"%s\": %w", path, err)
	}
	fullURL := c.baseURL.ResolveReference(relPath).String()
	req, _ := http.NewRequest("GET", fullURL, nil) //nolint:errcheck
	authHeader := fmt.Sprintf("Bearer %s", c.authToken)
	req.Header.Set("Authorization", authHeader)
	res, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		return xerrors.Errorf("error requesting \"%s\": %w", fullURL, err)
	}
	defer res.Body.Close()
	parsedResponse := &response{Data: responseObj}
	if err = json.NewDecoder(res.Body).Decode(parsedResponse); err != nil {
		return xerrors.Errorf("error decoding response from \"%s\": %w", fullURL, err)
	}
	return nil
}

func (c *client) workspaceGID() (string, error) {
	ctx := context.Background()
	const path = "workspaces"
	workspaces := new([]entry)
	if err := c.request(ctx, path, workspaces); err != nil {
		return "", err
	}
	return (*workspaces)[0].Gid, nil
}

func (c *client) projectGIDs(workspaceGID string) ([]string, error) {
	ctx := context.Background()
	path := fmt.Sprintf("workspaces/%s/projects", workspaceGID)
	allProjects := new([]entry)
	if err := c.request(ctx, path, allProjects); err != nil {
		return nil, err
	}
	var projects []string
	for _, project := range *allProjects {
		projects = append(projects, project.Gid)
	}
	return projects, nil
}

func (c *client) allTasks(projectGIDs []string, config *configuration.Configuration) []task {
	results := make(chan taskResult)
	for _, projectGID := range projectGIDs {
		config.WG.Add(1)
		go projectTasks(c, projectGID, config, results)
	}
	go func() {
		config.WG.Wait()
		close(results)
	}()
	var tasks []task
	for r := range results {
		if r.Err != nil {
			fmt.Printf("%v", r.Err)
			continue
		}
		tasks = append(tasks, r.Tasks...)
	}
	return tasks
}

func projectTasks(c *client, projectGID string, config *configuration.Configuration, results chan<- taskResult) {
	defer config.WG.Done()
	ctx := context.Background()
	path := fmt.Sprintf("projects/%s/tasks?opt_fields=name,completed,completed_at&completed_since=%s", projectGID, config.EarliestDate) //nolint:lll
	var tasks []task
	if err := c.request(ctx, path, &tasks); err != nil {
		results <- taskResult{
			Tasks: nil,
			Err:   xerrors.Errorf("error requesting tasks for project %s: %v", projectGID, err),
		}
		return
	}
	filteredTasks := filterEmptyTasks(tasks)
	results <- taskResult{
		Tasks: filteredTasks,
		Err:   nil,
	}
}

func filterEmptyTasks(tasks []task) []task {
	var filteredTasks []task
	for i, task := range tasks {
		if task.Name != "" {
			filteredTasks = append(filteredTasks, tasks[i])
		}
	}
	return filteredTasks
}

func printCompletedTasks(tasks []task, config *configuration.Configuration) {
	var completedTasks []task
	for _, task := range tasks {
		if task.Completed && task.CompletedAt.Before(config.TodayMidnight) {
			completedTasks = append(completedTasks, task)
		}
	}
	sort.Slice(completedTasks, func(i, j int) bool { return completedTasks[i].CompletedAt.Before(completedTasks[j].CompletedAt) }) //nolint:lll
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
	fmt.Println()
}
