package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"

	"github.com/docopt/docopt-go"
	"github.com/gorilla/websocket"
)

var (
	version = "[manual build]"
	usage   = "taild " + version + `

Tail daemon for given file.

Usage:
  taild [options] <file>
  taild -h | --help
  taild --version

Options:
  --listen <addr>  Listen specified address.
                    [default: :80]
  -h --help        Show this screen.
  --version        Show version.
`
)

type Handler struct {
	filename string
}

func main() {
	args, err := docopt.Parse(usage, nil, true, version, false)
	if err != nil {
		panic(err)
	}

	var (
		filename = args["<file>"].(string)
	)

	handler := &Handler{
		filename: filename,
	}

	http.Handle("/", handler)

	err = http.ListenAndServe(args["--listen"].(string), nil)
	if err != nil {
		log.Fatal(err)
	}
}

func (handler *Handler) ServeHTTP(
	response http.ResponseWriter,
	request *http.Request,
) {
	upgrader := &websocket.Upgrader{
		ReadBufferSize:  1,
		WriteBufferSize: 1,
	}

	connection, err := upgrader.Upgrade(response, request, nil)
	if err != nil {
		http.Error(response, err.Error(), http.StatusInternalServerError)
		return
	}

	defer connection.Close()

	writer, err := connection.NextWriter(websocket.TextMessage)
	if err != nil {
		log.Println(err)
		return
	}

	query := request.URL.Query()

	params := []string{}

	if query.Get("f") == "1" {
		params = append(params, "-f")
	}

	rawLines := query.Get("n")
	if rawLines != "" {
		lines, err := strconv.Atoi(rawLines)
		if err != nil {
			fmt.Fprintln(writer, err)
			return
		}

		params = append(params, "-n", fmt.Sprint(lines))
	}

	params = append(params, handler.filename)

	log.Printf("%q", append([]string{"tail"}, params...))

	cmd := exec.Command("tail", params...)
	cmd.Stdout = writer

	err = cmd.Start()
	if err != nil {
		fmt.Fprintln(writer, err)
		return
	}

	defer writer.Close()

	err = cmd.Wait()
	if err != nil {
		fmt.Fprintln(writer, err)
		return
	}
}
