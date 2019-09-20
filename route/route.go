package route

import (
	"log"
	"net"
	"os"

	"github.com/vishvananda/netlink"
)

func AddIP(link netlink.Link, ip *net.IPNet) error {
	route := netlink.Route{
		Dst:       ip,
		LinkIndex: link.Attrs().Index,
	}
	if err := netlink.RouteReplace(&route); err != nil {
		if os.IsPermission(err) {
			log.Fatal("Root privileges are needed to update routing table.\n")
		}
		return err
	}
	return nil
}
