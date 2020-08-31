all: prep build run

prep:
	@mkdir -p bin/ output/

build:
	@go build -o bin/faced ./cmd/faced/...

run:
	@bin/faced
