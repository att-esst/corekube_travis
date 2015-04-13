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
    docker run --name overlord_test -d -v /tmp/conf.json:/conf.json overlord_test
fi
