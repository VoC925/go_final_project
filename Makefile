run: build
	@./bin/app

build:
	@go build -o bin/app service/cmd/main.go