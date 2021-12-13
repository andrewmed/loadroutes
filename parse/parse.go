package parse

import (
	"bufio"
	"io"
	"log"
	"net"
	"regexp"
	"strings"
)

// domains up to 9 letters only in com or org tldn
// ("substantial" domains only)
const REGEXP = `^((www\.)?[a-z]{1,9}\.(com|org))$`

func Parse(reader *bufio.Reader) ([]*net.IPNet, map[string]struct{}) {
	addresses := []*net.IPNet{}
	names := map[string]struct{}{}

	re := regexp.MustCompile(REGEXP)

	var line int

	for {
		s, err := reader.ReadString('\n')
		if err == io.EOF {
			return addresses, names
		}
		if err != nil {
			log.Fatal(err)
		}
		line++

		if strings.HasPrefix(s, "Updated:") {
			continue
		}

		tokens := strings.Split(s, ";")

		// extract IP address
		rawAddresses := strings.Split(tokens[0], "|")
		var ipNet *net.IPNet
		for _, rawAddress := range rawAddresses {
			addr := strings.TrimSpace(rawAddress)
			if len(addr) == 0 {
				continue
			}

			if strings.Contains(addr, "/") {
				_, ipNet, err = net.ParseCIDR(addr)
				if err != nil {
					log.Printf("Parsing line %d: %s", line, err)
					continue
				}
			} else {
				ip := net.ParseIP(addr)
				if ip == nil {
					log.Printf("Parsing line %d: %s", line, s)
					continue
				}
				ipNet = &net.IPNet{
					IP:   ip,
					Mask: net.IPv4Mask(255, 255, 255, 255),
				}
			}
			addresses = append(addresses, ipNet)
		}

		// extract DNS name
		if name := strings.TrimSpace(strings.TrimSpace(tokens[1])); len(name) > 0 {
			m := re.FindAllString(name, 2)
			if len(m) > 0 {
				names[m[0]] = struct{}{}
			}
		}
	}
}
