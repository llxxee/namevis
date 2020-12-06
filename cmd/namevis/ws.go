package main

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/10th-ndn-hackathon/namevis/ndnparse"
	"github.com/10th-ndn-hackathon/namevis/pcapinput"
	"github.com/gorilla/websocket"
	"github.com/usnistgov/ndn-dpdk/ndn"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// the map of packet name -> count at the last seen second
var interestAggMap = map[string]int{}
var dataAggMap = map[string]int{}
var lastAggTime int64 = 0

const pingInterval = 5 * time.Second
const forceAggInterval = 1 * time.Second

func handleWs(cfg pcapinput.Config, w http.ResponseWriter, r *http.Request, suffixlen int) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	source, e := pcapinput.Open(ctx, cfg)
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(e.Error()))
		return
	}

	conn, e := upgrader.Upgrade(w, r, nil)
	if e != nil {
		return
	}
	defer conn.Close()

	var lastPing time.Time
	lastPong := time.Now()
	conn.SetPongHandler(func(string) error {
		log.Print("pong")
		lastPong = time.Now()
		return nil
	})

	go func() {
		for {
			_, _, e := conn.ReadMessage()
			if e != nil {
				log.Print(e)
				return
			}
		}
	}()

	pingTicker := time.NewTicker(pingInterval)
	aggTicker := time.NewTicker(forceAggInterval)
	defer func() {
		pingTicker.Stop()
		aggTicker.Stop()
	}()

	parsed := ndnparse.Parse(ctx, source)
	for {
		select {
		case packet, ok := <-parsed:
			if !ok {
				log.Print("do final aggregation before exit")
				sendAggPackets(conn)
				return
			}
			if suffixlen == 0 {
				log.Print(packet)
				conn.WriteJSON(packet)
				continue
			} else {
				// do aggregation
				stripEndIdx := len(packet.Name) - suffixlen
				if stripEndIdx < 0 {
					stripEndIdx = 0
				}
				aggPacketName := packet.Name.GetPrefix(stripEndIdx).String()
				// we use second as aggregation unit
				packetTimeSec := packet.Timestamp / 1000
				if packetTimeSec > lastAggTime {
					sendAggPackets(conn)
				}
				lastAggTime = packetTimeSec
				if packet.Type == "I" {
					interestAggMap[aggPacketName]++
				} else if packet.Type == "D" {
					dataAggMap[aggPacketName]++
				}
			}

		case t := <-pingTicker.C:
			if lastPong.Before(lastPing) {
				log.Print("disconnect")
				return
			}
			conn.WriteMessage(websocket.PingMessage, []byte{0x20})
			lastPing = t
			log.Print("ping")

		case t := <-aggTicker.C:
			if lastAggTime < t.Unix() {
				log.Printf("force aggregation at %v, last agg time %v", t.Unix(), lastAggTime)
				sendAggPackets(conn)
			}
		}
	}
}

func init() {
	http.HandleFunc("/live.websocket", func(w http.ResponseWriter, r *http.Request) {
		initAggMap()

		query := r.URL.Query()
		var cfg pcapinput.Config
		cfg.Device = query.Get("device")
		cfg.SnapLen, _ = strconv.Atoi(query.Get("snaplen"))
		var suffixlen int
		suffixlen, _ = strconv.Atoi(query.Get("suffixlen"))
		handleWs(cfg, w, r, suffixlen)
	})
	http.HandleFunc("/file.websocket", func(w http.ResponseWriter, r *http.Request) {
		initAggMap()

		query := r.URL.Query()
		var cfg pcapinput.Config
		cfg.Filename = query.Get("filename")
		cfg.SnapLen, _ = strconv.Atoi(query.Get("snaplen"))
		var suffixlen int
		suffixlen, _ = strconv.Atoi(query.Get("suffixlen"))
		handleWs(cfg, w, r, suffixlen)
	})
}

func initAggMap() {
	interestAggMap = make(map[string]int)
	dataAggMap = make(map[string]int)
}

func getAggMsgAndClearAggMap() []ndnparse.Packet {
	var packets []ndnparse.Packet
	var emptySigner ndn.Name
	for name, count := range interestAggMap {
		packets = append(packets, ndnparse.Packet{
			Timestamp: lastAggTime * 1000, // the packet sent to client should be in millisec
			Name:      ndn.ParseName(name),
			Type:      "I",
			Signer:    emptySigner,
			Count:     count,
		})
	}
	for name, count := range dataAggMap {
		packets = append(packets, ndnparse.Packet{
			Timestamp: lastAggTime * 1000,
			Name:      ndn.ParseName(name),
			Type:      "D",
			Signer:    emptySigner,
			Count:     count,
		})
	}
	initAggMap()
	return packets
}

func sendAggPackets(conn *websocket.Conn) {
	packets := getAggMsgAndClearAggMap()
	for _, packet := range packets {
		log.Print("writing agg packet ", packet)
		conn.WriteJSON(packet)
	}
}
