all: dist/proxy

dist/proxy:
	go build -o dist/proxy github.com/SimonRichardson/cmdproxy/cmd/proxy

clean: FORCE
	rm -rf dist

FORCE:
