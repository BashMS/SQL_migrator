FROM golang:1.22.3 as vendor

RUN set -ex \
        && apt-get update \
        && apt-get install -y --no-install-recommends --no-install-suggests --allow-unauthenticated \
    && go get -u -t \
        github.com/vektra/mockery/.../ \
    && rm -rf /var/lib/apt/lists/* /var/cache/*

WORKDIR /src

COPY go.mod go.sum ./
COPY Makefile ./
RUN go mod download

FROM vendor as app
COPY main.go main.go
COPY ./cmd ./cmd
COPY ./generate ./generate
COPY ./templates ./templates
COPY ./pkg ./pkg
COPY ./internal ./internal

FROM app as int-test

ENV GOOS linux
ENV GOARCH amd64
ENV CGO_ENABLED 0

COPY ./config ./config
RUN make build

COPY ./test ./test

CMD ["go", "test", "./test/integration", "-tags", "integration"]
