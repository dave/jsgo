package main

import (
	"bytes"
	"encoding/json"
	"net/http"
)

//types
type GistFile struct {
	Content string `json:"content"`
}

type Gist struct {
	Description string              `json:"description"`
	Public      bool                `json:"public"`
	Files       map[string]GistFile `json:"files"`
}

func store(content string) (string, error) {

	gist := Gist{
		Description: "gist created at jsgo.io",
		Public:      true,
		Files: map[string]GistFile{
			"main.go": {Content: string(content)},
		},
	}

	client := http.DefaultClient
	//client.Timeout = time.Second * 10

	b, err := json.Marshal(gist)
	if err != nil {
		return "", err
	}
	br := bytes.NewBuffer(b)

	resp, err := client.Post("https://api.github.com/gists", "application/json", br)
	if err != nil {
		return "", err
	}

	var response struct {
		Id string `json:"id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", err
	}
	return response.Id, nil
}
