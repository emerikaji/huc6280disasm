package op

import (
	"fmt"
	"os"
)

func (r *Runner) doBranch(pos uint32) (labelpos uint32, newpos uint32) {
	b := r.w.GetByte(pos)
	pos++
	offset := int32(int8(b))
	// TODO does not properly handle jumps across page boundaries
	bpos := uint32(int32(pos) + offset)
	r.w.MkLabel(bpos, "loc", lpLoc)
	r.disassemble(bpos)
	return bpos, pos
}

// xxx label
func (r *Runner) branch(m string) Code {
	return func(pos uint32) (disassembled string, newpos uint32, done bool) {
		labelpos, pos := r.doBranch(pos)
		r.w.SetLabelPlace(pos-2, labelpos)
		return fmt.Sprintf("%s\t%%s", m), pos, false
	}
}

// xxx #nn,zz,label
func (r *Runner) zpBitBr(m string, n int) Code {
	return func(pos uint32) (disassembled string, newpos uint32, done bool) {
		b := r.w.GetByte(pos)
		pos++
		labelpos, pos := r.doBranch(pos)
		r.w.SetLabelPlace(pos-3, labelpos)
		r.comment(pos-3, uint16(b))
		return fmt.Sprintf("%s\t#%d,$%02X,%%s", m, n, b), pos, false
	}
}

// jmp hhll
func (r *Runner) jmpAbsolute(pos uint32) (disassembled string, newpos uint32, done bool) {
	w, pos := r.w.GetWord(pos)
	phys, err := r.env.Physical(w)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot get physical address for jmp to $%04X: %v\n", w, err)
		return fmt.Sprintf("jmp\t$%04X", w), pos, true
	}
	r.w.MkLabel(phys, "loc", lpLoc)
	r.w.SetLabelPlace(pos-3, phys)
	r.disassemble(phys)
	return fmt.Sprintf("jmp\t%%s"), pos, true
}

// jmp hhll,x
func (r *Runner) jmpAbsolutex(pos uint32) (disassembled string, newpos uint32, done bool) {
	w, pos := r.w.GetWord(pos)
	r.comment(pos-3, w)
	return fmt.Sprintf("jmp\t$%04X,x", w), pos, true
}

// jmp (hhll)
func (r *Runner) jmpAbsoluteIndirect(pos uint32) (disassembled string, newpos uint32, done bool) {
	w, pos := r.w.GetWord(pos)
	r.comment(pos-3, w)
	return fmt.Sprintf("jmp\t($%04X)", w), pos, true
}

// jsr hhll
func (r *Runner) jsrAbsolute(pos uint32) (disassembled string, newpos uint32, done bool) {
	w, pos := r.w.GetWord(pos)
	phys, err := r.env.Physical(w)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot get physical address for jsr to $%04X: %v\n", w, err)
		return fmt.Sprintf("jsr\t$%04X", w), pos, true
	}
	r.w.MkLabel(phys, "sub", lpSub)
	r.w.SetLabelPlace(pos-3, phys)
	r.disassemble(phys)
	return fmt.Sprintf("jsr\t%%s"), pos, false
}

// rti
func (r *Runner) rtiNoArgs(pos uint32) (disassembled string, newpos uint32, done bool) {
	return "rti", pos, true
}

// rts
func (r *Runner) rtsNoArgs(pos uint32) (disassembled string, newpos uint32, done bool) {
	return "rts", pos, true
}
