package parse

import (
	"bufio"
	"io"
	"log"
	"net"
	"strings"
)

var line int

// Parse parses all addresses/ranges on a line, if EOF returns empty slice
func Parse(reader *bufio.Reader, names map[string]struct{}) []*net.IPNet {

	var addresses []*net.IPNet

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

		tokens := strings.Split(s, ";")
		if names != nil && len(tokens) > 1 {
			if name := strings.TrimSpace(tokens[1]); len(name) > 0 {
				names[name] = struct{}{}
			}
		}
		rawAddresses := strings.Split(tokens[0], "|")
		var ipNet *net.IPNet
		for _, rawAddress := range rawAddresses {
			addr := strings.Trim(rawAddress, " ")
			if len(addr) == 0 {
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
					log.Printf("Line %d: %s", line, s)
					continue
				}

				ipNet = &net.IPNet{
					IP:   ip,
					Mask: net.IPMask{255, 255, 255, 255},
				}
			}
			addresses = append(addresses, ipNet)
		}
		if len(addresses) > 0 {
			return addresses
		}
	}
}
