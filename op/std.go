package op

import (
	"fmt"
)

// xxx #nn
func (r *Runner) immediate(m string) Code {
	return func(pos uint32) (disassembled string, newpos uint32, done bool) {
		r.env.Invalidate()
		b := r.w.GetByte(pos)
		pos++
		return fmt.Sprintf("%s\t#$%02X", m, b), pos, false
	}
}

// xxx zz
func (r *Runner) zeropage(m string) Code {
	return func(pos uint32) (disassembled string, newpos uint32, done bool) {
		r.env.Invalidate()
		b := r.w.GetByte(pos)
		pos++
		r.comment(pos-2, uint16(b))
		return fmt.Sprintf("%s\t$%02X", m, b), pos, false
	}
}

// xxx zz,x
func (r *Runner) zeropagex(m string) Code {
	return func(pos uint32) (disassembled string, newpos uint32, done bool) {
		r.env.Invalidate()
		b := r.w.GetByte(pos)
		pos++
		r.comment(pos-2, uint16(b))
		return fmt.Sprintf("%s\t$%02X,x", m, b), pos, false
	}
}

// xxx zz,y
func (r *Runner) zeropagey(m string) Code {
	return func(pos uint32) (disassembled string, newpos uint32, done bool) {
		r.env.Invalidate()
		b := r.w.GetByte(pos)
		pos++
		r.comment(pos-2, uint16(b))
		return fmt.Sprintf("%s\t$%02X,y", m, b), pos, false
	}
}

// xxx (zz)
func (r *Runner) indirect(m string) Code {
	return func(pos uint32) (disassembled string, newpos uint32, done bool) {
		r.env.Invalidate()
		b := r.w.GetByte(pos)
		pos++
		r.comment(pos-2, uint16(b))
		return fmt.Sprintf("%s\t($%02X)", m, b), pos, false
	}
}

// xxx (zz,x)
func (r *Runner) indirectx(m string) Code {
	return func(pos uint32) (disassembled string, newpos uint32, done bool) {
		r.env.Invalidate()
		b := r.w.GetByte(pos)
		pos++
		r.comment(pos-2, uint16(b))
		return fmt.Sprintf("%s\t($%02X,x)", m, b), pos, false
	}
}

// xxx (zz),y
func (r *Runner) indirecty(m string) Code {
	return func(pos uint32) (disassembled string, newpos uint32, done bool) {
		r.env.Invalidate()
		b := r.w.GetByte(pos)
		pos++
		r.comment(pos-2, uint16(b))
		return fmt.Sprintf("%s\t($%02X),y", m, b), pos, false
	}
}

// xxx #nn,zz
func (r *Runner) zpBit(m string, n int) Code {
	return func(pos uint32) (disassembled string, newpos uint32, done bool) {
		r.env.Invalidate()
		b := r.w.GetByte(pos)
		pos++
		r.comment(pos-2, uint16(b))
		return fmt.Sprintf("%s\t#%d,$%02X", m, n, b), pos, false
	}
}

// xxx hhll
func (r *Runner) absolute(m string) Code {
	return func(pos uint32) (disassembled string, newpos uint32, done bool) {
		r.env.Invalidate()
		w, pos := r.w.GetWord(pos)
		r.comment(pos-3, w)
		return fmt.Sprintf("%s\t$%04X", m, w), pos, false
	}
}

// xxx hhll,x
func (r *Runner) absolutex(m string) Code {
	return func(pos uint32) (disassembled string, newpos uint32, done bool) {
		r.env.Invalidate()
		w, pos := r.w.GetWord(pos)
		r.comment(pos-3, w)
		return fmt.Sprintf("%s\t$%04X,x", m, w), pos, false
	}
}

// xxx hhll,y
func (r *Runner) absolutey(m string) Code {
	return func(pos uint32) (disassembled string, newpos uint32, done bool) {
		r.env.Invalidate()
		w, pos := r.w.GetWord(pos)
		r.comment(pos-3, w)
		return fmt.Sprintf("%s\t$%04X,y", m, w), pos, false
	}
}

// xxx a
func (r *Runner) accumulator(m string) Code {
	return func(pos uint32) (disassembled string, newpos uint32, done bool) {
		r.env.Invalidate()
		return fmt.Sprintf("%s\ta", m), pos, false
	}
}

// xxx
func (r *Runner) noArgs(m string) Code {
	return func(pos uint32) (disassembled string, newpos uint32, done bool) {
		r.env.Invalidate()
		return fmt.Sprintf("%s", m), pos, false
	}
}

// xxx hhll,hhll,hhll
func (r *Runner) transfer(m string) Code {
	return func(pos uint32) (disassembled string, newpos uint32, done bool) {
		r.env.Invalidate()
		src, pos := r.w.GetWord(pos) // TODO make labels for src and dest?
		dest, pos := r.w.GetWord(pos)
		length, pos := r.w.GetWord(pos)
		r.comment(pos-7, src)
		r.comment(pos-7, dest)
		return fmt.Sprintf("%s\t$%04X,$%04X,$%04X", m, src, dest, length), pos, false
	}
}
