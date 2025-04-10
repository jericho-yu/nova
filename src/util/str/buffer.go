package str

import "bytes"

type Buffer struct{ original *bytes.Buffer }

var BufferApp Buffer

func (*Buffer) NewByString(original string) *Buffer { return &Buffer{bytes.NewBufferString(original)} }

func (*Buffer) NewByBytes(original []byte) *Buffer { return &Buffer{bytes.NewBuffer(original)} }

func (my *Buffer) String(stringList ...string) *Buffer {
	for _, s := range stringList {
		my.original.WriteString(s)
	}

	return my
}

func (my *Buffer) Byte(byteList ...byte) *Buffer {
	for _, b := range byteList {
		my.original.WriteByte(b)
	}

	return my
}

func (my *Buffer) Rune(runeList ...rune) *Buffer {
	for _, v := range runeList {
		my.original.WriteRune(v)
	}

	return my
}

func (my *Buffer) ToString() string { return my.original.String() }

func (my *Buffer) ToBytes() []byte { return my.original.Bytes() }
