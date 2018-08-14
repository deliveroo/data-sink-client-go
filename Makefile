all: install test doc

doc/install:
	@go get github.com/davecheney/godoc2md
doc:
	@godoc2md github.com/deliveroo/data-sink-client-go | \
		sed 's|/src/target/|./|g' > godoc.md

install:
	@go install ./...

test:
	@go test ./...

.PHONY: doc
