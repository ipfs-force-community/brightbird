.PHONY: build
build:
	go build -buildmode=c-shared -o out_mongo.so .

linux:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=gcc go build -buildmode=c-shared -o out_mongo.so .

clean:
	rm -rf *.so