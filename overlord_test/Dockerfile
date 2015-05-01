FROM google/golang:1.4

RUN apt-get update -y

ADD ./ /gopath/src/github.com/metral/corekube_travis
WORKDIR /gopath/src/github.com/metral/corekube_travis/overlord_test/
RUN ./setup.sh

ENTRYPOINT ["./entrypoint.sh"]
