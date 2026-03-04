.PHONY: build test install clean

build:
	go build -o pacto-plugin-schema-infer ./cmd

test:
	go test ./... -v

install: build
	cp pacto-plugin-schema-infer /usr/local/bin/

clean:
	rm -f pacto-plugin-schema-infer
