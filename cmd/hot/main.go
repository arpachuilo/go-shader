package main

import (
	. "gogl/window"
)

func main() {
	wnd := NewWindow()
	defer wnd.Close()
	wnd.HotWindow("./bin/plugins/", "plug.so")
}
