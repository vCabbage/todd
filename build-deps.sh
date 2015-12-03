#!/bin/bash
set -euo pipefail # Unofficial bash strict mode
IFS=$'\n\t'

godep save github.com/mierdin/todd/agent github.com/mierdin/todd/client github.com/mierdin/todd/common github.com/mierdin/todd/registry github.com/mierdin/todd/server