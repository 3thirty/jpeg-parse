build:
	go build -o ./bin/jpeg-parse ./cmd/jpeg-parse

test:
	go test ./cmd/jpeg-parse

clean:
	rm bin/jpeg-parse
