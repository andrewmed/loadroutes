// Program loadroutes adds routes to a network interface (probably, a vpn tunnel) directly from a dump file
// in the format as in https://github.com/zapret-info/z-i. Convenient to be called in a script upon raising the network interface.
package main

import (
	"bufio"
	"flag"
	"log"
	"net"
	"os"

	"context"
	"loadroutes/parse"
	"loadroutes/resolver"
	"loadroutes/route"
	"sort"
	"time"

	"github.com/vishvananda/netlink"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

var (
	version = "unknown"
	date    = "unknown"
)

const DNS_TIMEOUT_SEC = 1

var done int

func add(iface netlink.Link, ips []*net.IPNet, ip6 bool, logRadix int) {
	for _, ipNet := range ips {
		ip := (*ipNet).IP // hack for error reporting because AddIP nullifies struct after call
		if !ip6 && ip.To4() == nil {
			continue
		}
		if err := route.AddIP(iface, ipNet); err != nil {
			log.Printf("Adding %v: %s", ip, err)
			continue
		}

		done++
		if done%logRadix == 0 {
			log.Println(done)
		}
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("Loadroutes", version, date)

	ifaceName := flag.String("iface", "", "Network interface name (required)")
	input := flag.String("input", "", "Path to input dump file (see https://github.com/zapret-info/z-i) (required)")
	dnsName := flag.String("dns", "", "DNS server (if not specified no DNS resolution performed)")
	ip6 := flag.Bool("ip6", false, "Process IPv6 addresses as well (if not specified, IPv4 is used only)")

	flag.Parse()

	if *ifaceName == "" || *input == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	log.Printf("Input file: %s, iface: %s, DNS: %s\n", *input, *ifaceName, *dnsName)
	if *dnsName == "" {
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

	iface, err := netlink.LinkByName(*ifaceName)
	if err != nil {
		log.Fatalf("No such network interface: %s", *ifaceName)
	}

	add(iface, ips, *ip6, 100000)

	if *dnsName == "" {
		return
	}

	dnsResolver := net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, network, *dnsName+":53")
		},
	}

	// resolve in order of domain length (on presumption that the shorter domain, the more important it is)
	namesSlice := make([]string, 0, len(namesMap))
	for name := range namesMap {
		namesSlice = append(namesSlice, name)
	}
	sort.Slice(namesSlice, func(i, j int) bool {
		return len(namesSlice[i]) < len(namesSlice[j])
	})

	var resolutionErrs int
	for _, name := range namesSlice {
		ctx, cancelFn := context.WithTimeout(context.Background(), time.Second*DNS_TIMEOUT_SEC)
		defer cancelFn()
		ips, err := resolver.Resolve(ctx, &dnsResolver, name, *ip6)
		if err != nil {
			resolutionErrs++
			if resolutionErrs%100 == 0 {
				log.Printf("DNS resolution errors so far: %d, last error: %s", resolutionErrs, err)
			}
			continue
		}
		add(iface, ips, *ip6, 100)
	}
}
