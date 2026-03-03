package op

import (
	"fmt"
	"os"

	"github.com/emerikaji/huc6280disasm/system"
)

// paging reference: http://turbo.mindrec.com/tginternals/hw/

// lda #nn
func (r *Runner) ldaImmediate(pos uint32) (disassembled string, newpos uint32, done bool) {
	b := r.w.GetByte(pos)
	pos++
	r.env.Seta(system.ValidByte{Value: b, Valid: true})
	return fmt.Sprintf("lda\t#$%02X", b), pos, false
}

// inc a
func (r *Runner) incAccumulator(pos uint32) (disassembled string, newpos uint32, done bool) {
	// whether or not a is valid does not matter here (if a is invalid the value will not be used anyway)
	r.env.Inca()
	return fmt.Sprintf("inc\ta"), pos, false
}

// dec a
func (r *Runner) decAccumulator(pos uint32) (disassembled string, newpos uint32, done bool) {
	// whether or not a is valid does not matter here (if a is invalid the value will not be used anyway)
	r.env.Deca()
	return fmt.Sprintf("dec\ta"), pos, false
}

// pha
func (r *Runner) phaNoArgs(pos uint32) (disassembled string, newpos uint32, done bool) {
	if r.useStack {
		r.env.Pusha()
	}
	return "pha", pos, false
}

// php, phx, phy
func (r *Runner) push(m string) Code {
	return func(pos uint32) (disassembled string, newpos uint32, done bool) {
		if r.useStack {
			r.env.PushInvalid()
		}
		return fmt.Sprintf("%s", m), pos, false
	}
}

// pla
func (r *Runner) plaNoArgs(pos uint32) (disassembled string, newpos uint32, done bool) {
	if r.useStack {
		r.env.Popa()
	} else {
		r.env.Invalidate()
	}
	return "pla", pos, false
}

// plp, plx, ply
func (r *Runner) pop(m string) Code {
	return func(pos uint32) (disassembled string, newpos uint32, done bool) {
		if r.useStack {
			r.env.Pop()
		}
		return fmt.Sprintf("%s", m), pos, false
	}
}

// sta zz
func (r *Runner) staZeropage(pos uint32) (disassembled string, newpos uint32, done bool) {
	b := r.w.GetByte(pos)
	pos++
	r.comment(pos-2, uint16(b))
	return fmt.Sprintf("sta\t$%02X", b), pos, false
}

// sta zz,x
func (r *Runner) staZeropagex(pos uint32) (disassembled string, newpos uint32, done bool) {
	b := r.w.GetByte(pos)
	pos++
	r.comment(pos-2, uint16(b))
	return fmt.Sprintf("sta\t$%02X,x", b), pos, false
}

// sta (zz)
func (r *Runner) staIndirect(pos uint32) (disassembled string, newpos uint32, done bool) {
	b := r.w.GetByte(pos)
	pos++
	r.comment(pos-2, uint16(b))
	return fmt.Sprintf("sta\t($%02X)", b), pos, false
}

// sta (zz,x)
func (r *Runner) staIndirectx(pos uint32) (disassembled string, newpos uint32, done bool) {
	b := r.w.GetByte(pos)
	pos++
	r.comment(pos-2, uint16(b))
	return fmt.Sprintf("sta\t($%02X,x)", b), pos, false
}

// sta (zz),y
func (r *Runner) staIndirecty(pos uint32) (disassembled string, newpos uint32, done bool) {
	b := r.w.GetByte(pos)
	pos++
	r.comment(pos-2, uint16(b))
	return fmt.Sprintf("sta\t($%02X),y", b), pos, false
}

// sta hhll
func (r *Runner) staAbsolute(pos uint32) (disassembled string, newpos uint32, done bool) {
	w, pos := r.w.GetWord(pos)
	r.comment(pos-3, w)
	return fmt.Sprintf("sta\t$%04X", w), pos, false
}

