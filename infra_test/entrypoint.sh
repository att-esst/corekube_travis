#!/bin/bash

source /tmp/env
/gopath/bin/infra_test \
    --authUrl=$TRAVIS_OS_AUTH_URL \
    --keypair=$TRAVIS_OS_KEYPAIR \
    --password=$TRAVIS_OS_PASSWORD \
    --username=$TRAVIS_OS_USERNAME \
    --tenantId=$TRAVIS_OS_TENANT_ID \
    --templateFile="/tmp/template.yaml"
