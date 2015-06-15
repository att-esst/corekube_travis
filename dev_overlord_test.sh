#!/bin/bash

rm -rf /gopath/bin/overlord_test
pushd overlord_test
go build
./overlord_test --authUrl=$TRAVIS_OS_AUTH_URL --keypair=$TRAVIS_OS_KEYPAIR --password=$TRAVIS_OS_PASSWORD --username=$TRAVIS_OS_USERNAME --tenantId=$TRAVIS_OS_TENANT_ID --templateFile="/tmp/corekube-cloudservers.yaml"
popd
