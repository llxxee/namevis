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
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

const pingInterval = 5 * time.Second

func handleWs(cfg pcapinput.Config, w http.ResponseWriter, r *http.Request) {
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
	defer pingTicker.Stop()

	parsed := ndnparse.Parse(ctx, source)
	for {
		select {
		case packet, ok := <-parsed:
			if !ok {
				return
			}
			log.Print(packet)
			conn.WriteJSON(packet)
		case t := <-pingTicker.C:
			if lastPong.Before(lastPing) {
				log.Print("disconnect")
				return
			}
			conn.WriteMessage(websocket.PingMessage, []byte{0x20})
			lastPing = t
			log.Print("ping")
		}
	}
}

func init() {
	http.HandleFunc("/live.websocket", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		var cfg pcapinput.Config
		cfg.Device = query.Get("device")
		cfg.SnapLen, _ = strconv.Atoi(query.Get("snaplen"))
		handleWs(cfg, w, r)
	})
	http.HandleFunc("/file.websocket", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		var cfg pcapinput.Config
		cfg.Filename = query.Get("filename")
		handleWs(cfg, w, r)
	})
}
