FROM google/golang:stable

RUN apt-get update -y
RUN apt-get install net-tools -y
RUN go get github.com/tools/godep

WORKDIR /gopath/src/github.com/metral/overlord_test
ADD . /gopath/src/github.com/metral/overlord_test
RUN godep get ./...

CMD []
ENTRYPOINT ["/gopath/bin/overlord_test"]
