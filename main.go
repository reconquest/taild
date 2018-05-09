package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"github.com/docopt/docopt-go"
	"github.com/gorilla/websocket"
	"github.com/reconquest/karma-go"
	"github.com/reconquest/sign-go"
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

	server := &http.Server{
		Addr:    args["--listen"].(string),
		Handler: handler,
	}

	go sign.Notify(func(signal os.Signal) bool {
		log.Printf("got signal, shutting down service")

		ctx, _ := context.WithTimeout(context.Background(), time.Second*20)

		err := server.Shutdown(ctx)
		if err != nil {
			log.Fatalln(karma.Format(err, "unable to shut down server"))
			return false
		}

		log.Printf("http service has been shut down")

		return false
	}, syscall.SIGTERM)

	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatalln(
			karma.Format(
				err,
				"unable to start http server at %q",
				server.Addr,
			),
		)
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
	cmd.Stderr = writer

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
