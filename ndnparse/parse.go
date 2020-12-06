// Package ndnparse parses NDN packets and extracts names.
package ndnparse

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/usnistgov/ndn-dpdk/ndn"
	"github.com/usnistgov/ndn-dpdk/ndn/tlv"
)

// Packet contains information about an NDN packet.
type Packet struct {
	Timestamp int64    `json:"timestamp"`
	Name      ndn.Name `json:"name"`
	Type      string   `json:"type"`
	Signer    ndn.Name `json:"signer"`
	Count     int      `json:"count"`
}

// Parse parses NDN packets.
func Parse(ctx context.Context, source gopacket.PacketDataSource) <-chan Packet {
	output := make(chan Packet)
	go parseLoop(ctx, source, output)
	return output
}

func parseLoop(ctx context.Context, source gopacket.PacketDataSource, output chan<- Packet) {
	var eth layers.Ethernet
	var ip layers.IPv4
	var udp layers.UDP
	var payload gopacket.Payload
	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &eth, &ip, &udp, &payload)
	decoded := []gopacket.LayerType{}
	for {
		packetData, ci, e := source.ReadPacketData()
		if errors.Is(e, io.EOF) {
			close(output)
			return
		}
		timestamp := ci.Timestamp.UnixNano() / int64(time.Millisecond)

		if e = parser.DecodeLayers(packetData, &decoded); e != nil {
			continue
		}

		var packet ndn.Packet
		if e = tlv.Decode(payload.Payload(), &packet); e != nil {
			continue
		}

		switch {
		case packet.Interest != nil:
			var signer ndn.Name
			if packet.Interest.SigInfo != nil {
				signer = packet.Interest.SigInfo.KeyLocator.Name
			}
			output <- Packet{
				Timestamp: timestamp,
				Name:      packet.Interest.Name,
				Type:      "I",
				Signer:    signer,
				Count:     1,
			}
		case packet.Data != nil:
			var signer ndn.Name
			if packet.Data.SigInfo != nil {
				signer = packet.Data.SigInfo.KeyLocator.Name
			}
			output <- Packet{
				Timestamp: timestamp,
				Name:      packet.Data.Name,
				Type:      "D",
				Signer:    signer,
				Count:     1,
			}
		}
	}
}
