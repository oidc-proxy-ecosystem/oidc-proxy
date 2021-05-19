FROM scratch as client

ENV DOCKER_CLIENT_VERSION=19.03.8
ENV DOCKER_API_VERSION=1.40
ADD https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKER_CLIENT_VERSION}.tgz .

FROM golang:1.16-buster
ENV TZ=Asia/Tokyo
ARG ARG_DOCKER_CLIENT_VERSION=19.03.8

ENV DOCKER_CLIENT_VERSION=${ARG_DOCKER_CLIENT_VERSION}
ENV DOCKER_API_VERSION=1.40
COPY --from=client docker-${DOCKER_CLIENT_VERSION}.tgz .

RUN apt-get update \
    && apt-get install -y --no-install-recommends tzdata protobuf-compiler \
    && cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime \
    && echo "Asia/Tokyo" >  /etc/timezone \
    && tar xzvf docker-${DOCKER_CLIENT_VERSION}.tgz \
    && mv docker/* /usr/bin/ \
    && rm -rf docker-${DOCKER_CLIENT_VERSION}.tgz \
    && rm -rf /var/lib/apt/lists/* \
    && go get google.golang.org/protobuf/cmd/protoc-gen-go \
    google.golang.org/grpc/cmd/protoc-gen-go-grpc \
    github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc

WORKDIR /src/

CMD [ "sh" ]