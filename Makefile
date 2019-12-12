NAME    = onkyoctl
MAIN    = ./cmd/$(NAME)
BINDIR  = ./bin
ARMDIR	= $(BINDIR)/linux/arm

build:
	mkdir -p $(BINDIR)
	go build -o $(BINDIR)/$(NAME) $(MAIN)

arm:
	mkdir -p $(ARMDIR)
	env GOOS=linux GOARCH=arm go build -o $(ARMDIR)/$(NAME) $(MAIN)

test:
	go test

deps:
	go get gopkg.in/alecthomas/kingpin.v2
	go get github.com/go-ini/ini
