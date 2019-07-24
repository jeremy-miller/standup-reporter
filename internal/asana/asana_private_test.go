package asana

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/jeremy-miller/standup-reporter/internal/configuration"
)

var (
	cl     *client          //nolint:gochecknoglobals
	mux    *http.ServeMux   //nolint:gochecknoglobals
	server *httptest.Server //nolint:gochecknoglobals
)

type testObj struct {
	Gid  string `json:"gid"`
	Name string `json:"name"`
}

func setup() {
	authToken := "123abc"
	cl = getClient(authToken)
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)
	u, _ := url.Parse(server.URL) //nolint:errcheck
	cl.baseURL = u
}

func teardown() {
	server.Close()
}

func TestGetClient(t *testing.T) {
	assert := assert.New(t)
	authToken := "123abc"
	defaultBaseURL := "https://app.asana.com/api/1.0/"
	c := getClient(authToken)
	assert.Equal(authToken, c.authToken)
	assert.Equal(defaultBaseURL, c.baseURL.String())
	assert.IsType(http.Client{}, c.client)
	assert.Equal(time.Second*10, c.client.Timeout)
}

func TestRequestSuccess(t *testing.T) {
	setup()
	defer teardown()
	assert := assert.New(t)
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"data":[
			{"gid":"1","name":"test"}
		]}`)
	})
	responseObj := new([]testObj)
	err := cl.request(context.Background(), "test", responseObj)
	assert.Nil(err)
	expected := []testObj{
		{Gid: "1", Name: "test"},
	}
	assert.Equal(&expected, responseObj)
}

func TestRequestInvalidPath(t *testing.T) {
	responseObj := new([]testObj)
	err := cl.request(context.Background(), ":", responseObj)
	expected := "error parsing relative path \":\": parse :: missing protocol scheme"
	assert.EqualError(t, err, expected)
}

func TestRequestInvalidDo(t *testing.T) {
	setup()
	defer teardown()
	responseObj := new([]testObj)
	err := cl.request(context.Background(), "test", responseObj)
	assert.NotNil(t, err)
}

func TestRequestInvalidJSON(t *testing.T) {
	setup()
	defer teardown()
	assert := assert.New(t)
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"data":[
			{"gid":1,"name":"test"}
		]}`)
	})
	responseObj := new([]testObj)
	err := cl.request(context.Background(), "test", responseObj)
	assert.Error(err)
}

func TestWorkspaceGIDSuccess(t *testing.T) {
	setup()
	defer teardown()
	assert := assert.New(t)
	mux.HandleFunc("/workspaces", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"data":[
			{"gid":"1","name":"Workspace 1"}
		]}`)
	})
	actualGID, err := cl.workspaceGID()
	assert.Nil(err)
	expectedGID := "1"
	assert.Equal(expectedGID, actualGID)
}

func TestWorkspaceGIDFailure(t *testing.T) {
	setup()
	defer teardown()
	_, err := cl.workspaceGID()
	assert.NotNil(t, err)
}

func TestProjectGIDsSuccessSomeProjects(t *testing.T) {
	setup()
	defer teardown()
	assert := assert.New(t)
	const workspaceGID = "12345"
	pattern := fmt.Sprintf("/workspaces/%s/projects", workspaceGID)
	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"data":[
			{"gid":"1","name":"Project 1"},
			{"gid":"2","name":"Project 2"}
		]}`)
	})
	actualProjectsGIDs, err := cl.projectGIDs(workspaceGID)
	assert.Nil(err)
	expectedProjectGIDs := []string{"1", "2"}
	assert.Equal(expectedProjectGIDs, actualProjectsGIDs)
}

