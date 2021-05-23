
all:: web

web::
	go run cmd/ipa-server/ipa-server.go

debug::
	go run cmd/ipa-server/ipa-server.go -d

build::
	go build cmd/ipa-server/ipa-server.go

test::
	go test ./...
