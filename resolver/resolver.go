package resolver

import(
	"net"
	"context"
)

func Resolve(ctx context.Context, resolver *net.Resolver, name string, ip6 bool) ([]*net.IPNet, error) {
	proto := "ip" 
	if !ip6 {
		proto = "ip4"
	}
	ips, err := resolver.LookupIP(ctx, proto, name)
	if err != nil {
		return nil, err
	}
	var result []*net.IPNet
	for _, ip := range ips {
		result = append(result, &net.IPNet{
			IP: ip,
			Mask: net.IPv4Mask(255,255,255,255),
		})
	}
	return result, nil
}

