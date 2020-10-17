package main

import (
	"context"
	"net/http"
	"strconv"

	"github.com/10th-ndn-hackathon/namevis/ndnparse"
	"github.com/10th-ndn-hackathon/namevis/pcapinput"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

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

	parsed := ndnparse.Parse(ctx, source)
	for packet := range parsed {
		e = conn.WriteJSON(packet)
		if e != nil {
			break
		}
	}
	conn.Close()
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
