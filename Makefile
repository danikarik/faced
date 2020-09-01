all: prep build

prep:
	@mkdir -p bin/ output/

build:
	@go build -o bin/compare ./cmd/compare/...
	@go build -o bin/recognize ./cmd/recognize/...

recognize:
	@bin/recognize

compare:
	@bin/compare
