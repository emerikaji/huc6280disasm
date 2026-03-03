// Package system keeps track of the variables during disassembly.
package system

import (
	"fmt"
)

// TODO change "Valid" to "known"

type ValidByte struct {
	Value byte
	Valid bool
}

type Environment struct {
	a         ValidByte
	pages     [8]ValidByte
	carryflag ValidByte
	stack     []ValidByte
}

func NewEnv() *Environment {
	e := new(Environment)
	// TODO verify all this
	e.a.Valid = false
	for i := 0; i < 7; i++ {
		e.pages[i].Valid = false
	}
	e.pages[7].Value = 0x00 // we need the vectors at startup
	e.pages[7].Valid = true
	e.carryflag.Value = 0
	e.carryflag.Valid = false
	return e
}

func (e *Environment) Physical(logical uint16) (uint32, error) {
	page := (logical & 0xE000) >> 13
	if !e.pages[page].Valid {
		return 0, fmt.Errorf("attempt to get physical address of logical $%X, but the page has not yet been initialized", logical)
	}
	physical := uint32(logical) &^ 0xE000
	physical |= 0x2000 * uint32(e.pages[page].Value)
	return physical, nil
}

func (e *Environment) Valid() bool {
	return e.a.Valid
}

func (e *Environment) Invalidate() {
	e.a.Valid = false
	e.carryflag.Valid = false
}

func (e *Environment) Push(v ValidByte) {
	e.stack = append(e.stack, v)
}

func (e *Environment) Pop() ValidByte {
	if len(e.stack) == 0 {
		return ValidByte{0, false} // TODO correct?
	}
	t := e.stack[len(e.stack)-1]
	e.stack = e.stack[:len(e.stack)-1]
	return ValidByte{t.Value, t.Valid}
}

func (e *Environment) Seta(v ValidByte) {
	e.a = v
}

func (e *Environment) Inca() {
	e.a.Value++
}

func (e *Environment) Deca() {
	e.a.Value--
}

func (e *Environment) Pusha() {
	e.Push(e.a)
}

func (e *Environment) PushInvalid() {
	e.Push(ValidByte{e.a.Value, false}) // Value of a irrelevant
}

func (e *Environment) Popa() {
	e.a = e.Pop()
}

func (e *Environment) PageToa(curpage int) {
	e.a = e.pages[curpage]
}

func (e *Environment) AToPage(curpage int) {
	e.pages[curpage] = e.a
}

func (e *Environment) Save() *Environment {
	destEnv := new(Environment)
	destEnv.a = e.a
	destEnv.pages = e.pages
	destEnv.carryflag = e.carryflag
	destEnv.stack = make([]ValidByte, len(e.stack))
	copy(destEnv.stack, e.stack)
	return destEnv
}

func (e *Environment) Restore(existEnv *Environment) {
	e.a = existEnv.a
	e.pages = existEnv.pages
	e.carryflag = existEnv.carryflag
	e.stack = make([]ValidByte, len(existEnv.stack))
	copy(e.stack, existEnv.stack)
}
