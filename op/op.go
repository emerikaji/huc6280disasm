// Package op runs the instructions from the opcodes to generate assembly.
package op

import (
	"github.com/emerikaji/huc6280disasm/readwriter"
	"github.com/emerikaji/huc6280disasm/system"
)

const (
	lpLoc = iota
	lpLocret
	lpSub
	lpUser
)

var vectorLocs = map[uint32]string{
	0x1FFE: "EntryPoint",
	0x1FFC: "NMI",
	0x1FFA: "TimerInterrupt",
	0x1FF8: "IRQ1",
	0x1FF6: "IRQ2_BRK",
}

// Code reference: http://shu.emuunlim.com/download/pcedocs/pce_cpu.html

// TODO adjust immediates so that they have effect on a
// TODO figure out which have no effect on a and make them not mark a as invalid

type Code func(pos uint32) (disassembled string, newpos uint32, done bool)

type Runner struct {
	env      *system.Environment
	w        *readwriter.ReadWriter
	useStack bool
	opcodes  [0x100]Code
}

func NewRunner(bytes []byte, useStack bool) *Runner {
	r := &Runner{
		env:      system.NewEnv(),
		w:        readwriter.NewReadWriter(bytes),
		useStack: useStack,
	}
	
	r.opcodes = [0x100]Code{
		// adc: add with carry
		0x69: r.immediate("adc"), // adc #nn
		0x65: r.zeropage("adc"),  // adc zz
		0x75: r.zeropagex("adc"), // adc zz,x
		0x72: r.indirect("adc"),  // adc (zz)
		0x61: r.indirectx("adc"), // adc (zz,x)
		0x71: r.indirecty("adc"), // adc (zz),y
		0x6D: r.absolute("adc"),  // adc hhll
		0x7D: r.absolutex("adc"), // adc hhll,x
		0x79: r.absolutey("adc"), // adc hhll,y

		// and: bitwise and
		0x29: r.immediate("and"), // and #nn
		0x25: r.zeropage("and"),  // and zz
		0x35: r.zeropagex("and"), // and zz,x
		0x32: r.indirect("and"),  // and (zz)
		0x21: r.indirectx("and"), // and (zz,x)
		0x31: r.indirecty("and"), // and (zz),y
		0x2D: r.absolute("and"),  // and hhll
		0x3D: r.absolutex("and"), // and hhll,x
		0x39: r.absolutey("and"), // and hhll,y

		// asl: arithmetic shift left
		0x06: r.zeropage("asl"),    // asl zz
		0x16: r.zeropagex("asl"),   // asl zz,x
		0x0E: r.absolute("asl"),    // asl hhll
		0x1E: r.absolutex("asl"),   // asl hhll,x
		0x0A: r.accumulator("asl"), // asl a

		// bbr: branch on bit clear (reset)
		0x0F: r.zpBitBr("bbr", 0), // bbr #0,zz,hhll
		0x1F: r.zpBitBr("bbr", 1), // bbr #1,zz,hhll
		0x2F: r.zpBitBr("bbr", 2), // bbr #2,zz,hhll
		0x3F: r.zpBitBr("bbr", 3), // bbr #3,zz,hhll
		0x4F: r.zpBitBr("bbr", 4), // bbr #4,zz,hhll
		0x5F: r.zpBitBr("bbr", 5), // bbr #5,zz,hhll
		0x6F: r.zpBitBr("bbr", 6), // bbr #6,zz,hhll
		0x7F: r.zpBitBr("bbr", 7), // bbr #7,zz,hhll

		// bcc: branch on carry clear
		0x90: r.branch("bcc"), // bcc hhll

		// bbs: branch on bit set
		0x8F: r.zpBitBr("bbs", 0), // bbs #0,zz,hhll
		0x9F: r.zpBitBr("bbs", 1), // bbs #1,zz,hhll
		0xAF: r.zpBitBr("bbs", 2), // bbs #2,zz,hhll
		0xBF: r.zpBitBr("bbs", 3), // bbs #3,zz,hhll
		0xCF: r.zpBitBr("bbs", 4), // bbs #4,zz,hhll
		0xDF: r.zpBitBr("bbs", 5), // bbs #5,zz,hhll
		0xEF: r.zpBitBr("bbs", 6), // bbs #6,zz,hhll
		0xFF: r.zpBitBr("bbs", 7), // bbs #7,zz,hhll

		// bcs: branch on carry set
		0xB0: r.branch("bcs"), // bcs hhll

		// beq: branch on equal
		0xF0: r.branch("beq"), // beq hhll

		// bit: test bit of accumulator
		0x89: r.immediate("bit"), // bit #nn
		0x24: r.zeropage("bit"),  // bit zz
		0x34: r.zeropagex("bit"), // bit zz,x
		0x2C: r.absolute("bit"),  // bit hhll
		0x3C: r.absolutex("bit"), // bit hhll,x

		// bmi: branch on minus
		0x30: r.branch("bmi"), // bmi hhll

		// bne: branch on not equal
		0xD0: r.branch("bne"), // bne hhll

		// bpl: branch on plus
		0x10: r.branch("bpl"), // bpl hhll

		// bra: branch
		0x80: r.branch("bra"), // bra hhll

		// brk: software break
		0x00: r.noArgs("brk"), // brk

		// bsr: branch to subroutine
		0x44: r.branch("bsr"), // bsr hhll
		// TODO make it produce a sub_ label

		// bvs: branch on overflow set
		0x70: r.branch("bvs"), // bvs hhll

		// bvc: branch on overflow clear
		0x50: r.branch("bvc"), // bvc hhll

		// clc: clear carry flag
		0x18: r.noArgs("clc"), // clc

		// cla: clear accumulator
		0x62: r.noArgs("cla"), // cla

		// cld: clear decimal flag
		0xD8: r.noArgs("cld"), // cld

		// cli: ENABLE interrupts (clears interrupt disable flag)
		0x58: r.noArgs("cli"), // cli

		// clv: clear overflow flag
		0xB8: r.noArgs("clv"), // clv

		// cly: clear y register
		0xC2: r.noArgs("cly"), // cly

		// clx: clear x register
		0x82: r.noArgs("clx"), // clx

		// cpx: compare x
		0xE0: r.immediate("cpx"), // cpx #nn
		0xE4: r.zeropage("cpx"),  // cpx zz
		0xEC: r.absolute("cpx"),  // cpx hhll

		// csh: set CPU speed to the higher speed
		0xD4: r.noArgs("csh"), // csh

		// csl: set CPU speed to the lower speed
		0x54: r.noArgs("csl"), // csl

		// cmp: compare a
		0xC9: r.immediate("cmp"), // cmp #nn
		0xC5: r.zeropage("cmp"),  // cmp zz
		0xD5: r.zeropagex("cmp"), // cmp zz,x
		0xD2: r.indirect("cmp"),  // cmp (zz)
		0xC1: r.indirectx("cmp"), // cmp (zz,x)
		0xD1: r.indirecty("cmp"), // cmp (zz),y
		0xCD: r.absolute("cmp"),  // cmp hhll
		0xDD: r.absolutex("cmp"), // cmp hhll,x
		0xD9: r.absolutey("cmp"), // cmp hhll,y

		// dex: decrement x
		0xCA: r.noArgs("dex"), // dex

		// dec: decrement
		0xC6: r.zeropage("dec"),  // dec zz
		0xD6: r.zeropagex("dec"), // dec zz,x
		0xCE: r.absolute("dec"),  // dec hhll
		0xDE: r.absolutex("dec"), // dec hhll,x
		0x3A: r.decAccumulator,   // dec a

		// cpy: compare y
		0xC0: r.immediate("cpy"), // cpy #nn
		0xC4: r.zeropage("cpy"),  // cpy zz
		0xCC: r.absolute("cpy"),  // cpy hhll

		// eor: exclusive or
		0x49: r.immediate("eor"), // eor #nn
		0x45: r.zeropage("eor"),  // eor zz
		0x55: r.zeropagex("eor"), // eor zz,x
		0x52: r.indirect("eor"),  // eor (zz)
		0x41: r.indirectx("eor"), // eor (zz,x)
		0x51: r.indirecty("eor"), // eor (zz),y
		0x4D: r.absolute("eor"),  // eor hhll
		0x5D: r.absolutex("eor"), // eor hhll,x
		0x59: r.absolutey("eor"), // eor hhll,y

		// inc: increment
		0xE6: r.zeropage("inc"),  // inc zz
		0xF6: r.zeropagex("inc"), // inc zz,x
		0xEE: r.absolute("inc"),  // inc hhll
		0xFE: r.absolutex("inc"), // inc hhll,x
		0x1A: r.incAccumulator,   // inc a

		// inx: increment x
		0xE8: r.noArgs("inx"), // inx

		// dey: decrement y
		0x88: r.noArgs("dey"), // dey

		// iny: increment y
		0xC8: r.noArgs("iny"), // iny

		// jmp: jump
		0x4C: r.jmpAbsolute,         // jmp hhll
		0x6C: r.jmpAbsoluteIndirect, // jmp (hhll)
		0x7C: r.jmpAbsolutex,        // jmp hhll,x

		// jsr: jump to subroutine
		0x20: r.jsrAbsolute, // jsr hhll

		// lda: load to a
		0xA9: r.ldaImmediate,     // lda #nn
		0xA5: r.zeropage("lda"),  // lda zz
		0xB5: r.zeropagex("lda"), // lda zz,x
		0xB2: r.indirect("lda"),  // lda (zz)
		0xA1: r.indirectx("lda"), // lda (zz,x)
		0xB1: r.indirecty("lda"), // lda (zz),y
		0xAD: r.absolute("lda"),  // lda hhll
		0xBD: r.absolutex("lda"), // lda hhll,x
		0xB9: r.absolutey("lda"), // lda hhll,y

		// ldx: load to x
		0xA2: r.immediate("ldx"), // ldx #nn
		0xA6: r.zeropage("ldx"),  // ldx zz
		0xB6: r.zeropagey("ldx"), // ldx zz,y
		0xAE: r.absolute("ldx"),  // ldx hhll
		0xBE: r.absolutey("ldx"), // ldx hhll,y

		// ldy: load to y
		// TODO - assuming x for the two indexed ones based on http://www6.atpages.jp/~appsouko/work/PCE/6280op.html as my original source for the opcodes say x but show y in its examples; what is correct, x or y?
		0xA0: r.immediate("ldy"), // ldy #nn
		0xA4: r.zeropage("ldy"),  // ldy zz
		0xB4: r.zeropagex("ldy"), // ldy zz,x
		0xAC: r.absolute("ldy"),  // ldy hhll
		0xBC: r.absolutex("ldy"), // ldy hhll,x

		// lsr: logical shift right
		0x46: r.zeropage("lsr"),    // lsr zz
		0x56: r.zeropagex("lsr"),   // lsr zz,x
		0x4E: r.absolute("lsr"),    // lsr hhll
		0x5E: r.absolutex("lsr"),   // lsr hhll,x
		0x4A: r.accumulator("lsr"), // lsr a

		// ora: bitwise or
		0x09: r.immediate("ora"), // ora #nn
		0x05: r.zeropage("ora"),  // ora zz
		0x15: r.zeropagex("ora"), // ora zz,x
		0x12: r.indirect("ora"),  // ora (zz)
		0x01: r.indirectx("ora"), // ora (zz,x)
		0x11: r.indirecty("ora"), // ora (zz),y
		0x0D: r.absolute("ora"),  // ora hhll
		0x1D: r.absolutex("ora"), // ora hhll,x
		0x19: r.absolutey("ora"), // ora hhll,y

		// nop: no operation
		0xEA: r.noArgs("nop"), // nop

		// pha: r.push a
		0x48: r.phaNoArgs, // pha

		// php: push p (status register)
		0x08: r.push("php"), // php

		// phx: push x
		0xDA: r.push("phx"), // phx

		// phy: push y
		0x5A: r.push("phy"), // phy

		// pla: pop a
		0x68: r.plaNoArgs, // pla

		// plp: pop p
		0x28: r.pop("plp"), // plp

		// plx: pop x
		0xFA: r.pop("plx"), // plx

		// ply: pop y
		0x7A: r.pop("ply"), // ply

		// rol: rotate left
		0x26: r.zeropage("rol"),    // rol zz
		0x36: r.zeropagex("rol"),   // rol zz,x
		0x2E: r.absolute("rol"),    // rol hhll
		0x3E: r.absolutex("rol"),   // rol hhll,x
		0x2A: r.accumulator("rol"), // rol a

		// rmb: clear (reset) bit
		0x07: r.zpBit("rmb", 0), // rmb #0,zz
		0x17: r.zpBit("rmb", 1), // rmb #1,zz
		0x27: r.zpBit("rmb", 2), // rmb #2,zz
		0x37: r.zpBit("rmb", 3), // rmb #3,zz
		0x47: r.zpBit("rmb", 4), // rmb #4,zz
		0x57: r.zpBit("rmb", 5), // rmb #5,zz
		0x67: r.zpBit("rmb", 6), // rmb #6,zz
		0x77: r.zpBit("rmb", 7), // rmb #7,zz

		// ror: rotate right
		0x66: r.zeropage("ror"),    // ror zz
		0x76: r.zeropagex("ror"),   // ror zz,x
		0x6E: r.absolute("ror"),    // ror hhll
		0x7E: r.absolutex("ror"),   // ror hhll,x
		0x6A: r.accumulator("ror"), // ror a

		// rti: return from interrupt
		0x40: r.rtiNoArgs, // rti

		// rts: return from subroutine
		0x60: r.rtsNoArgs, // rts

		// sax: swap a and x
		0x22: r.noArgs("sax"), // sax

		// say: swap a and y
		0x42: r.noArgs("say"), // say

		// sbc: subtract with borrow (carry)
		0xE9: r.immediate("sbc"), // sbc #nn
		0xE5: r.zeropage("sbc"),  // sbc zz
		0xF5: r.zeropagex("sbc"), // sbc zz,x
		0xF2: r.indirect("sbc"),  // sbc (zz)
		0xE1: r.indirectx("sbc"), // sbc (zz,x)
		0xF1: r.indirecty("sbc"), // sbc (zz),y
		0xED: r.absolute("sbc"),  // sbc hhll
		0xFD: r.absolutex("sbc"), // sbc hhll,x
		0xF9: r.absolutey("sbc"), // sbc hhll,y

		// sed: set decimal flag
		0xF8: r.noArgs("sed"), // sed

		// sec: set carry flag
		0x38: r.noArgs("sec"), // sec

		// sei: DISABLE interrupts (sets interrupt disable flag)
		0x78: r.noArgs("sei"), // sei

		// set: set T flag (changes next Code that operates on a to operate on the zero page address pointed to by x instead)
		0xF4: r.noArgs("set"), // set

		// st0: store in HuC6270 address register
		0x03: r.immediate("st0"), // st0 #nn

		// st1: store in HuC6270 data register low
		0x13: r.immediate("st1"), // st1 #nn

		// st2: store in HuC6270 data register high
		0x23: r.immediate("st2"), // st2 #nn

		// smb: set bit
		0x87: r.zpBit("smb", 0), // smb #0,zz
		0x97: r.zpBit("smb", 1), // smb #1,zz
		0xA7: r.zpBit("smb", 2), // smb #2,zz
		0xB7: r.zpBit("smb", 3), // smb #3,zz
		0xC7: r.zpBit("smb", 4), // smb #4,zz
		0xD7: r.zpBit("smb", 5), // smb #5,zz
		0xE7: r.zpBit("smb", 6), // smb #6,zz
		0xF7: r.zpBit("smb", 7), // smb #7,zz

		// sta: store a
		0x85: r.staZeropage,  // sta zz
		0x95: r.staZeropagex, // sta zz,x
		0x92: r.staIndirect,  // sta (zz)
		0x81: r.staIndirectx, // sta (zz,x)
		0x91: r.staIndirecty, // sta (zz),y
		0x8D: r.staAbsolute,  // sta hhll
		0x9D: r.staAbsolutex, // sta hhll,x
		0x99: r.staAbsolutey, // sta hhll,y

		// stx: store x
		0x86: r.zeropage("stx"),  // stx zz
		0x96: r.zeropagey("stx"), // stx zz,y
		0x8E: r.absolute("stx"),  // stx hhll

		// sty: store y
		0x84: r.zeropage("sty"),  // sty zz
		0x94: r.zeropagex("sty"), // sty zz,x
		0x8C: r.absolute("sty"),  // sty hhll

		// stz: store zero
		0x64: r.zeropage("stz"),  // stz zz
		0x74: r.zeropagex("stz"), // stz zz,x
		0x9C: r.absolute("stz"),  // stz hhll
		0x9E: r.absolutex("stz"), // stz hhll,x

		// tai: fill destination buffer with source word
		0xF3: r.transfer("tai"), // tai hhll,hhll,hhll

		// sxy: swap x and y
		0x02: r.noArgs("sxy"), // sxy

		// tam: transfer a to memory page register(s)
		0x53: r.tamPageRegs, // tam #nn,...

		// tax: transfer a to x
		0xAA: r.noArgs("tax"), // tax

		// tay: transfer a to y
		0xA8: r.noArgs("tay"), // tay

		// tia: copy source buffer to destination word (for instance, if destination word is a memory-mapped data port)
		0xE3: r.transfer("tia"), // tia hhll,hhll,hhll

		// tdd: copy source buffer GOING DOWN to destination buffer GOING DOWN
		0xC3: r.transfer("tdd"), // tdd hhll,hhll,hhll

		// tin: copy source buffer to destination byte
		0xD3: r.transfer("tin"), // tin hhll,hhll,hhll

		// tii: copy source buffer to destination buffer
		0x73: r.transfer("tii"), // tii hhll,hhll,hhll

		// tma: transfer memory page register to a
		0x43: r.tmaPageRegs, // tma #nn

		// trb: test value against bits set in a, then clear (reset) those bits in the value
		0x14: r.zeropage("trb"), // trb zz
		0x1C: r.absolute("trb"), // trb hhll

		// tsb: test value against bits set in a, then sets those bits in the value
		0x04: r.zeropage("tsb"), // tsb zz
		0x0C: r.absolute("tsb"), // tsb hhll

		// tst: test bits (and also clears them? I don't quite understand what's going on here)
		0x83: r.tstZeropage,  // tst #nn,zz
		0xA3: r.tstZeropagex, // tst #nn,zz,x
		0x93: r.tstAbsolute,  // tst #nn,hhll
		0xB3: r.tstAbsolutex, // tst #nn,hhll,x

		// tsx: transfer s to x
		0xBA: r.noArgs("tsx"), // tsx

		// txa: transfer x to a
		0x8A: r.noArgs("txa"), // txa

		// tya: transfer y to a
		0x98: r.noArgs("tya"), // tya

		// txs: transfer x to s
		0x9A: r.noArgs("txs"), // txs
	}

	return r
}

func (r *Runner) comment(pos uint32, logical uint16) {
	physical, err := r.env.Physical(logical)
	if err != nil {
		r.w.AddFailComment(pos-2, logical, err)
	} else {
		r.w.AddOpComment(pos-2, logical, physical)
	}
}

func (r *Runner) disassemble(pos uint32) {
	if !r.w.CheckPos(pos) {
		return // position outside rom
	}

	for {
		if r.w.CheckInstruction(pos) {
			break // reached a point we previously reached
		}

		b := r.w.GetByte(pos)
		opc := r.opcodes[b]
		if opc == nil {
			r.w.AddIllegalOpComment(pos)
		}

		s, newpos, done := opc(pos + 1)
		r.w.SetInstruction(pos, newpos, s)

		if done {
			break
		}

		pos = newpos
	}
}

func (r *Runner) StartDisassembly() (string, error) {
	// autoanalyze vectors
	for addr, label := range vectorLocs {
		posw, _ := r.w.GetWord(addr)
		pos, err := r.env.Physical(posw)
		if err != nil {
			return label, err
		}
		r.w.AddLabel(pos, label, lpSub)
		r.disassemble(pos)
	}

	// TODO read additional starts from standard input

	r.w.Print()

	return "", nil
}
