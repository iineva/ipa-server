
all:: web

web::
	go run cmd/ipa-server/ipa-server.go

debug::
	go run cmd/ipa-server/ipa-server.go -d