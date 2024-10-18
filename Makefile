BINARY_NAME=copilot-proxies
OUT_DIR=out

all: build

build: prepare
	GOOS=linux GOARCH=amd64 go build -o ${OUT_DIR}/${BINARY_NAME}-linux-amd64 main.go
	GOOS=windows GOARCH=amd64 go build -o ${OUT_DIR}/${BINARY_NAME}-windows-amd64.exe main.go
	GOOS=windows GOARCH=amd64 go build -ldflags="-H windowsgui" -o ${OUT_DIR}/${BINARY_NAME}-windows-amd64-gui.exe main.go
	GOOS=darwin GOARCH=amd64 go build -o ${OUT_DIR}/${BINARY_NAME}-darwin-amd64 main.go
	GOOS=linux GOARCH=arm64 go build -o ${OUT_DIR}/${BINARY_NAME}-linux-arm64 main.go
	GOOS=darwin GOARCH=arm64 go build -o ${OUT_DIR}/${BINARY_NAME}-darwin-arm64 main.go

prepare:
	mkdir -p ${OUT_DIR}

clean:
	rm -rf ${OUT_DIR}

.PHONY: all build prepare clean