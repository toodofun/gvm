FROM golang:1.24.1-alpine3.21 AS builder
WORKDIR /build

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk add --no-cache ca-certificates gcc libtool make musl-dev protoc git bash && \
    go env -w GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make all

FROM scratch
COPY --from=builder /build/bin /