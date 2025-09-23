package bpack

import (
	"errors"
	"io"

	"github.com/brickingsoft/bytebuffers"
)

var errVarintOverflow = errors.New("varint integer overflow")

func writeVarInt(buf bytebuffers.Buffer, prefix byte, n byte, i uint64) {
	k := uint64((1 << n) - 1)
	if i < k {
		b := byte(i)
		if prefix != 0 {
			b ^= prefix
		}
		_ = buf.WriteByte(b)
		return
	}
	b := byte(k)
	if prefix != 0 {
		b ^= prefix
	}
	_ = buf.WriteByte(b)
	i -= k
	for ; i >= 128; i >>= 7 {
		_ = buf.WriteByte(byte(0x80 | (i & 0x7f)))
	}
	_ = buf.WriteByte(byte(i))
}

func readVarInt(buf bytebuffers.Buffer, n byte) (i uint64, err error) {
	if n < 1 || n > 8 {
		panic("bad n")
	}
	if buf.Len() == 0 {
		err = io.EOF
		return
	}
	prefix, _ := buf.ReadByte()
	i = uint64(prefix)
	if n < 8 {
		i &= (1 << uint64(n)) - 1
	}
	if i < (1<<uint64(n))-1 {
		return
	}
	var m uint64
	for buf.Len() > 0 {
		b, _ := buf.ReadByte()
		i += uint64(b&127) << m
		if b&128 == 0 {
			return
		}
		m += 7
		if m >= 63 { // TODO: proper overflow check. making this up.
			err = errVarintOverflow
			return
		}
	}
	return
}