func TestProjectGIDsSuccessNoProjects(t *testing.T) {
	setup()
	defer teardown()
	assert := assert.New(t)
	const workspaceGID = "12345"
	pattern := fmt.Sprintf("/workspaces/%s/projects", workspaceGID)
	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"data":[]}`)
	})
	actualProjectsGIDs, err := cl.projectGIDs(workspaceGID)
	assert.Nil(err)
	var expectedProjectGIDs []string
	assert.Equal(expectedProjectGIDs, actualProjectsGIDs)
}

func TestProjectGIDsFailure(t *testing.T) {
	setup()
	defer teardown()
	const workspaceGID = "12345"
	_, err := cl.projectGIDs(workspaceGID)
	assert.NotNil(t, err)
}

func TestAllTasksOneProjectNoTasks(t *testing.T) {
	setup()
	defer teardown()
	now := time.Now().Local()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	var wg sync.WaitGroup
	conf := &configuration.Configuration{
		TodayMidnight: midnight,
		EarliestDate:  midnight.AddDate(0, 0, -1).Format(time.RFC3339),
		WG:            &wg,
	}
	projectGIDs := []string{"1"}
	pattern := fmt.Sprintf("/projects/%s/tasks", projectGIDs[0])
	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"data":[]}`)
	})
	actualTasks := cl.allTasks(projectGIDs, conf)
	var expectedTasks []task
	assert.Equal(t, expectedTasks, actualTasks)
}

func TestAllTasksOneProjectSomeTasks(t *testing.T) {
	setup()
	defer teardown()
	now := time.Now().Local()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	var wg sync.WaitGroup
	conf := &configuration.Configuration{
		TodayMidnight: midnight,
		EarliestDate:  midnight.AddDate(0, 0, -1).Format(time.RFC3339),
		WG:            &wg,
	}
	projectGIDs := []string{"1"}
	pattern := fmt.Sprintf("/projects/%s/tasks", projectGIDs[0])
	completedAt := time.Date(now.Year(), now.Month(), now.Day()-1, 12, 0, 0, 0, time.Local)
	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"data":[
			{"completed":true,"completed_at":"%s","name":"Task 1"}
		]}`, completedAt.Format(time.RFC3339))
	})
	actualTasks := cl.allTasks(projectGIDs, conf)
	expectedTasks := []task{
		{Completed: true, CompletedAt: completedAt, Name: "Task 1"},
	}
	assert.Equal(t, expectedTasks, actualTasks)
}

func TestAllTasksOneProjectEmptyTask(t *testing.T) {
	setup()
	defer teardown()
	now := time.Now().Local()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	var wg sync.WaitGroup
	conf := &configuration.Configuration{
		TodayMidnight: midnight,
		EarliestDate:  midnight.AddDate(0, 0, -1).Format(time.RFC3339),
		WG:            &wg,
	}
	projectGIDs := []string{"1"}
	pattern := fmt.Sprintf("/projects/%s/tasks", projectGIDs[0])
	completedAt := time.Date(now.Year(), now.Month(), now.Day()-1, 12, 0, 0, 0, time.Local)
	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"data":[
			{"completed":true,"completed_at":"%s","name":""}
		]}`, completedAt.Format(time.RFC3339))
	})
	actualTasks := cl.allTasks(projectGIDs, conf)
	var expectedTasks []task
	assert.Equal(t, expectedTasks, actualTasks)
}

