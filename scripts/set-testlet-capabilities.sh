#!/bin/bash

# Copyright 2016 Matt Oswalt. Use or modification of this
# source code is governed by the license provided here:
# https://github.com/toddproject/todd/blob/master/LICENSE

# This script configures appropriate capabilities for testlet binaries

set -e
set -u
set -o pipefail

# Ensure GOPATH is set
: ${GOPATH:?"Please ensure GOPATH is set, and run sudo with -E when performing 'make install'"}

# Enable raw socket capabilities on toddping
#
# NOTE - disabled for now, as `setcap` command not present in alpine-based golang docker image
# May want to explore this further, but for now, the image works without this
# setcap cap_net_raw+ep $GOPATH/bin/toddping