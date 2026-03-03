// Package readwriter reads ROM information
package readwriter

import (
	"fmt"
	"os"
)

const operandString = "---"

type ReadWriter struct {
	bytes           []byte
	instructions    map[uint32]string
	labels          map[uint32]string
	labelpriorities map[uint32]int
	labelplaces     map[uint32]uint32
	comments        map[uint32]string
}

func NewReadWriter(bytes []byte) *ReadWriter {
	return &ReadWriter{
		bytes:           bytes,
		instructions:    map[uint32]string{},
		labels:          map[uint32]string{},
		labelpriorities: map[uint32]int{},
		labelplaces:     map[uint32]uint32{},
		comments:        map[uint32]string{},
	}
}

func (w *ReadWriter) CheckPos(pos uint32) bool {
	if pos >= uint32(len(w.bytes)) {
		fmt.Fprintf(os.Stderr, "cannot disassemble at $%X as it is past ROM (size $%X bytes)\n", pos, len(w.bytes))
		return false
	}

	return true
}

func (w *ReadWriter) CheckInstruction(pos uint32) bool {
	_, already := w.instructions[pos]

	return already
}

func (w *ReadWriter) GetInstruction(pos uint32) string {
	return w.instructions[pos]
}

func (w *ReadWriter) SetInstruction(pos uint32, newpos uint32, s string) {
	w.instructions[pos] = s
	for i := pos + 1; i < newpos; i++ {
		w.instructions[i] = operandString
	}
}

func (w *ReadWriter) GetByte(pos uint32) byte {
	return w.bytes[pos]
}

func (w *ReadWriter) GetWord(pos uint32) (word uint16, newpos uint32) {
	word = uint16(w.bytes[pos])
	pos++
	word |= uint16(w.bytes[pos]) << 8
	pos++
	return word, pos
}
func (w *ReadWriter) AddLabel(pos uint32, label string, priority int) {
	if w.labels[pos] != "" { // if already defined as a different vector, concatenate the labels to make sure everything is represented
		// TODO because this uses a map, it will not be in vector order
		w.labels[pos] = w.labels[pos] + "_" + label
	} else {
		w.labels[pos] = label
	}
	w.labelpriorities[pos] = priority
}

func (w *ReadWriter) SetLabelPlace(pos uint32, labelpos uint32) {
	w.labelplaces[pos] = labelpos
}

// TODO MkLabel should watch for labels that cross into multi-byte instructions (that's what operandString is for)
func (w *ReadWriter) MkLabel(bpos uint32, prefix string, priority int) (label string) {
	mk := false
	if w.labels[bpos] == "" { // new label
		mk = true
	} else if w.labelpriorities[bpos] <= priority { // higher (or same) priority label
		mk = true
	}
	if mk {
		w.labels[bpos] = fmt.Sprintf("%s_%X", prefix, bpos)
		w.labelpriorities[bpos] = priority
	}
	return w.labels[bpos]
}

func (w *ReadWriter) AddComment(pos uint32, format string, args ...interface{}) {
	c := fmt.Sprintf(format, args...)
	if w.comments[pos] != "" {
		w.comments[pos] += " | "
	}
	w.comments[pos] += c
}

func (w *ReadWriter) AddIllegalOpComment(pos uint32) {
	w.AddComment(pos, "illegal opcode")
}

func (w *ReadWriter) AddFailComment(pos uint32, logical uint16, err error) {
	w.AddComment(pos, "$%04X - cannot get physical address (%v)", logical, err)
}

func (w *ReadWriter) AddOpComment(pos uint32, logical uint16, physical uint32) {
	w.AddComment(pos, "$%04X -> $%X", logical, physical)
}

func (w *ReadWriter) Print() {
	lbu32 := uint32(len(w.bytes))
	for i := uint32(0); i < lbu32; i++ {
		if label, ok := w.labels[i]; ok {
			fmt.Printf("%s:\n", label)
		}
		if instruction, ok := w.instructions[i]; ok && instruction != operandString {
			if labelpos, ok := w.labelplaces[i]; ok { // need to add a label
				if w.labels[labelpos] == "" {
					w.labels[labelpos] = fmt.Sprintf("<no label for $%X>", labelpos)
				}
				instruction = fmt.Sprintf(instruction, w.labels[labelpos])
			}
			fmt.Printf("\t%s\t\t; $%X", instruction, i)
			if comment, ok := w.comments[i]; ok {
				fmt.Printf(" | %s", comment)
			}
			fmt.Println()
		}
	}
}
