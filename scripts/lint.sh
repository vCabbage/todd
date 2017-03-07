# Install dependencies
go get -u github.com/alecthomas/gometalinter

# Lint checks, skipping vendor and assets directory
gometalinter --install --disable-all --enable=golint --skip=assets --vendor ./...
