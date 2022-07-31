# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

NOW=$(shell date +%s)
BINARY_NAME=bin/opengl_programs
PLUG_FILES=\
			plug.go\
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
			vector.go\
			object.go

WINDOW_FILE=$(shell stat -f window_`go env GOOS`.go || stat -f window_unsupported.go)
MAIN_FILES=\
			main.go\
			$(WINDOW_FILE)


.PHONY: clean deps test

all: test build run
build: $(MAIN_FILES) $(PLUG_FILES)
	$(GOBUILD) -o $(BINARY_NAME) -v $(MAIN_FILES)
	$(GOBUILD) -buildmode=plugin -ldflags="-X 'main.BuildDate=${NOW}'" -o bin/plugins/plug.so $(PLUG_FILES)
plug: $(PLUG_FILES)
	$(GOBUILD) -buildmode=plugin -ldflags="-X 'main.BuildDate=${NOW}'" -o bin/plugins/${NOW}.so $(PLUG_FILES)
run: build
	./$(BINARY_NAME)
test: 
	$(GOTEST) -v ./...
clean: 
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
deps:
	$(GOGET) 
