package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type AsanaResponse struct {
	Data []Response `json:"data"`
}

type Response struct {
	Gid string `json:"gid"`
}

func main() {
	accessToken := os.Getenv("ASANA_PERSONAL_ACCESS_TOKEN")
	authHeader := fmt.Sprintf("Bearer %s", accessToken)
	client := http.Client{}

	workspaceGID := getWorkspaceGID(&client, authHeader)
	projectGIDs := getProjectGIDs(&client, authHeader, workspaceGID)
	fmt.Printf("%#v", projectGIDs)
}

func getWorkspaceGID(client *http.Client, authHeader string) string {
	workspaceURL := "https://app.asana.com/api/1.0/workspaces"
	req, err := http.NewRequest("GET", workspaceURL, nil)
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
	projectURL := fmt.Sprintf("https://app.asana.com/api/1.0/workspaces/%s/projects", workspaceGID)
	req, err := http.NewRequest("GET", projectURL, nil)
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
