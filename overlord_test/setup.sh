#!/bin/bash

# Clone "metral/corekube" just to have access to latest corekube-heat.yaml Heat
# template. The tests in "metral/corekube_travis" below use the template to
# deploy a new cluster by overwriting the template's git-command parameter
# to clone overlord and switch to either the PR or commit in the TravisCI
# environment. The overlord code in this build is not used as the code
# is pulled by corekube itself during deployment by way of the git-command
# parameter.
git clone https://github.com/metral/corekube $HOME/corekube
pushd $HOME/corekube
cp corekube-heat.yaml /tmp/template.yaml
echo "========================================"
echo "corekube commit: `git rev-parse --short HEAD`"
echo "========================================"
popd

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

