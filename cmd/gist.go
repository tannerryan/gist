// Copyright (c) 2019 Tanner Ryan. All rights reserved.
// Use of this source code is governed by a BSD-style
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
	"github.com/urfave/cli/v2"    // Copyright (c) 2016 Jeremy Saenz. All rights reserved.
)

const (
	appName    = "gist"
	appUsage   = "unofficial toolkit for file uploads to GitHub gist"
	appVersion = "2.0.1"
	// appAuthor  = "Tanner Ryan (https://github.com/TheTannerRyan/gist)"
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
		&cli.StringFlag{
			Name:    "token",
			Aliases: []string{"t"},
			Usage:   "required GitHub Gist access token",
			EnvVars: []string{"GIST_TOKEN"},
		},
		&cli.BoolFlag{
			Name:    "clipboard",
			Aliases: []string{"c"},
			Usage:   "read from clipboard",
		},
		&cli.StringFlag{
			Name:        "name",
			Aliases:     []string{"n"},
			Usage:       "comma separated file name override for Gist",
			Destination: &fileNames,
		},
		&cli.StringFlag{
			Name:        "description",
			Aliases:     []string{"d"},
			Usage:       "gist description",
			Destination: &gistDescription,
		},
	}
	app.Commands = []*cli.Command{
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
		{
			Name:    "license",
			Aliases: []string{"l"},
			Usage:   "show licensing information",
			Action: func(c *cli.Context) error {
				// execute license
				return cmdLicense(c)
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
	app.Authors = []*cli.Author{
		{
			Name:  "Tanner Ryan",
			Email: "tanner@txryan.com",
		},
	}
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

	// determine input mode
	switch mode := checkInputMode(c.Args(), c.Bool("clipboard")); mode {
	case modeStdin:
		if err := execStdin(c, overwrittenNames, &files); err != nil {
			return err
		}
	case modeGlobs:
		if err := execGlobs(c, overwrittenNames, &files); err != nil {
			return err
		}
	case modeClipboard:
		if err := execClipboard(c, overwrittenNames, &files); err != nil {
			return err
		}
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

// execStdin is triggered when stdin input is provided. It will read the data
// from stdin and update the file array. It may return an error.
func execStdin(c *cli.Context, names []string, files *[]*file) error {
	// return error if more than 1 file name override is defined
	if len(names) > 1 {
		return errExtraNames
	}

	// gist file name for stdin (default "gistfile1.txt")
	fileName := "gistfile1.txt"
	if len(names) == 1 {
		fileName = names[0]
	}

	// buffer lines content from stdout
	var lines []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// update files to contain single file (stdin)
	*files = []*file{
		{
			Name:    fileName,
			Content: strings.Join(lines, "\n"),
		},
	}

	fmt.Printf("Uploading %s as %s\n", "stdin", fileName)
	return nil
}

// execGlobs is triggered when glob input is provided. It will read the data
// from the globs and update the file array. It may return an error.
func execGlobs(c *cli.Context, names []string, files *[]*file) error {
	// return error if more overrides are defined than inputs
	if len(names) > c.Args().Len() {
		return errExtraNames
	}

	// length of user provided file names
	namesLength := len(names)
	// read each globbed file
	for i, glob := range c.Args().Slice() {
		contents, err := ioutil.ReadFile(glob)
		if err != nil {
			fmt.Println("Failed to read " + glob)
			return errFileRead
		}

		fileName := glob
		if i+1 <= namesLength {
			// insert custom file name
			fileName = names[i]
		} else {
			// only keep file name (strip preceding directory)
			parts := strings.Split(fileName, "/")
			partsLength := len(parts)
			if partsLength > 1 {
				fileName = parts[partsLength-1]
			}
		}
		// create new file entity
		file := &file{
			Name:    fileName,
			Content: string(contents),
		}
		*files = append(*files, file)

		fmt.Printf("Uploading %s as %s\n", glob, fileName)
	}

	return nil
}

// execClipboard is triggered when clipboard flag is provided. It will read the
// data from the clipboard and update the file array. It may return an error.
func execClipboard(c *cli.Context, names []string, files *[]*file) error {
	// return error if more than 1 file name override is defined
	if len(names) > 1 {
		return errExtraNames
	}

	// gist file name for stdin (default "gistfile1.txt")
	fileName := "gistfile1.txt"
	if len(names) == 1 {
		fileName = names[0]
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
	*files = []*file{
		{
			Name:    fileName,
			Content: pastedText,
		},
	}

	fmt.Printf("Uploading %s as %s\n", "clipboard", fileName)
	return nil
}

// cmdLicense is triggerd on license command
func cmdLicense(c *cli.Context) error {
	fmt.Println(`BSD 2-Clause License

Copyright (c) 2019 Tanner Ryan. All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this
   list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
   this list of conditions and the following disclaimer in the documentation
   and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

---

Ato Araki's Go clipboard library is under a BSD 3-clause license:
https://github.com/atotto/clipboard/blob/master/LICENSE

Jeremy Saenz's Go command line library is under a MIT license:
https://github.com/urfave/cli/blob/master/LICENSE

Once again, all rights reserved.`)
	return nil
}
