// Command namevis runs a WebSocket server that streams captured NDN packet names.
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	var httpListen string
	app := &cli.App{
		Name: "namevis",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "http",
				Usage:       "HTTP listen endpoint",
				Value:       "127.0.0.1:6847", // NVIS=6847
				Destination: &httpListen,
			},
		},
		Action: func(c *cli.Context) error {
			return http.ListenAndServe(httpListen, nil)
		},
	}
	e := app.Run(os.Args)
	if e != nil {
		log.Fatal(e)
	}
}
