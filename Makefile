#SHELL := /bin/bash

# all: prepare todd-server todd-client todd-agent clean

all: install_deps build

# prepare: install_deps
# 	rm -rf build/
# 	mkdir -p build/
# 	rm -f $(GOPATH)/bin/todd-server
# 	rm -f $(GOPATH)/bin/todd-client
# 	rm -f $(GOPATH)/bin/todd-agent


clean:
	rm -rf build/
	rm -f $(GOPATH)/bin/todd-server
	rm -f $(GOPATH)/bin/todd-client
	rm -f $(GOPATH)/bin/todd-agent
	# add stop containers and kill processes here?

build: clean
	mkdir -p build/
	mkdir -p assets/
	go-bindata -o assets/factcollectors_unpack.go -pkg="assets" facts/collectors/...
	GOPATH=$$(pwd)/Godeps/_workspace:$$GOPATH; cd server && go build -o ../build/todd-server
	GOPATH=$$(pwd)/Godeps/_workspace:$$GOPATH; cd client && go build -o ../build/todd-client
	GOPATH=$$(pwd)/Godeps/_workspace:$$GOPATH; cd agent && go build -o ../build/todd-agent
	docker build -t mierdin/todd -f Dockerfile .

build-no-docker: clean
	mkdir -p build/
	mkdir -p assets/
	go-bindata -o assets/factcollectors_unpack.go -pkg="assets" facts/collectors/...
	GOPATH=$$(pwd)/Godeps/_workspace:$$GOPATH; cd server && go build -o ../build/todd-server
	GOPATH=$$(pwd)/Godeps/_workspace:$$GOPATH; cd client && go build -o ../build/todd-client
	GOPATH=$$(pwd)/Godeps/_workspace:$$GOPATH; cd agent && go build -o ../build/todd-agent

start:
	scripts/start-containers.sh

stop:
	scripts/stop-containers.sh

# clusterclean:
# 	rm -rf test/uploads*

# run: reticulum
# 	./reticulum -config=test/config0.json

fmt:
	go fmt github.com/mierdin/todd/...

install: build-no-docker
	cp -f build/todd-server $(GOPATH)/bin
	cp -f build/todd-client $(GOPATH)/bin
	cp -f build/todd-agent $(GOPATH)/bin
	rm -rf build/

# test: reticulum
# 	go test .

# coverage: reticulum
# 	go test . -coverprofile=coverage.out
# 	go tool cover -html=coverage.out -o coverage.html

install_deps:
	go get github.com/tools/godep
	go get -u github.com/jteeuwen/go-bindata/...	
	GOPATH=$$(godep path) godep restore

update_deps:
	godep save -r ./...
