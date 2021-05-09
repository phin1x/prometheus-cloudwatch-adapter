SHELL := /bin/bash

clean:
	rm -rf build

build: clean
	mkdir build
	go build -o build/prometheus-cloudwatch-adapter main.go

