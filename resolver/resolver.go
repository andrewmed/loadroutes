package resolver

import(
	"net"
	"context"
	"sync"
	"sync/atomic"
	"time"
	"log"
)

const TIMEOUT_SEC int = 5
const PARALLELISM int = 20

type Resolver struct {
	r net.Resolver
	ip6 bool
	logRadix int
	errs int64
}

func NewResolver(server string, ip6 bool, logRadix int) *Resolver {
	return &Resolver {
		r: net.Resolver {
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{}
				return d.DialContext(ctx, network, server+":53")
			},
		},
		ip6: ip6,
		logRadix: logRadix,
	}
}

func (self Resolver) Resolve(ctx context.Context, name string) ([]*net.IPNet, error) {
	proto := "ip" 
	if self.ip6 {
		proto = "ip4"
	}
	ips, err := self.r.LookupIP(ctx, proto, name)
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

func (self Resolver) Start(wg *sync.WaitGroup, names chan string, addresses chan []*net.IPNet) {
	for i := 0; i < PARALLELISM; i++ {
		wg.Add(1)	
		go func() {
			for name := range names {
				ctx, cancelFn := context.WithTimeout(context.Background(), time.Second*time.Duration(TIMEOUT_SEC))
				defer cancelFn()
				ips, err := self.Resolve(ctx, name)
				if err != nil {
					atomic.AddInt64(&self.errs, 1)
					errNo := atomic.LoadInt64(&self.errs)
					if errNo%int64(self.logRadix) == 0 {
						log.Printf("DNS resolution errors so far: %d, last error: %s", errNo, err)
					}
					continue
				}
				addresses<-ips
			}
			wg.Done()
		}()
	}
}

