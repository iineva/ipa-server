# builder
FROM golang:1.22.6 AS builder

WORKDIR /src/

COPY go.mod go.sum ./

RUN --mount=type=cache,id=gomod,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,id=gomod,target=/go/pkg/mod \
    --mount=type=cache,id=gobuild,target=/root/.cache/go-build \
    CGO_ENABLED=1 go build \
    -ldflags '-linkmode "external" --extldflags "-static"' \
    -o /src/ipasd cmd/ipasd/ipasd.go

# runtime
FROM alpine:3.10

LABEL maintainer="Steven <s@ineva.cn>"

WORKDIR /app

COPY --from=builder /src/ipasd /app
COPY --chmod=755 docker-entrypoint.sh /docker-entrypoint.sh

ENTRYPOINT ["/docker-entrypoint.sh"]
