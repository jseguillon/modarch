BINARY := bin/kgraph

.PHONY: build test run-render run-animate clean

build:
	go build -o $(BINARY) ./cmd/kgraph

test:
	go test ./...

run-render: build
	./$(BINARY) render -i examples/testdata/mock.yaml -o out/graph.dot -png out/graph.png

run-animate: build
	./$(BINARY) animate -i examples/testdata/mock.yaml -name example -out out/frames -apng out/example.png

clean:
	rm -rf bin out
