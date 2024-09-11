# builder
FROM golang:1.22.6 AS builder
WORKDIR /src/
COPY go.mod /src/
COPY go.sum /src/
RUN --mount=type=cache,id=gomod,target=/go/pkg/mod \
  --mount=type=cache,id=gobuild,target=/root/.cache/go-build \
  go mod download && \
  go mod tidy
COPY . /src/
RUN --mount=type=cache,id=gomod,target=/go/pkg/mod \
  --mount=type=cache,id=gobuild,target=/root/.cache/go-build \
  CGO_ENABLED=1 go build -ldflags '-linkmode "external" --extldflags "-static"' cmd/ipasd/ipasd.go

# runtime
FROM ineva/alpine:3.10.3
LABEL maintainer="Steven <s@ineva.cn>"
WORKDIR /app
COPY --from=builder /src/ipasd /app
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh
ENTRYPOINT /docker-entrypoint.sh
