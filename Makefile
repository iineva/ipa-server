
all:: web

web::
	go run cmd/ipasd/ipasd.go

debug::
	go run cmd/ipasd/ipasd.go -d

build::
	go build cmd/ipasd/ipasd.go

test::
	go test ./...
