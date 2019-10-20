PROJECT_NAME = onkyoctl
MAIN = ./cmd/$(PROJECT_NAME)
BINDIR = ./bin


build:
	mkdir -p $(BINDIR)
	go build -o $(BINDIR)/$(PROJECT_NAME) $(MAIN)

test:
	go test

deps:
	go get gopkg.in/alecthomas/kingpin.v2
	go get github.com/go-ini/ini