// sta hhll,x
func (r *Runner) staAbsolutex(pos uint32) (disassembled string, newpos uint32, done bool) {
	w, pos := r.w.GetWord(pos)
	r.comment(pos-3, w)
	return fmt.Sprintf("sta\t$%04X,x", w), pos, false
}

// sta hhll,y
func (r *Runner) staAbsolutey(pos uint32) (disassembled string, newpos uint32, done bool) {
	w, pos := r.w.GetWord(pos)
	r.comment(pos-3, w)
	return fmt.Sprintf("sta\t$%04X,y", w), pos, false
}

// tam #nn,...
func (r *Runner) tamPageRegs(pos uint32) (disassembled string, newpos uint32, done bool) {
	b := r.w.GetByte(pos)
	pos++
	prstring := ""
	curpage := 0
	for i := 0; i < 8; i++ {
		if b&1 != 0 { // mark this one
			if !r.env.Valid() {
				r.w.AddComment(pos-2, "(!) cannot apply new page because a is not valid")
			} else {
				r.env.AToPage(curpage)
			}
			prstring += fmt.Sprintf("#%d,", curpage)
		}
		b >>= 1
		curpage++
	}
	if prstring == "" {
		fmt.Fprintf(os.Stderr, "tam defining nothing at $%X\n", pos-2)
		prstring = "<nothing>"
	} else {
		prstring = prstring[:len(prstring)-1] // strip trailing comma
	}
	return fmt.Sprintf("tam\t%s", prstring), pos, false
}

var tmapages = map[byte]int{
	0x01: 0,
	0x02: 1,
	0x04: 2,
	0x08: 3,
	0x10: 4,
	0x20: 5,
	0x40: 6,
	0x80: 7,
}

// tma #nn
func (r *Runner) tmaPageRegs(pos uint32) (disassembled string, newpos uint32, done bool) {
	b := r.w.GetByte(pos)
	pos++
	if _, ok := tmapages[b]; !ok {
		fmt.Fprintf(os.Stderr, "tma with invalid argument $%02X specified\n", b)
		r.env.Invalidate() // don't know what to do
		return fmt.Sprintf("tma\t<invalid $%02X>", b), pos, false
	}
	page := tmapages[b]
	r.env.PageToa(page)
	return fmt.Sprintf("tma\t#%d", page), pos, false
}

// these are only special because of their unique formats
// TODO do any of them touch a?

// tst #nn,zz
func (r *Runner) tstZeropage(pos uint32) (disassembled string, newpos uint32, done bool) {
	r.env.Invalidate()
	b := r.w.GetByte(pos)
	pos++
	z := r.w.GetByte(pos)
	pos++
	r.comment(pos-3, uint16(z))
	return fmt.Sprintf("tst\t#$%02X,%02X", b, z), pos, false
}

// tst #nn,zz,x
func (r *Runner) tstZeropagex(pos uint32) (disassembled string, newpos uint32, done bool) {
	r.env.Invalidate()
	b := r.w.GetByte(pos)
	pos++
	z := r.w.GetByte(pos)
	pos++
	r.comment(pos-3, uint16(z))
	return fmt.Sprintf("tst\t#$%02X,%02X,x", b, z), pos, false
}

// tst #nn,hhll
func (r *Runner) tstAbsolute(pos uint32) (disassembled string, newpos uint32, done bool) {
	r.env.Invalidate()
	b := r.w.GetByte(pos)
	pos++
	w, pos := r.w.GetWord(pos)
	r.comment(pos-4, w)
	return fmt.Sprintf("tst\t#$%02X,%04X", b, w), pos, false
}

// tst #nn,hhll,x
func (r *Runner) tstAbsolutex(pos uint32) (disassembled string, newpos uint32, done bool) {
	r.env.Invalidate()
	b := r.w.GetByte(pos)
	pos++
	w, pos := r.w.GetWord(pos)
	r.comment(pos-4, w)
	return fmt.Sprintf("tst\t#$%02X,%04X,x", b, w), pos, false
}
