#!/bin/bash

docker rm -f overlord_test
DIR="$(cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

result=`docker build --rm -t overlord_test $DIR/.`
echo "$result"

echo ""
echo "=========================================================="
echo ""

build_status=`echo $result | grep "Successfully built"`

if [ "$build_status" ] ; then
    cp ~/travis /tmp/travis
    pushd $GOPATH/src/github.com/metral/overlord;
    COMMIT=`git rev-parse --short HEAD`
    echo $COMMIT;
    sed -i "s#TRAVIS_COMMIT.*#TRAVIS_COMMIT=$COMMIT#g" /tmp/travis
    popd

    docker run --env-file=/tmp/travis -i --entrypoint /bin/bash -t --rm -v ~/go/src/github.com/metral/goheat:/gopath/src/github.com/metral/goheat -v ~/go/src/github.com/metral/goutils:/gopath/src/github.com/metral/goutils -v ~/go/src/github.com/metral/corekube_travis:/gopath/src/github.com/metral/corekube_travis overlord_test -s
fi
