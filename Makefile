.PHONY : test clean

build:
	go build -o ./bin/jpeg-parse ./cmd/jpeg-parse

test:
	go test ./cmd/jpeg-parse ./internal/jpeg

run:
	./bin/jpeg-parse ./test/data/minneapolis.jpg

clean:
	rm bin/jpeg-parse
