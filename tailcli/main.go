package main

import (
	"io"
	"log"
	"os"
	"strings"

	"github.com/docopt/docopt-go"
	"github.com/gorilla/websocket"
)

var (
	version = "[manual build]"
	usage   = "tailcli " + version + `

Usage:
  tailcli [options] <address> [-f] [-n <n>]
  tailcli -h | --help
  tailcli --version

Options:
  -f --follow     Follow output.
  -n --lines <n>  Output the last N lines.
  -h --help       Show this screen.
  --version       Show version.
`
)

func main() {
	args, err := docopt.Parse(usage, nil, true, version, false)
	if err != nil {
		panic(err)
	}

	address := args["<address>"].(string)
	remote := address
	if !strings.HasPrefix(remote, "ws://") {
		remote = "ws://" + remote
	}

	if !strings.HasSuffix(remote, "/") {
		remote += "/"
	}

	params := []string{}
	if args["--follow"].(bool) {
		params = append(params, "f=1")
	}

	if lines, ok := args["--lines"].(string); ok {
		params = append(params, "n="+lines)
	}

	remote += "?" + strings.Join(params, "&")

	connection, _, err := websocket.DefaultDialer.Dial(
		remote, nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("connected to %s", address)

	for {
		_, reader, err := connection.NextReader()
		if err != nil {
			log.Fatal(err)
		}

		_, err = io.Copy(os.Stdout, reader)
		if err != nil {
			log.Fatal(err)
		}
	}
}
