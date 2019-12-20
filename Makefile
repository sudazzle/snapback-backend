#VAR := value
#$(VAR)

# get all the dependencies
# go get -d ./...

run:
	go run main.go

build:
# 	echo "Compiling for every OS and Platform"
# 	GOOS=freebsd GOARCH=386 go build -o bin/main-freebsd-386 main.go
# 	GOOS=windows GORCH=386 go build -o bin/main-windows-386 main.go
	GOOS=linux GOARCH=amd64 go build -o bin/snapback-linux-amdd64 main.go
	cp .env bin/.env
	cp -r  ui bin
	cp snapback.service bin/snapback.service



# grep -v is inverting line except /vender
# /... recursive
# PKGS := $(shell go list ./... | grep -v /vendor)

# .PHONY: test
# test:
# 	go test $(PKGS)

zip:
# -z : Compress archive using gzip program in Linux or Unix
# -c : Create archive on Linux
# -v : Verbose i.e display progress while creating archive
# -f : Archive File name
	
	tar -zcvf prodbuild/snapback-linux-amd64.tar.gz bin/
