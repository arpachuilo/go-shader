package engine

type Cleaner struct {
	hooks []func()
}

func NewCleaner() *Cleaner {
	return &Cleaner{
		hooks: make([]func(), 0),
	}
}

func (self *Cleaner) Add(f func()) {
	self.hooks = append(self.hooks, f)
}

func (self Cleaner) Run() {
	for _, f := range self.hooks {
		f()
	}
}

// Maybe not do this? Risky, objects will have to remove hooks
// type Key struct {
// 	hooks []glfw.KeyCallback
// }
//
// func NewKey() *Key {
// 	return &Key{
// 		hooks: make([]glfw.KeyCallback, 0),
// 	}
// }
//
// func (self *Key) Add(f glfw.KeyCallback) {
// 	self.hooks = append(self.hooks, f)
// }
//
// func (self Key) Run(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
// 	for _, f := range self.hooks {
// 		f(w, key, scancode, action, mods)
// 	}
// }
//
// type Resize struct {
// 	hooks []glfw.SizeCallback
// }
//
// func NewResize() *Resize {
// 	return &Resize{
// 		hooks: make([]glfw.SizeCallback, 0),
// 	}
// }
//
// func (self *Resize) Add(f glfw.SizeCallback) {
// 	self.hooks = append(self.hooks, f)
// }
//
// func (self Resize) Run(w *glfw.Window, width, height int) {
// 	for _, f := range self.hooks {
// 		f(w, width, height)
// 	}
// }
