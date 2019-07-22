package asana

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	workspaceGID := "12345"
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
	workspaceGID := "12345"
	pattern := fmt.Sprintf("/workspaces/%s/projects", workspaceGID)
	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"data":[]}`)
	})
	actualProjectsGIDs, err := cl.projectGIDs(workspaceGID)
	assert.Nil(err)
	expectedProjectGIDs := new([]string)
	assert.Equal(*expectedProjectGIDs, actualProjectsGIDs)
}

func TestProjectGIDsFailure(t *testing.T) {
	setup()
	defer teardown()
	workspaceGID := "12345"
	_, err := cl.projectGIDs(workspaceGID)
	assert.NotNil(t, err)
}
