BRANCH = $(shell git rev-parse --abbrev-ref HEAD)
HEAD = $(shell git rev-parse --short HEAD)
DATE = $(shell date +%G%m%d)
VERSION := $(BRANCH)-$(HEAD)

build:
	go build -ldflags "-s -w -X main.version=$(VERSION) -X main.date=$(DATE)"  ./cmd/loadroutes/loadroutes.go

test: build
	go test ./...
	@echo NOW TESTING WITH VIRTUAL ETH ...
	wget -nc https://raw.githubusercontent.com/zapret-info/z-i/master/dump.csv
	sudo whoami
	sudo ip link add veth42s type veth peer name veth42
	sudo ip link set veth42s up
	sudo ./loadroutes -iface veth42s -dump dump.csv
	ip r | wc -l
	sudo ip link del veth42s

install: build
	cp loadroutes /usr/local/bin

clean:
	go clean ./...
	rm -f ./loadroutes
	rm -f ./dump.csv
