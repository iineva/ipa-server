
all:: web

web::
	go run cmd/ipasd/ipasd.go -del

debug::
	go run cmd/ipasd/ipasd.go -d -del

build::
	go build cmd/ipasd/ipasd.go

test::
	go test ./...
