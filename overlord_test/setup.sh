#!/bin/bash


if [ ! -d "$GOPATH/src/github.com/metral/overlord" ]; then
    pushd $GOPATH/src/github.com/metral
    git clone -b $TRAVIS_BRANCH https://github.com/metral/overlord
    git -C overlord checkout -qf $TRAVIS_COMMIT
    popd
fi
# Clone "metral/corekube" just to have access to latest corekube Heat
# template. The tests in "metral/corekube_travis" below use the template to
# deploy a new cluster by overwriting the template's git-command parameter
# to clone overlord and switch to either the PR or commit in the TravisCI
# environment. The overlord code in this build is not used as the code
# is pulled by corekube itself during deployment by way of the git-command
# parameter.
git clone https://github.com/metral/corekube $HOME/corekube
pushd $HOME/corekube
cp corekube-cloudservers.yaml /tmp/
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
cp $GOPATH/src/github.com/metral/overlord/conf.json /tmp/
