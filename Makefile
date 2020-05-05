SHELL = /bin/bash
TARGETS = redminesync
PKGNAME = redminesync

all: $(TARGETS)

$(TARGETS): %: cmd/%/main.go
	go get ./...
	go build -o $@ $<

clean:
	rm -f $(TARGETS)
