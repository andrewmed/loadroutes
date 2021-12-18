package route

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"github.com/vishvananda/netlink"
)

type Router struct {
	link netlink.Link
	processed int
	ip6 bool
}

func NewRouter(iface string, ip6 bool) (*Router, error) {
	link, err := netlink.LinkByName(iface)
	if err != nil {
		return nil, fmt.Errorf("no such network interface: %s", iface)
	} 
	router := Router {
		link:link,
		ip6:ip6,
	}
	return &router, nil
}

func (self *Router) Add(ips []*net.IPNet, logRadix int) error {
	for _, ipNet := range ips {
		ip := (*ipNet).IP // hack for error reporting because AddIP nullifies struct after call
		if !self.ip6 && ip.To4() == nil {
			continue
		}
		route := netlink.Route{
			Dst:       ipNet,
			LinkIndex: self.link.Attrs().Index,
		}
		if err := netlink.RouteReplace(&route); err != nil {
			if os.IsPermission(err) {
				return fmt.Errorf("no permissions")
			}
			log.Printf("Adding %v: %s", ip, err)
			continue
		}
		self.processed++
		if self.processed%logRadix == 0 {
			log.Println(self.processed)
		}
	}
	return nil
}

func (self *Router) Start(wg *sync.WaitGroup, addresses chan []*net.IPNet, logRadix int) {
	wg.Add(1)
	go func() {
		for address := range addresses {
			self.Add(address, logRadix)
		}
		wg.Done()
	}()
}
