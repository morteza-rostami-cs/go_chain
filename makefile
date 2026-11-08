
build: 
	go build -o ./bin/go_chain

run: build
	./bin/go_chain

build-cmd:
	go install ./cmd/tmc/...
