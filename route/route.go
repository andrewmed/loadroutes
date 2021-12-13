package route

import (
	"net"
	"log"
	"os"

	"github.com/vishvananda/netlink"
)


func AddIP(iface netlink.Link, ip *net.IPNet) error {
	route := netlink.Route{
		Dst:       ip,
		LinkIndex: iface.Attrs().Index,
	}
	if err := netlink.RouteReplace(&route); err != nil {
		if os.IsPermission(err) {
			log.Fatal("Root privileges are needed to update routing table.")
		}
		return err
	}
	return nil
}
