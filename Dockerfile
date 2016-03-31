FROM golang:1.5
MAINTAINER Matt Oswalt <matt@keepingitclassless.net> (@mierdin)

LABEL version="0.1"

env PATH /go/bin:$PATH

RUN mkdir /etc/todd

RUN mkdir -p /opt/todd/agent/assets/factcollectors
RUN mkdir -p /opt/todd/server/assets/factcollectors
RUN mkdir -p /opt/todd/agent/assets/testlets
RUN mkdir -p /opt/todd/server/assets/testlets

RUN apt-get update \
 && apt-get install -y vim curl iperf git

# Install ToDD
COPY . /go/src/github.com/Mierdin/todd

RUN cd /go/src/github.com/Mierdin/todd && make install

COPY ./etc/agent.cfg /etc/todd/agent.cfg
COPY ./etc/server.cfg /etc/todd/server.cfg
COPY ./etc/agent-dev.cfg /etc/todd/agent-dev.cfg
COPY ./etc/server-dev.cfg /etc/todd/server-dev.cfg