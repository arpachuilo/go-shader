# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

NOW=$(shell date +%s)
BINARY_NAME=bin/opengl_programs
PLUG_FILES=cmd/$(program)/main.go

WINDOW_FILE=$(shell stat -f window_`go env GOOS`.go || stat -f window_unsupported.go)
HOT_FILES=cmd/hot/main.go


.PHONY: clean deps test

all: test build run
build: $(HOT_FILES) $(PLUG_FILES)
	$(GOBUILD) -o $(BINARY_NAME) -v $(HOT_FILES)
	$(GOBUILD) -buildmode=plugin -ldflags="-X 'main.BuildDate=${NOW}'" -o bin/plugins/plug.so $(PLUG_FILES)
plug: $(PLUG_FILES)
	$(GOBUILD) -buildmode=plugin -ldflags="-X 'main.BuildDate=${NOW}'" -o bin/plugins/${NOW}.so $(PLUG_FILES)
run: build
	./$(BINARY_NAME)
test: 
	$(GOTEST) -v ./...
clean: 
	$(GOCLEAN)
	rm bin/plugins/*
	rm -f $(BINARY_NAME)
deps:
	$(GOGET) 
