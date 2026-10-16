FROM --platform=$BUILDPLATFORM golang:1.22.6 AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src/

COPY go.mod go.sum ./

RUN --mount=type=cache,id=gomod,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,id=gomod,target=/go/pkg/mod \
    --mount=type=cache,id=gobuild,target=/root/.cache/go-build \
    CGO_ENABLED=1 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -o /src/ipasd cmd/ipasd/ipasd.go

# runtime
FROM alpine:3.21

WORKDIR /app

COPY --from=builder /src/ipasd /app/ipasd
COPY docker-entrypoint.sh /docker-entrypoint.sh

RUN chmod +x /docker-entrypoint.sh

ENTRYPOINT ["/docker-entrypoint.sh"]
