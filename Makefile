BINARY_NAME=smv
BINARY_VERSION=v1.0.0

run:
	go run main.go
 
build:
	go build -o ${BINARY_NAME} main.go

buildall: buildmac buildmacarm buildlinux #buildwindows

buildmac:
	GOOS=darwin GOARCH=amd64 go build -o bin/${BINARY_NAME}-${BINARY_VERSION}-macos-amd64/${BINARY_NAME}

buildmacarm:
	GOOS=darwin GOARCH=arm64 go build -o bin/${BINARY_NAME}-${BINARY_VERSION}-macos-arm64/${BINARY_NAME}

buildlinux:
	GOOS=linux GOARCH=amd64 go build -o bin/${BINARY_NAME}-${BINARY_VERSION}-linux-amd64/${BINARY_NAME}

buildwindows:
	GOOS=windows GOARCH=amd64 go build -o bin/${BINARY_NAME}-${BINARY_VERSION}-windows-amd64/${BINARY_NAME}

format:
	go fmt
 
clean:
	go clean
	rm -r bin/
	rm ${BINARY_NAME}