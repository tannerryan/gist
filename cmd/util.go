// Copyright (c) 2018 Tanner Ryan. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package gist

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"

	"gopkg.in/urfave/cli.v1" // Copyright (c) 2016 Jeremy Saenz. All rights reserved.
)

// simplified errors when interacting with GitHub's API
var (
	errNetwork     = errors.New("Error: cannot send request to GitHub")
	errBadResponse = errors.New("Error: cannot read reply from GitHub")
	errBadAuth     = errors.New("Error: invalid API token")
)

// inputType is an enum for the type of input modes
type inputType int

// valid values for inputType
const (
	modeStdin     inputType = 0 // stdout is being provided
	modeGlobs     inputType = 1 // globs are being provided
	modeClipboard inputType = 2 // clipboard is being used
	modeError     inputType = 3 // no input is provided (error must be triggered)
)

// checkInputMode takes the cli arguments and determines the input type
func checkInputMode(args cli.Args, clip bool) inputType {
	if clip {
		return modeClipboard
	}
	if len(args) == 0 {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			return modeStdin
		}
		return modeError
	}
	return modeGlobs
}

// file is represented as a name and as content
type file struct {
	Name    string
	Content string
}

// payload is the structure for POST GitHub requests
type payload struct {
	Description string                       `json:"description"`
	Public      bool                         `json:"public"`
	Files       map[string]map[string]string `json:"files"`
}

// jsonBuilder builds the JSON to be sent in POST requests when uploading
// content. The builder will return a bytes buffer of the JSON payload, or will
// return an error.
func jsonBuilder(description string, public bool, files []*file) (*bytes.Buffer, error) {
	// map for file name + content
	fileMap := make(map[string]map[string]string)
	for _, f := range files {
		entry := map[string]string{
			"content": f.Content,
		}
		fileMap[f.Name] = entry
	}
	// assemble payload
	data := payload{
		Description: description,
		Public:      public,
		Files:       fileMap,
	}
	// encode payload
	buff := new(bytes.Buffer)
	encoder := json.NewEncoder(buff)
	err := encoder.Encode(data)
	if err != nil {
		return nil, err
	}
	return buff, nil
}

// githubResponse is used for parsing GitHub's request reply
type githubResponse struct {
	URL string `json:"html_url"`
}

// sendContent takes the formatted payload with a token and sends the request to
// GitHub's API. It will return a URL string or an error.
func sendContent(payload *bytes.Buffer, token string) (string, error) {
	if token == "" {
		return "", errBadAuth
	}

	url := "https://api.github.com/gists"
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", errNetwork
	}
	defer resp.Body.Close()

	switch status := resp.Status; status {
	case "201 Created":
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", errBadResponse
		}
		var data githubResponse
		if err := json.Unmarshal(body, &data); err != nil {
			return "", errBadResponse
		}
		return data.URL, nil

	case "401 Unauthorized":
		return "", errBadAuth
	}

	// uncommon error encountered
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errBadResponse
	}
	return "", errors.New(string(body))
}
