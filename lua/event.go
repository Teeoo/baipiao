package Lua

import (
	lua "github.com/yuin/gopher-lua"
)

type EventModule struct{}

func NewEventModule() *EventModule {
	return &EventModule{}
}

func (l *EventModule) Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"Push": l.Send,
	})
	L.Push(mod)
	return 1
}

func (l *EventModule) Send(L *lua.LState) int {
	return 0
}
