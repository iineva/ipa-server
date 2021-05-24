# builder
FROM golang:1.16 AS builder
COPY go.mod /src/
COPY go.sum /src/
RUN cd /src && go mod download
COPY . /src/
RUN cd /src && CGO_ENABLED=0 go build cmd/ipasd/ipasd.go

# runtime
FROM ineva/alpine:3.10.3
LABEL maintainer="Steven <s@ineva.cn>"
WORKDIR /app
COPY --from=builder /src/ipasd /app
ENTRYPOINT /app/ipasd
