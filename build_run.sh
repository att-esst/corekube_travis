#!/bin/bash

RUN_TEST=$1
ENV_FILE=$2
TEMPLATE=$3

docker rm -f $RUN_TEST
DIR="$(cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

result=`docker build --no-cache --rm -t $RUN_TEST -f $DIR/$RUN_TEST/Dockerfile .`
echo "$result"

echo ""
echo "=========================================================="
echo ""

build_status=`echo $result | grep "Successfully built"`

if [ "$build_status" ] ; then
    if [ "$TEMPLATE" ] ; then
        docker run --name $RUN_TEST -d -v $ENV_FILE:/tmp/env -v $TEMPLATE:/tmp/corekube-cloudservers.yaml $RUN_TEST
    else
        docker run --name $RUN_TEST -d -v $ENV_FILE:/tmp/env $RUN_TEST
    fi
fi
