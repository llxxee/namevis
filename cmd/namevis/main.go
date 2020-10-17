package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/10th-ndn-hackathon/namevis/ndnparse"
	"github.com/10th-ndn-hackathon/namevis/pcapinput"
	"github.com/urfave/cli/v2"
)

var (
	pcapinputConfig pcapinput.Config
)

func main() {
	app := &cli.App{
		Name: "namevis",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "device",
				Aliases:     []string{"i"},
				Usage:       "network device name for live capture",
				Destination: &pcapinputConfig.Device,
			},
			&cli.IntFlag{
				Name:        "snaplen",
				Aliases:     []string{"s"},
				Usage:       "capture length for live capture",
				Destination: &pcapinputConfig.SnapLen,
			},
			&cli.StringFlag{
				Name:        "filename",
				Aliases:     []string{"r"},
				Usage:       "PCAP filename",
				Destination: &pcapinputConfig.Filename,
			},
		},
		Action: appMain,
	}
	e := app.Run(os.Args)
	if e != nil {
		log.Fatal(e)
	}
}

func appMain(c *cli.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	source, e := pcapinput.Open(ctx, pcapinputConfig)
	if e != nil {
		return e
	}

	parsed := ndnparse.Parse(ctx, source)

	for packet := range parsed {
		fmt.Println(packet)
	}
	return nil
}
