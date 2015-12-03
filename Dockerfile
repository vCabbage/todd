FROM golang:1.5
MAINTAINER Matt Oswalt <matt@keepingitclassless.net> (@mierdin)

LABEL version="0.1"

# Install godep
RUN go get github.com/tools/godep

RUN mkdir -p /tmp/factcollectors
RUN mkdir -p /tmp/agentfactcollectors

# Upload ToDD source
COPY . /go/src/github.com/mierdin/todd

COPY etc/agent_config.cfg /etc/agent_config.cfg
COPY etc/server_config.cfg /etc/server_config.cfg

# (set an explicit GOARM of 5 for maximum compatibility)
ENV GOARM 5

WORKDIR /go/src/github.com/mierdin/todd

RUN godep restore


# Remove all this - only copy the binaries, and leave the building to the makefile
RUN cd server && go build -o /go/bin/todd-server
RUN cd client && go build -o /go/bin/todd-client
RUN cd agent && go build -o /go/bin/todd-agent