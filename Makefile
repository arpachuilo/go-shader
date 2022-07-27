# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

NOW=$(shell date +%s)
BINARY_NAME=bin/opengl_programs
FILES=\
			assets.go\
			cyclic_array.go \
			julia.go\
			live_edit.go\
			mandelbrot.go\
			overlay.go\
			program.go\
			renderer.go\
			smooth.go\
			turtle.go\
			input.go\
			life.go\
			noop.go\
			recorder.go\
			shader.go\
			texture.go\
			pong.go\
			vector.go

all: test build run
build:
	$(GOBUILD) -o $(BINARY_NAME) -v plug.go $(FILES)
test: 
	$(GOTEST) -v ./...
clean: 
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
plug:
	go build -buildmode=plugin -ldflags="-X 'main.BuildDate=${NOW}'" -o bin/plugins/${NOW}.so plug.go $(FILES)
run:
	$(GOBUILD) -o $(BINARY_NAME) -v main.go
	./$(BINARY_NAME)
deps:
	$(GOGET) 
