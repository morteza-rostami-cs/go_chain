
build: 
	go build -o ./bin/go_chain

run: build
	./bin/go_chain

build-cmd:
	go install ./cmd/tmc/...

run-node:
	tmc run --datadir=./data

tree:
	tree -a -I "_archives|_notes|.git"