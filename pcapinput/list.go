package pcapinput

import (
	"net"

	"github.com/google/gopacket/pcap"
)

// Device contains information about a network interface.
type Device struct {
	Name      string   `json:"name"`
	Addresses []net.IP `json:"addresses"`
}

// ListDevices returns information about available network interfaces.
func ListDevices() (list []Device) {
	ifs, e := pcap.FindAllDevs()
	if e != nil {
		return
	}
	for _, netif := range ifs {
		d := Device{
			Name: netif.Name,
		}
		for _, a := range netif.Addresses {
			d.Addresses = append(d.Addresses, a.IP)
		}
		list = append(list, d)
	}
	return list
}
