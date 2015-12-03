docker kill todd-rabbit
docker rm todd-rabbit

docker run -d \
    --hostname todd-rabbit \
    --name todd-rabbit \
    -p 8085:15672 \
    -p 5672:5672 \
    -e RABBITMQ_DEFAULT_USER=guest \
    -e RABBITMQ_DEFAULT_PASS=guest \
    rabbitmq:3-management