func TestAllTasksOneProjectEmptyTaskNonEmptyTask(t *testing.T) {
	setup()
	defer teardown()
	now := time.Now().Local()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	var wg sync.WaitGroup
	conf := &configuration.Configuration{
		TodayMidnight: midnight,
		EarliestDate:  midnight.AddDate(0, 0, -1).Format(time.RFC3339),
		WG:            &wg,
	}
	projectGIDs := []string{"1"}
	pattern := fmt.Sprintf("/projects/%s/tasks", projectGIDs[0])
	completedAt := time.Date(now.Year(), now.Month(), now.Day()-1, 12, 0, 0, 0, time.Local)
	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"data":[
			{"completed":true,"completed_at":"%s","name":""},
			{"completed":true,"completed_at":"%s","name":"Task 1"}
		]}`, completedAt.Format(time.RFC3339), completedAt.Format(time.RFC3339))
	})
	actualTasks := cl.allTasks(projectGIDs, conf)
	expectedTasks := []task{
		{Completed: true, CompletedAt: completedAt, Name: "Task 1"},
	}
	assert.Equal(t, expectedTasks, actualTasks)
}

func TestAllTasksOneProjectFailure(t *testing.T) {
	setup()
	defer teardown()
	assert := assert.New(t)
	oldStdout := os.Stdout
	r, w, _ := os.Pipe() //nolint:errcheck
	os.Stdout = w
	now := time.Now().Local()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	var wg sync.WaitGroup
	conf := &configuration.Configuration{
		TodayMidnight: midnight,
		EarliestDate:  midnight.AddDate(0, 0, -1).Format(time.RFC3339),
		WG:            &wg,
	}
	projectGIDs := []string{"1"}
	actualTasks := cl.allTasks(projectGIDs, conf)
	outputChan := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r) //nolint:errcheck
		outputChan <- buf.String()
	}()
	w.Close()
	os.Stdout = oldStdout
	actualOutput := <-outputChan
	var expectedTasks []task
	assert.Equal(expectedTasks, actualTasks)
	const expectedOutput = "error requesting tasks for project 1"
	assert.Contains(actualOutput, expectedOutput)
}

func TestAllTasksMultipleProjectsAllTasks(t *testing.T) {
	setup()
	defer teardown()
	now := time.Now().Local()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	var wg sync.WaitGroup
	conf := &configuration.Configuration{
		TodayMidnight: midnight,
		EarliestDate:  midnight.AddDate(0, 0, -1).Format(time.RFC3339),
		WG:            &wg,
	}
	projectGIDs := []string{"1", "2"}
	completedAt := time.Date(now.Year(), now.Month(), now.Day()-1, 12, 0, 0, 0, time.Local)
	pattern1 := fmt.Sprintf("/projects/%s/tasks", projectGIDs[0])
	mux.HandleFunc(pattern1, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"data":[
			{"completed":true,"completed_at":"%s","name":"Task 1"}
		]}`, completedAt.Format(time.RFC3339))
	})
	pattern2 := fmt.Sprintf("/projects/%s/tasks", projectGIDs[1])
	mux.HandleFunc(pattern2, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"data":[
			{"completed":true,"completed_at":"%s","name":"Task 2"}
		]}`, completedAt.Format(time.RFC3339))
	})
	actualTasks := cl.allTasks(projectGIDs, conf)
	expectedTasks := []task{
		{Completed: true, CompletedAt: completedAt, Name: "Task 1"},
		{Completed: true, CompletedAt: completedAt, Name: "Task 2"},
	}
	assert.ElementsMatch(t, expectedTasks, actualTasks)
}

func TestAllTasksMultipleProjectsSomeTasksSomeNone(t *testing.T) {
	setup()
	defer teardown()
	now := time.Now().Local()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	var wg sync.WaitGroup
	conf := &configuration.Configuration{
		TodayMidnight: midnight,
		EarliestDate:  midnight.AddDate(0, 0, -1).Format(time.RFC3339),
		WG:            &wg,
	}
	projectGIDs := []string{"1", "2"}
	completedAt := time.Date(now.Year(), now.Month(), now.Day()-1, 12, 0, 0, 0, time.Local)
	pattern1 := fmt.Sprintf("/projects/%s/tasks", projectGIDs[0])
	mux.HandleFunc(pattern1, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"data":[
			{"completed":true,"completed_at":"%s","name":"Task 1"}
		]}`, completedAt.Format(time.RFC3339))
	})
	pattern2 := fmt.Sprintf("/projects/%s/tasks", projectGIDs[1])
	mux.HandleFunc(pattern2, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"data":[]}`)
	})
	actualTasks := cl.allTasks(projectGIDs, conf)
	expectedTasks := []task{
		{Completed: true, CompletedAt: completedAt, Name: "Task 1"},
	}
	assert.ElementsMatch(t, expectedTasks, actualTasks)
}

func TestAllTasksMultipleProjectsSomeTasksSomeError(t *testing.T) {
	setup()
	defer teardown()
	assert := assert.New(t)
	oldStdout := os.Stdout
	r, w, _ := os.Pipe() //nolint:errcheck
	os.Stdout = w
	now := time.Now().Local()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	var wg sync.WaitGroup
	conf := &configuration.Configuration{
		TodayMidnight: midnight,
		EarliestDate:  midnight.AddDate(0, 0, -1).Format(time.RFC3339),
		WG:            &wg,
	}
	projectGIDs := []string{"1", "2"}
	completedAt := time.Date(now.Year(), now.Month(), now.Day()-1, 12, 0, 0, 0, time.Local)
	pattern := fmt.Sprintf("/projects/%s/tasks", projectGIDs[0])
	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"data":[
			{"completed":true,"completed_at":"%s","name":"Task 1"}
		]}`, completedAt.Format(time.RFC3339))
	})
	actualTasks := cl.allTasks(projectGIDs, conf)
	outputChan := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r) //nolint:errcheck
		outputChan <- buf.String()
	}()
	w.Close()
	os.Stdout = oldStdout
	actualOutput := <-outputChan
	expectedTasks := []task{
		{Completed: true, CompletedAt: completedAt, Name: "Task 1"},
	}
	assert.ElementsMatch(expectedTasks, actualTasks)
	const expectedOutput = "error requesting tasks for project 2"
	assert.Contains(actualOutput, expectedOutput)
}

