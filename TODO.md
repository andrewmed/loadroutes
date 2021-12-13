Currently the program is bounded by syscalls in _netlink_ library

- consider replacing netlink library with go standard libs for parsing and socket calls (`src/net/ip.go:679` and friends)
- parallel DNS queries
