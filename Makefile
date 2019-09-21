PROJECT_NAME = onkyoctl
MAIN = ./cmd/$(PROJECT_NAME)
BINDIR = ./bin


build:
	mkdir -p $(BINDIR)
	go build -o $(BINDIR)/$(PROJECT_NAME) $(MAIN)

test:
	go test -o $(BINDIR)/$(PROJECT_NAME).test ./$(PROJECT_NAME)
