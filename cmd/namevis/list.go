package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/10th-ndn-hackathon/namevis/pcapinput"
)

func init() {
	http.HandleFunc("/devices.json", func(w http.ResponseWriter, r *http.Request) {
		list := pcapinput.ListDevices()
		j, _ := json.Marshal(list)

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(j)
	})
	http.HandleFunc("/files.json", func(w http.ResponseWriter, r *http.Request) {
		cwd, _ := os.Getwd()
		files, _ := ioutil.ReadDir(cwd)
		var filenames []string
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			if filename := file.Name(); strings.HasSuffix(filename, ".pcap") || strings.HasSuffix(filename, ".pcapng") {
				filenames = append(filenames, filename)
			}
		}
		j, _ := json.Marshal(filenames)

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(j)
	})
}
