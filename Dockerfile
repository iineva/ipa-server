# builder
FROM golang:1.16 AS builder
COPY go.mod /src/
COPY go.sum /src/
RUN cd /src && go mod download
COPY . /src/

# install lzfse
RUN git clone https://github.com/lzfse/lzfse && cd lzfse && make install INSTALL_PREFIX=/usr/local
RUN cd /src && go build cmd/ipasd/ipasd.go

# runtime
FROM ineva/alpine:3.10.3
LABEL maintainer="Steven <s@ineva.cn>"
WORKDIR /app
COPY --from=builder /src/ipasd /app
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh
ENTRYPOINT /docker-entrypoint.sh
