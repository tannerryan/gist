// Copyright (c) 2019 Tanner Ryan. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package gist is an unofficial toolkit for file uploads to GitHub gist.

Installation

macOS (via Homebrew):

    brew update
    brew install TheTannerRyan/bin/gist

Manual:

Download the latest release for your platform (Darwin/macOS, Linux, Windows):
https://github.com/TheTannerRyan/gist/releases/latest. Unpack the tar, and copy
the binary to a directory that is in the PATH. Here is an example on macOS
(Darwin):

    wget https://github.com/TheTannerRyan/gist/releases/download/v2.0.0/gist-darwin_amd64.tar.gz
    tar -xzf gist-darwin_amd64.tar.gz
    mv gist /usr/local/bin

The /usr/local/bin directory will work with most variants of UNIX. For Windows,
you will have to add the parent directory to the system path.

Configuration

To use gist, you need to create a Github personal access token. To create a
token, go to https://github.com/settings/tokens. Click the "generate new token"
button and enter any description. For the scope, just select "gist". Then click
generate token.

Once you have a token, you should set the "GIST_KEY" environment variable to the
token value. If you do not want to use an environment variable, you will have to
copy and paste the token each time you would like to upload content.

Usage

Global usage:

    $ gist --help
    NAME:
    gist - unofficial toolkit for file uploads to GitHub gist

    USAGE:
    gist [global options] command [command options] [arguments...]

    VERSION:
    2.0.0

    AUTHOR:
    Tanner Ryan (https://github.com/TheTannerRyan/gist)

    COMMANDS:
        public, p   upload one or more public files
        secret, s   upload one or more secret files (shh! it's a secret)
        license, l  show licensing information
        help, h     Shows a list of commands or help for one command

    GLOBAL OPTIONS:
    --help, -h     show help
    --version, -v  print the version

Upload usage (same for secret uploads):

    $ gist p --help
    NAME:
    gist public - upload one or more public files

    USAGE:
    gist public [command options] [arguments...]

    OPTIONS:
    --token value, -t value        required GitHub Gist access token [$GIST_KEY]
    --clipboard, -c                read from clipboard
    --name value, -n value         comma separated file name override for Gist
    --description value, -d value  gist description

Aliases

All of the commands have short and long versions:

    p / public
    s / secret
    h / help

The flags also have short aliases:

    -t / --token
    -c / --clipboard
    -n / --name
    -d / --description

Examples

The interface behaves the way it looks:

    # single file (secret)
    gist s hello.txt

    # multiple files (public)
    gist p hello1.txt hello2.txt

    # all text files
    gist p *.txt

    # rename single
    gist p old.txt -n=new.txt

    # rename multiple
    gist p bad1.txt bad2.txt good3.txt -n=good1.txt,good2.txt

    # upload with gist description
    gist p story.log -d="this is my daily log"

    # upload without GIST_KEY environment variable
    gist p file.txt -t="abc123..."

    # upload from stdin
    cat network.log | gist p
    gist p < network.log

    # upload from clipboard
    gist p -c

If single or multiple files are being provided, and there are no file name
overrides, the original file names will be used. For stdin and the clipboard, if
no name is provided, the file will be uploaded as gistfile1.txt.

License

Copyright (c) 2019 Tanner Ryan. All rights reserved. Use of this source code is
governed by a BSD-style license that can be found in the LICENSE file.

Ato Araki's Go clipboard library is under a BSD 3-clause license. Jeremy Saenz's
Go command line library is under a MIT license. Once again, all rights reserved.

*/
package main
