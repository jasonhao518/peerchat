help:
	@echo "PeerChat Makefile (Requires Go v1.16+)"
	@echo "'help' - Displays the command usage"
	@echo "'build' - Builds the application into a binary/executable" 
	@echo "'install' - Installs the application"
	@echo "'build-windows' - Builds the application for Windows platforms"
	@echo "'build-darwin' - Builds the application for MacOSX platforms"
	@echo "'build-linux' - Builds the application for Linux platforms"
	@echo "'build-all' - Builds the application for all platforms"

build-x86:
	@CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build  -ldflags="-s -w" -o ./bin/libp2p-proxy-x86_64.dylib -buildmode=c-shared ./cmd/*
	@CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build  -ldflags="-s -w" -o ./bin/p2p-proxy.dll -buildmode=c-shared ./cmd/* 

build-arm64:
	@CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build  -ldflags="-s -w"  -o ./bin/libp2p-proxy-aarch64.dylib -buildmode=c-shared ./cmd/*

build:
	@echo Compiling PeerChat
	@go build -o peerchat ./cmd/*
	@echo Compile Complete. Run './peerchat(.exe)'

install:
	@echo Installing PeerChat
	go install .
	@echo install Complete. Run 'peerchat'.

build-windows:
	@echo Cross Compiling PeerChat for Windows x86
	@GOOS=windows GOARCH=386 go build -o ./bin/peerchat-windows-x32.exe
	@echo Cross Compiling PeerChat for Windows x64
	@GOOS=windows GOARCH=amd64 go build -o ./bin/peerchat-windows-x64.exe

build-darwin:
	@echo Cross Compiling PeerChat for MacOSX x64
	@GOOS=darwin GOARCH=amd64 go build -o ./bin/peerchat-darwin-x64

build-linux:
	@echo Cross Compiling PeerChat for Linux x32
	@GOOS=linux GOARCH=386 go build -o ./bin/peerchat-linux-x32
	@echo Cross Compiling PeerChat for Linux x64
	@GOOS=linux GOARCH=amd64 go build -o ./bin/peerchat-linux-x64
	@echo Cross Compiling PeerChat for Linux Arm32
	@GOOS=linux GOARCH=arm go build -o ./bin/peerchat-linux-arm32
	@echo Cross Compiling PeerChat for Linux Arm64
	@GOOS=linux GOARCH=arm64 go build -o ./bin/peerchat-linux-arm64

build-all: build-windows build-darwin build-linux
	@echo Cross Compiled PeerChat for all platforms
