#!/bin/bash

go get github.com/tools/godep
echo "========================================"
echo "corekube_travis commit: `git rev-parse --short HEAD`"
echo "========================================"
godep get ./...

# Copy conf.json from overlord in godeps to
# /tmp where overlord's lib expects it - we use lib in the infra_test to
# piece together the etcd api & client port
mkdir -p /tmp/
cp $GOPATH/src/github.com/metral/overlord/conf.json /tmp/
