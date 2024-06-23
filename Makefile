.PHONY: dc run test lint build clean

dc:
	docker-compose up --remove-orphans --build

build:
	mkdir -p output	
	go build -o output/a2-modules-graph cmd/A2-modules-graph/main.go

run: clean build
	cd output && ./a2-modules-graph
	# Generate SVG from the dot file
	dot -Tsvg output/modules_graph.dot -o output/modules_graph.svg

	# Output the result
	@echo "SVG file generated: output/modules_graph.svg"

test:
	go test -race ./...

lint:
	golangci-lint run

clean:
	rm -rf output
