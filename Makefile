VERSION := 2.5
DOCKER_IMAGE := ineva/ipa-server
DOCKER_TARGET := $(DOCKER_IMAGE):$(VERSION)

all:: web

web::
	go run cmd/ipasd/ipasd.go -del

debug::
	go run cmd/ipasd/ipasd.go -d -del

build::
	go build cmd/ipasd/ipasd.go

test::
	go test ./...

image::
	docker build --platform linux/amd64 -t $(DOCKER_TARGET) .

push::
	docker push $(DOCKER_TARGET)