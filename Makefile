#SHELL := /bin/bash

all: install_deps compile

clean:
	rm -rf build/
	rm -f $(GOPATH)/bin/todd-server
	rm -f $(GOPATH)/bin/todd
	rm -f $(GOPATH)/bin/todd-agent

build: compile
	docker build -t mierdin/todd -f Dockerfile .
	rm -rf build/

compile: install_deps clean
	mkdir -p build/
	mkdir -p assets/
	$(GOPATH)/bin/go-bindata -o assets/assets_unpack.go -pkg="assets" agent/...
	cd server && $(GOPATH)/bin/godep go build -o ../build/todd-server
	cd client && $(GOPATH)/bin/godep go build -o ../build/todd
	cd agent && $(GOPATH)/bin/godep go build -o ../build/todd-agent

fmt:
	go fmt github.com/mierdin/todd/...

configureenv:
	mkdir -p /etc/todd
	cp -f etc/* /etc/todd/
	mkdir -p /opt/todd/agent/assets/factcollectors
	mkdir -p /opt/todd/server/assets/factcollectors
	mkdir -p /opt/todd/agent/assets/testlets
	mkdir -p /opt/todd/server/assets/testlets
	chmod -R 777 /opt/todd

install: configureenv
	cp -f build/todd-server /usr/bin
	cp -f build/todd /usr/bin
	cp -f build/todd-agent /usr/bin
	rm -rf build/

test: 
	godep go test ./... -cover

install_deps:
	go get github.com/tools/godep
	go get -u github.com/jteeuwen/go-bindata/...

update_deps:
	godep save ./...
