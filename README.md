# NDN Passive Name Visualizer

This repository is the server component of NDN Passive Name Visualizer.
It is meant to be used with the [web application](https://github.com/llxxee/namevis-web).

## Installation

This program supports Linux and Windows.
On Linux, it requires `libpcap-dev` package.
On Windows, it requires `npcap` system service.

To install from source code, you need Go 1.15 or higher.

```bash
go get github.com/10th-ndn-hackathon/namevis/cmd/namevis
```

You can also download prebuilt binaries on *Actions* tab.

## Protocol

`namevis` program starts an HTTP server on `127.0.0.1:6847`.

`GET /devices.json` returns a JSON document describing available network devices for packet capture.
The response is an array, in which each item represent a network device.
Example:

```json
[
  {
    "name": "eth1",
    "addresses": ["192.0.2.1"]
  }
]
```

`GET /files.json` returns a JSON document describing PCAP files in the current working directory.
The response is an array, in which each item is a file name.
Example:

```json
[
  "1.pcap",
  "2.pcap"
]
```

`new WebSocket("http://127.0.0.1:6847/live.websocket?device=eth1")` starts a live capture session on "eth1" interface.
Each message represents a captured NDN packet.
Example:

```jsonc
{
  "timestamp": 123456789000, // Unix timestamp in milliseconds
  "name": "/8=P/8=Q",        // NDN packet name in canonical format
  "type": "I"                // I for Interest, D for Data
}
```

`new WebSocket("http://127.0.0.1:6847/file.websocket?filename=1.pcap")` reads from the specified file.
Message format is same as live capture session.
