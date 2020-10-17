// Package pcapinput enables PCAP live capture or file reading.
package pcapinput

import (
	"context"
	"errors"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

// Error conditions.
var (
	ErrNoInputConfig = errors.New("either device or filename must be specified")
)

// Config contains PCAP input options.
type Config struct {
	Device   string `json:"device,omitempty"`
	SnapLen  int    `json:"snaplen,omitempty"`
	Filename string `json:"filename,omitempty"`
}

// Open opens PCAP handle.
func Open(ctx context.Context, cfg Config) (source gopacket.PacketDataSource, e error) {
	var hdl *pcap.Handle
	switch {
	case cfg.Device != "" && cfg.Filename == "":
		if cfg.SnapLen <= 0 {
			cfg.SnapLen = 9200
		}
		hdl, e = pcap.OpenLive(cfg.Device, int32(cfg.SnapLen), true, pcap.BlockForever)
	case cfg.Filename != "" && cfg.Device == "":
		hdl, e = pcap.OpenOffline(cfg.Filename)
	default:
		e = ErrNoInputConfig
	}
	if e != nil {
		return nil, e
	}

	go func() {
		<-ctx.Done()
		hdl.Close()
	}()

	return hdl, nil
}
