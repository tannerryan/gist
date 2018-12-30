// Copyright (c) 2018 Tanner Ryan. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package gist

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/atotto/clipboard" // Copyright (c) 2013 Ato Araki. All rights reserved.
	"gopkg.in/urfave/cli.v1"      // Copyright (c) 2016 Jeremy Saenz. All rights reserved.
)

const (
	appName    = "gist"
	appUsage   = "unofficial toolkit for file uploads to GitHub gist"
	appVersion = "1.0.0"
	appAuthor  = "Tanner Ryan (https://github.com/TheTannerRyan/gist)"
)

var (
	httpClient      = &http.Client{} // HTTP client for sending requests
	fileNames       = ""             // string to possibly be populated for file name overrides
	gistDescription = ""             // string to possibly be populated with gist description
	errNoData       = errors.New("Error: no input data has been specified")
	errExtraNames   = errors.New("Error: more override file names than inputs have been provided")
	errFileRead     = errors.New("Error: cannot read all files")
	errClipboard    = errors.New("Error: cannot read data from clipboard")
	errCopyToken    = errors.New("Error: the clipboard is populated with the API token")
)

// Run is the main entrypoint for gist.
func Run() error {
	app := cli.NewApp()
	setup(app)

	flags := []cli.Flag{
		cli.StringFlag{
			Name:   "token, t",
			Usage:  "required GitHub Gist access token",
			EnvVar: "GIST_KEY",
		},
		cli.BoolFlag{
			Name:  "clipboard, c",
			Usage: "read from clipboard",
		},
		cli.StringFlag{
			Name:        "name, n",
			Usage:       "comma separated file name override for Gist",
			Destination: &fileNames,
		},
		cli.StringFlag{
			Name:        "description, d",
			Usage:       "gist description",
			Destination: &gistDescription,
		},
	}
	app.Commands = []cli.Command{
		{
			Name:    "public",
			Aliases: []string{"p"},
			Usage:   "upload one or more public files",
			Action: func(c *cli.Context) error {
				// execute public upload
				return cmdExec(c, true)
			},
			Flags: flags,
		},
		{
			Name:    "secret",
			Aliases: []string{"s"},
			Usage:   "upload one or more secret files (shh! it's a secret)",
			Action: func(c *cli.Context) error {
				// execute secret upload
				return cmdExec(c, false)
			},
			Flags: flags,
		},
	}
	// return errors
	return app.Run(os.Args)
}

func setup(app *cli.App) {
	app.Name = appName
	app.Usage = appUsage
	app.Version = appVersion
	app.Author = appAuthor
	app.EnableBashCompletion = true
}

// cmdExec is triggered on public and secret uploads
func cmdExec(c *cli.Context, public bool) error {
	// if file names are to be overwritten, get the values
	var overwrittenNames []string
	if fileNames != "" {
		overwrittenNames = strings.Split(fileNames, ",")
	}

	var files []*file

	switch mode := checkInputMode(c.Args(), c.Bool("clipboard")); mode {
	case modeStdin:
		// return error if more than 1 file name override is defined
		if len(overwrittenNames) > 1 {
			return errExtraNames
		}
		// gist file name for stdin (default "gistfile1.txt")
		fileName := "gistfile1.txt"
		if len(overwrittenNames) == 1 {
			fileName = overwrittenNames[0]
		}
		// buffer lines content from stdout
		var lines []string
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		// update files to contain single file (stdin)
		files = []*file{
			{
				Name:    fileName,
				Content: strings.Join(lines, "\n"),
			},
		}
		fmt.Printf("Uploading %s as %s\n", "stdin", fileName)

	case modeGlobs:
		// return error if more overrides are defined than inputs
		if len(overwrittenNames) > len(c.Args()) {
			return errExtraNames
		}
		// length of user provided file names
		overrideLength := len(overwrittenNames)
		// read each globbed file
		for i, glob := range c.Args() {
			contents, err := ioutil.ReadFile(glob)
			if err != nil {
				fmt.Println("failed to read " + glob)
				return errFileRead
			}
			// check and insert custom file name (if applicable)
			fileName := glob
			if i+1 <= overrideLength {
				fileName = overwrittenNames[i]
			}
			// create new file entity
			file := &file{
				Name:    fileName,
				Content: string(contents),
			}
			files = append(files, file)
			fmt.Printf("Uploading %s as %s\n", glob, fileName)
		}

	case modeClipboard:
		// return error if more than 1 file name override is defined
		if len(overwrittenNames) > 1 {
			return errExtraNames
		}
		// gist file name for stdin (default "gistfile1.txt")
		fileName := "gistfile1.txt"
		if len(overwrittenNames) == 1 {
			fileName = overwrittenNames[0]
		}
		pastedText, err := clipboard.ReadAll()
		if err != nil {
			return errClipboard
		}
		// return error if clipboard is the token
		if pastedText == c.String("token") {
			return errCopyToken
		}
		// update files to contain single file (clipboard)
		files = []*file{
			{
				Name:    fileName,
				Content: pastedText,
			},
		}
		fmt.Printf("Uploading %s as %s\n", "clipboard", fileName)

	default:
		return errNoData
	}

	// build payload, send request, print url or return error
	payload, err := jsonBuilder(gistDescription, public, files)
	if err != nil {
		return err
	}
	url, err := sendContent(payload, c.String("token"))
	if err != nil {
		return err
	}

	fmt.Println(url)
	return nil
}
