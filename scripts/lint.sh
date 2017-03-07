# Install dependencies
go get -u github.com/alecthomas/gometalinter
gometalinter --install > /dev/null

# Lint checks, skipping vendor and assets directory
gometalinter --disable-all --enable=golint --skip=assets --vendor ./...