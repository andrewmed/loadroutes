package parse

import (
	"bufio"
	"io"
	"log"
	"net"
	"strings"
)

var line int

// Parse parses all addresses/ranges on a line, if EOF return empty slice
func Parse(reader *bufio.Reader) []*net.IPNet {

	var addresses []*net.IPNet
	var ipv6 bool

	for {
		s, err := reader.ReadString('\n')
		if err == io.EOF {
			return addresses
		}
		if err != nil {
			log.Fatal(err)
		}
		line++

		// Validate line.
		if strings.HasPrefix(s, "Updated:") {
			continue
		}

		ipv6 = false

		tokens := strings.Split(s, ";")
		rawAddresses := strings.Split(tokens[0], "|")
		var ipNet *net.IPNet
		for _, rawAddress := range rawAddresses {
			addr := strings.Trim(rawAddress, " ")
			if len(addr) < 4 { // some lines are just broken
				continue
			}

			if strings.Contains(addr, "/") {
				_, ipNet, err = net.ParseCIDR(addr)
				if err != nil {
					log.Printf("Line %d: %s", line, err)
					continue
				}
			} else {
				ip := net.ParseIP(addr)
				if ip == nil {
					log.Printf("Line %d: %s", line, err)
					continue
				}
				// silently skip ipv6
				if ip.To4() == nil {
					ipv6 = true
					continue
				}
				ipNet = &net.IPNet{
					IP:   ip,
					Mask: net.IPMask{255, 255, 255, 255},
				}
			}
			addresses = append(addresses, ipNet)
		}
		// skip lines with only IPv6 addresses
		if len(addresses) > 0 || !ipv6 {
			return addresses
		}
	}
}
