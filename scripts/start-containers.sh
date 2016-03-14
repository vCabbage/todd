# Copyright 2016 Matt Oswalt. Use or modification of this
# source code is governed by the license provided here:
# https://github.com/Mierdin/todd/blob/master/LICENSE

docker-machine start docker-dev
eval $(docker-machine env docker-dev)

docker kill todd-agent-1
docker kill todd-agent-2
docker kill todd-agent-3
docker kill todd-rabbit
docker kill etcd
docker kill influx
docker kill grafana
docker rm todd-rabbit
docker rm etcd
docker rm influx
docker rm grafana
docker rm todd-agent-1
docker rm todd-agent-2
docker rm todd-agent-3

docker pull mierdin/todd

export HostIP=10.128.0.2
#export HostIP=$(docker-machine ip docker-dev)

docker run -d -v /usr/share/ca-certificates/:/etc/ssl/certs -p 4001:4001 -p 2380:2380 -p 2379:2379 \
 --name etcd quay.io/coreos/etcd:v2.0.8 \
 -name etcd0 \
 -advertise-client-urls http://${HostIP}:2379,http://${HostIP}:4001 \
 -listen-client-urls http://0.0.0.0:2379,http://0.0.0.0:4001 \
 -initial-advertise-peer-urls http://${HostIP}:2380 \
 -listen-peer-urls http://0.0.0.0:2380 \
 -initial-cluster-token etcd-cluster-1 \
 -initial-cluster etcd0=http://${HostIP}:2380 \
 -initial-cluster-state new

# I have a rabbitmq server at home, but in case I'm working on my laptop only, spin this up too:
docker run -d \
    --hostname todd-rabbit \
    --name todd-rabbit \
    -p 8085:15672 \
    -p 5672:5672 \
    -e RABBITMQ_DEFAULT_USER=guest \
    -e RABBITMQ_DEFAULT_PASS=guest \
    rabbitmq:3-management

docker run -d --volume=/var/influxdb:/data --name influx -p 8083:8083 -p 8086:8086 tutum/influxdb:0.9
docker run -d --volume=/var/lib/grafana:/var/lib/grafana --name grafana -p 3000:3000 grafana/grafana

sleep 5

docker run -d --name="todd-agent-1" mierdin/todd todd-agent
docker run -d --name="todd-agent-2" mierdin/todd todd-agent
docker run -d --name="todd-agent-3" mierdin/todd todd-agent

#docker run -d --name todd -p 8090:8090 -p 8080:8080 -p 8081:8081 mierdin/todd todd-server
