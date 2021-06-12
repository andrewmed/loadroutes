// Program loadroutes adds routes to a network interface (probably, a vpn tunnel) directly from a dump file
// in the format as in https://github.com/zapret-info/z-i. Convenient to be called in a script upon raising the network interface.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

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

const DNS = "8.8.4.4"

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("Loadroutes", version, date)

	iface := flag.String("iface", "", "Network interface name.")
	dumpPath := flag.String("dump", "", "Path to a dump file (see https://github.com/zapret-info/z-i).")
	namesPath := flag.String("names", "", "Save extracted domain names to a specified file (optional).")
	ipv6 := flag.Bool("ipv6", false, "Process IPv6 addresses as well (by default, disabled).")

	flag.Parse()
	if *iface == "" || *dumpPath == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	log.Printf("Using dump file: %s", *dumpPath)
	log.Printf("Adding to %s:", *iface)

	dumpFile, err := os.Open(*dumpPath)
	if err != nil {
		log.Fatal(err)
	}
	defer dumpFile.Close()

	link, err := netlink.LinkByName(*iface)
	if err != nil {
		log.Fatalf("No such network interface: %v.", iface)
	}

	dumpReader := transform.NewReader(dumpFile, charmap.Windows1251.NewDecoder())
	bDumpReader := bufio.NewReader(dumpReader)

	var names map[string]struct{}
	var namesFile *os.File
	if *namesPath != "" {
		namesFile, err = os.Create(*namesPath)
		if err != nil {
			log.Fatalf("Could not open %s for saving domain names: %s.", *namesPath, err)
		}
		defer namesFile.Close()
		names = make(map[string]struct{}, 150000) // preallocate
	}

	var done int

	for {
		ips := parse.Parse(bDumpReader, names)
		if len(ips) == 0 {
			break // EOF
		}

		for _, ipNet := range ips {
			ip := (*ipNet).IP // hack, because AddIP nullifies struct
			if !*ipv6 && ip.To4() == nil {
				continue
			}
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

	log.Printf("%d routes/ranges loaded", done)

	if namesFile != nil {
		for k, _ := range names {
			k := strings.TrimPrefix(k, "*")
			line := fmt.Sprintf("server=/%s/%s\n", k, DNS)
			_, err := namesFile.WriteString(line)
			if err != nil {
				log.Fatalf("Could not save to %s: %s.", *namesPath, err)
			}
		}
		log.Printf("%d domain names extracted to %s", len(names), namesFile.Name())
	}
}
