# Install dependencies
go get -u github.com/golang/lint/golint
go get -u github.com/alecthomas/gometalinter

# Lint checks, skipping vendor and assets directory
gometalinter --disable-all --enable=golint --skip=assets --vendor ./...
