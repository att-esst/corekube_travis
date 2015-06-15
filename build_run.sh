#!/bin/bash

RUN_TEST=$1
TEMPLATE=$2

docker rm -f $RUN_TEST
DIR="$(cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

result=`docker build --no-cache --rm -t $RUN_TEST -f $DIR/$RUN_TEST/Dockerfile .`
echo "$result"

echo ""
echo "=========================================================="
echo ""

build_status=`echo $result | grep "Successfully built"`

if [ "$build_status" ] ; then
    cp ~/travis /tmp/travis
    # til popd - for overlord_test - not used in infra_test
    pushd $GOPATH/src/github.com/metral/overlord;
    COMMIT=`git rev-parse --short HEAD`
    echo $COMMIT;
    sed -i "s#TRAVIS_COMMIT.*#TRAVIS_COMMIT=$COMMIT#g" /tmp/travis
    popd

    if [ "$TEMPLATE" ] ; then
        docker run --name $RUN_TEST -d --env-file=/tmp/travis -v $TEMPLATE:/tmp/corekube-cloudservers.yaml $RUN_TEST
    else
        docker run --name $RUN_TEST -d --env-file=/tmp/travis $RUN_TEST
    fi
fi