func TestAllTasksMultipleProjectsSomeNoneSomeError(t *testing.T) {
	setup()
	defer teardown()
	assert := assert.New(t)
	oldStdout := os.Stdout
	r, w, _ := os.Pipe() //nolint:errcheck
	os.Stdout = w
	now := time.Now().Local()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	var wg sync.WaitGroup
	conf := &configuration.Configuration{
		TodayMidnight: midnight,
		EarliestDate:  midnight.AddDate(0, 0, -1).Format(time.RFC3339),
		WG:            &wg,
	}
	projectGIDs := []string{"1", "2"}
	pattern := fmt.Sprintf("/projects/%s/tasks", projectGIDs[0])
	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"data":[]}`)
	})
	actualTasks := cl.allTasks(projectGIDs, conf)
	outputChan := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r) //nolint:errcheck
		outputChan <- buf.String()
	}()
	w.Close()
	os.Stdout = oldStdout
	actualOutput := <-outputChan
	var expectedTasks []task
	assert.ElementsMatch(expectedTasks, actualTasks)
	const expectedOutput = "error requesting tasks for project 2"
	assert.Contains(actualOutput, expectedOutput)
}

func TestAllTasksMultipleProjectsAllError(t *testing.T) {
	setup()
	defer teardown()
	assert := assert.New(t)
	oldStdout := os.Stdout
	r, w, _ := os.Pipe() //nolint:errcheck
	os.Stdout = w
	now := time.Now().Local()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	var wg sync.WaitGroup
	conf := &configuration.Configuration{
		TodayMidnight: midnight,
		EarliestDate:  midnight.AddDate(0, 0, -1).Format(time.RFC3339),
		WG:            &wg,
	}
	projectGIDs := []string{"1", "2"}
	actualTasks := cl.allTasks(projectGIDs, conf)
	outputChan := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r) //nolint:errcheck
		outputChan <- buf.String()
	}()
	w.Close()
	os.Stdout = oldStdout
	actualOutput := <-outputChan
	var expectedTasks []task
	assert.ElementsMatch(expectedTasks, actualTasks)
	const expectedOutput2 = "error requesting tasks for project 2"
	assert.Contains(actualOutput, expectedOutput2)
}

func TestPrintCompletedTasksNoTasks(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe() //nolint:errcheck
	os.Stdout = w
	var tasks []task
	now := time.Now().Local()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	var wg sync.WaitGroup
	conf := &configuration.Configuration{
		TodayMidnight: midnight,
		EarliestDate:  midnight.AddDate(0, 0, -1).Format(time.RFC3339),
		WG:            &wg,
	}
	printCompletedTasks(tasks, conf)
	outputChan := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r) //nolint:errcheck
		outputChan <- buf.String()
	}()
	w.Close()
	os.Stdout = oldStdout
	actualOutput := <-outputChan
	const expectedOutput = "\nYesterday's Activity:\n"
	assert.Equal(t, expectedOutput, actualOutput)
}

func TestPrintCompletedTasksAllAfterTodayMidnight(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe() //nolint:errcheck
	os.Stdout = w
	now := time.Now().Local()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	completedAt := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, time.Local)
	tasks := []task{
		{Completed: true, CompletedAt: completedAt, Name: "Task 1"},
		{Completed: true, CompletedAt: completedAt, Name: "Task 2"},
	}
	var wg sync.WaitGroup
	conf := &configuration.Configuration{
		TodayMidnight: midnight,
		EarliestDate:  midnight.AddDate(0, 0, -1).Format(time.RFC3339),
		WG:            &wg,
	}
	printCompletedTasks(tasks, conf)
	outputChan := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r) //nolint:errcheck
		outputChan <- buf.String()
	}()
	w.Close()
	os.Stdout = oldStdout
	actualOutput := <-outputChan
	const expectedOutput = "\nYesterday's Activity:\n"
	assert.Equal(t, expectedOutput, actualOutput)
}

func TestPrintCompletedTasksSomeAfterTodayMidnightSomeBeforeTodayMidnight(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe() //nolint:errcheck
	os.Stdout = w
	now := time.Now().Local()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	completedAt1 := time.Date(now.Year(), now.Month(), now.Day()-1, 12, 0, 0, 0, time.Local)
	completedAt2 := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, time.Local)
	tasks := []task{
		{Completed: true, CompletedAt: completedAt1, Name: "Task 1"},
		{Completed: true, CompletedAt: completedAt2, Name: "Task 2"},
	}
	var wg sync.WaitGroup
	conf := &configuration.Configuration{
		TodayMidnight: midnight,
		EarliestDate:  midnight.AddDate(0, 0, -1).Format(time.RFC3339),
		WG:            &wg,
	}
	printCompletedTasks(tasks, conf)
	outputChan := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r) //nolint:errcheck
		outputChan <- buf.String()
	}()
	w.Close()
	os.Stdout = oldStdout
	actualOutput := <-outputChan
	const expectedOutput = "\nYesterday's Activity:\n- Task 1\n"
	assert.Equal(t, expectedOutput, actualOutput)
}

func TestPrintCompletedTasksAllBeforeTodayMidnight(t *testing.T) { //nolint:dupl
	oldStdout := os.Stdout
	r, w, _ := os.Pipe() //nolint:errcheck
	os.Stdout = w
	now := time.Now().Local()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	completedAt1 := time.Date(now.Year(), now.Month(), now.Day()-1, 12, 0, 0, 0, time.Local)
	completedAt2 := time.Date(now.Year(), now.Month(), now.Day()-1, 12, 0, 0, 0, time.Local)
	tasks := []task{
		{Completed: true, CompletedAt: completedAt1, Name: "Task 1"},
		{Completed: true, CompletedAt: completedAt2, Name: "Task 2"},
	}
	var wg sync.WaitGroup
	conf := &configuration.Configuration{
		TodayMidnight: midnight,
		EarliestDate:  midnight.AddDate(0, 0, -1).Format(time.RFC3339),
		WG:            &wg,
	}
	printCompletedTasks(tasks, conf)
	outputChan := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r) //nolint:errcheck
		outputChan <- buf.String()
	}()
	w.Close()
	os.Stdout = oldStdout
	actualOutput := <-outputChan
	const expectedOutput = "\nYesterday's Activity:\n- Task 1\n- Task 2\n"
	assert.Equal(t, expectedOutput, actualOutput)
}

func TestPrintCompletedTasksNotSortedDateOrder(t *testing.T) { //nolint:dupl
	oldStdout := os.Stdout
	r, w, _ := os.Pipe() //nolint:errcheck
	os.Stdout = w
	now := time.Now().Local()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	completedAt1 := time.Date(now.Year(), now.Month(), now.Day()-1, 13, 0, 0, 0, time.Local)
	completedAt2 := time.Date(now.Year(), now.Month(), now.Day()-1, 12, 0, 0, 0, time.Local)
	tasks := []task{
		{Completed: true, CompletedAt: completedAt1, Name: "Task 1"},
		{Completed: true, CompletedAt: completedAt2, Name: "Task 2"},
	}
	var wg sync.WaitGroup
	conf := &configuration.Configuration{
		TodayMidnight: midnight,
		EarliestDate:  midnight.AddDate(0, 0, -1).Format(time.RFC3339),
		WG:            &wg,
	}
	printCompletedTasks(tasks, conf)
	outputChan := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r) //nolint:errcheck
		outputChan <- buf.String()
	}()
	w.Close()
	os.Stdout = oldStdout
	actualOutput := <-outputChan
	const expectedOutput = "\nYesterday's Activity:\n- Task 2\n- Task 1\n"
	assert.Equal(t, expectedOutput, actualOutput)
}

func TestPrintIncompleteTasksNoIncomplete(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe() //nolint:errcheck
	os.Stdout = w
	now := time.Now().Local()
	completedAt := time.Date(now.Year(), now.Month(), now.Day()-1, 13, 0, 0, 0, time.Local)
	tasks := []task{
		{Completed: true, CompletedAt: completedAt, Name: "Task 1"},
	}
	printIncompleteTasks(tasks)
	outputChan := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r) //nolint:errcheck
		outputChan <- buf.String()
	}()
	w.Close()
	os.Stdout = oldStdout
	actualOutput := <-outputChan
	const expectedOutput = "\nToday's Planned Activity:\n\n"
	assert.Equal(t, expectedOutput, actualOutput)
}

func TestPrintIncompleteTasksSomeIncompleteSomeComplete(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe() //nolint:errcheck
	os.Stdout = w
	now := time.Now().Local()
	completedAt := time.Date(now.Year(), now.Month(), now.Day()-1, 13, 0, 0, 0, time.Local)
	tasks := []task{
		{Completed: false, CompletedAt: completedAt, Name: "Task 1"},
		{Completed: true, CompletedAt: completedAt, Name: "Task 2"},
	}
	printIncompleteTasks(tasks)
	outputChan := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r) //nolint:errcheck
		outputChan <- buf.String()
	}()
	w.Close()
	os.Stdout = oldStdout
	actualOutput := <-outputChan
	const expectedOutput = "\nToday's Planned Activity:\n- Task 1\n\n"
	assert.Equal(t, expectedOutput, actualOutput)
}

func TestPrintIncompleteTasksAllIncomplete(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe() //nolint:errcheck
	os.Stdout = w
	now := time.Now().Local()
	completedAt := time.Date(now.Year(), now.Month(), now.Day()-1, 13, 0, 0, 0, time.Local)
	tasks := []task{
		{Completed: false, CompletedAt: completedAt, Name: "Task 1"},
	}
	printIncompleteTasks(tasks)
	outputChan := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r) //nolint:errcheck
		outputChan <- buf.String()
	}()
	w.Close()
	os.Stdout = oldStdout
	actualOutput := <-outputChan
	const expectedOutput = "\nToday's Planned Activity:\n- Task 1\n\n"
	assert.Equal(t, expectedOutput, actualOutput)
}
