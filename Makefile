all: build

build:
	mkdir -p bin
	go build -o bin/sdfs main.go

run:
	go run main.go $(ARGS)

clean:
	rm -rf bin
	rm -rf logs/*/*.log
	rm -rf blocks/*/*
