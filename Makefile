.PHONY: build run test bench vet fmt clean

build:
	go build -o termigocraft .

run:
	go run .

test:
	go test ./...

bench:
	go test ./internal/render/ -bench=. -benchmem -run=^$$

vet:
	go vet ./...

fmt:
	gofmt -l -w .

clean:
	rm -f termigocraft
