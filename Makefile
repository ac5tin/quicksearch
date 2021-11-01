test: 
	go test ./... -v -cover

build:
	go build -o bin/app

run:
	go run main.go

clean:
	rm -rf ./bin