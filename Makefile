.PHONY: dc run test lint

dc:
	docker-compose up --remove-orphans --build

run:
	go build -o A2-modules-graph cmd/A2-modules-graph/main.go ./A2-modules-graph 8080

test:
	go test -race ./...

lint:
	golangci-lint run
