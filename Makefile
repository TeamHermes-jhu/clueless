lint:
	cd ./backend && golangci-lint -c .golangci.yml run ./...

fmt:
	cd ./backend/ && golangci-lint fmt ./...

mod:
	cd ./backend/ && go mod tidy

test:
	cd ./backend/ && go test ./...

server:
	cd ./backend/ && go build -o server .
