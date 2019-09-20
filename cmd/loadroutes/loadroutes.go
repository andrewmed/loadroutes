// Program loadroutes adds routes to a network interface (probably, a vpn tunnel) directly from a dump file
// in the format as in https://github.com/zapret-info/z-i. Convenient to be called in a script upon raising the network interface.
package main

import (
	"bufio"
	"flag"
	"log"
	"os"

	"github.com/vishvananda/netlink"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"loadroutes/parse"
	"loadroutes/route"
)

var (
	version = "unknown"
	date    = "unknown"
)

func main() {
	log.Printf("loadroutes %s %s\n", version, date)

	iface := flag.String("iface", "", "Network interface name.")
	filename := flag.String("dump", "", "Path to a dump file (see https://github.com/zapret-info/z-i).")

	flag.Parse()
	if *iface == "" || *filename == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	log.Printf("Using dump file: %s", *filename)
	log.Printf("Adding to %s", *iface)

	file, err := os.Open(*filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	link, err := netlink.LinkByName(*iface)
	if err != nil {
		log.Fatalf("No such network interface: %v.\n", iface)
	}

	reader := transform.NewReader(file, charmap.Windows1251.NewDecoder())
	bReader := bufio.NewReader(reader)

	var done int

	for {
		ips := parse.Parse(bReader)
		if len(ips) == 0 {
			break // EOF
		}

		for _, ipNet := range ips {
			ip := (*ipNet).IP // hack, because add method nullifies struct
			if err := route.AddIP(link, ipNet); err != nil {
				log.Printf("Adding %v: %s", ip, err)
				continue
			}
			done += 1
			if done%100000 == 0 {
				log.Println(done)
			}
		}
	}
}
