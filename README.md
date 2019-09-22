# Loadroutes ![Build Status](https://travis-ci.com/andrewmed/loadroutes.svg?branch=master)![Go Report Card](https://goreportcard.com/badge/github.com/andrewmed/loadroutes)

Simple program to directly parse and load routes from dump file at [https://github.com/zapret-info/z-i]

Loads 900,000 addresses/ranges in less than 20 seconds on my machine. Gives significant (more than ten times) boost compared with loading in cli as in `while read ip; do ip route replace $ip dev $dev; done < preparsed-ip-list`

#### Why
Why not just set up a default gateway for a tunnel? Because it redirects all of your traffic. Just adding custom routes and keeping existing gateway intact let you have the best of the two worlds: direct access and fast ping to local sources, and correct pathway only to the ones necessary. 

#### Test
- Clone repo with `git clone https://github.com/andrewmed/loadroutes.git`
- Run test with `make test`. Test downloads real dump from  [https://github.com/zapret-info/z-i] and loads routes to a new virtual eth (no actual interfaces affected), root required, then ip route stats is printed, veth is deleted at the end

#### Install
Install with `go get github.com/andrewmed/loadroutes/...`

#### Use
Can be run automatically on the network interface raise (here named _tunnel_) with the script (on Debian-based distributions put in `/etc/ppp/ip-up.d`):
```
#!/bin/sh -e
[ $1 = 'tunnel' ] || exit 0
loadroutes -iface $1 -dump /path/to/dump.csv
```
#### Error handling
- On fatal error (io, permissions) exits immediately
- On parsing/loading route error logs to stderr the violating line and continues.

Current format of dump file is processed in full without errors

#### IPv6
Currently silently skips ipv6 addresses

#### Demo
![](demo.gif)

Copyright (c) 2019 Andmed. License: MIT

