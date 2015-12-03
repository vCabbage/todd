docker kill todd-rabbit
docker kill todd-server
docker kill todd-agent
docker kill etcd
docker rm todd-rabbit
docker rm todd-server
docker rm todd-agent
docker rm etcd

export HostIP=$(boot2docker ip)

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

docker build -t mierdin/todd .

docker run -d \
    --hostname todd-rabbit \
    --name todd-rabbit \
    -p 8085:15672 \
    -p 5672:5672 \
    -e RABBITMQ_DEFAULT_USER=guest \
    -e RABBITMQ_DEFAULT_PASS=guest \
    rabbitmq:3-management

sleep 5

docker run -d --name todd-server -p 8080:8080 -p 8090:8090 mierdin/todd todd-server
docker run -d --name todd-agent mierdin/todd todd-agent