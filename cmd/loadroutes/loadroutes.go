// Program loadroutes adds routes to a network interface (probably, a vpn tunnel) directly from a dump file
// in the format as in https://github.com/zapret-info/z-i. Convenient to be called in a script upon raising the network interface.
package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"sync"

	"loadroutes/parse"
	"loadroutes/resolver"
	"sort"
	"net"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"

	"loadroutes/router"
)

var (
	version = "unknown"
	date    = "unknown"
)


func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("Loadroutes", version, date)

	iface := flag.String("iface", "", "Network interface name (required)")
	input := flag.String("input", "", "Path to input dump file (see https://github.com/zapret-info/z-i) (required)")
	dns := flag.String("dns", "", "DNS server (if not specified no DNS resolution performed)")
	ip6 := flag.Bool("ip6", false, "Process IPv6 addresses as well (if not specified, IPv4 is used only)")

	flag.Parse()

	if *iface == "" || *input == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	log.Printf("Input file: %s, iface: %s, DNS: %s\n", *input, *iface, *dns)
	if *dns == "" {
		log.Println("No DNS server specified, DNS resolution will not be perfomed")
	}

	file, err := os.Open(*input)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	reader := bufio.NewReader(transform.NewReader(file, charmap.Windows1251.NewDecoder()))

	ips, namesMap := parse.Parse(reader)
	log.Printf("Parsed: ip addresses: %d, dns names: %d\n", len(ips), len(namesMap))
	router, err := route.NewRouter(*iface, *ip6)
	if err != nil {
		log.Fatalf("Initializing router: %s", err)
	}
	router.Add(ips, 100000)

	if *dns == "" {
		return
	}

	names := make(chan string)
	wgNames := sync.WaitGroup{}

	addresses := make(chan []*net.IPNet, 1000000)
	wgAddresses := sync.WaitGroup{}

	resolver.NewResolver(*dns, *ip6, 100).Start(&wgNames, names, addresses)
	router.Start(&wgAddresses, addresses, 100)

	// resolve in order of domain length (on presumption that the shorter domain, the more important it is)
	namesSlice := make([]string, 0, len(namesMap))
	for name := range namesMap {
		namesSlice = append(namesSlice, name)
	}
	sort.Slice(namesSlice, func(i, j int) bool {
		return len(namesSlice[i]) < len(namesSlice[j])
	})
	for _, name := range namesSlice {
		names<-name
	}

	close(names)
	wgNames.Wait()
	close(addresses)
	wgAddresses.Wait()
}
