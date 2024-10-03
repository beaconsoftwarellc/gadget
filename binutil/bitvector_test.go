package binutil

import (
	"encoding/binary"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

var setTests = []struct {
	toSet    []uint
	expected uint
	size     uint
}{
	{[]uint{0}, 1, 8},
	{[]uint{1}, 2, 8},
	{[]uint{0, 1}, 3, 8},
	{[]uint{2}, 4, 8},
	{[]uint{0, 2}, 5, 8},
	{[]uint{1, 2}, 6, 8},
	{[]uint{0, 1, 2}, 7, 8},
	{[]uint{0, 8}, 257, 16},
}

func validateSet(t *testing.T, setBits []uint, actual, expected uint) {
	if actual != expected {
		t.Errorf("BitVector.Set(%v) = %d Expected %d", setBits, actual, expected)
	}
}

func TestSet(t *testing.T) {
	for _, st := range setTests {
		bv := new(BitVector)
		for _, idx := range st.toSet {
			bv.Set(idx)
		}
		switch st.size {
		case 8:
			e := uint8(st.expected)
			a := uint8(bv.Bytes()[0])
			validateSet(t, st.toSet, uint(a), uint(e))
		case 16:
			e := uint16(st.expected)
			a := binary.BigEndian.Uint16(bv.Bytes())
			validateSet(t, st.toSet, uint(a), uint(e))
		}
	}
}

func TestUnSet(t *testing.T) {
	// sanityCheck(0x1234)
	i := uint(0)
	assert.Equal(t, i-1, uint(math.MaxUint))
	t.SkipNow()
	var bv = new(BitVector)
	// 0-3 are unused
	// command
	// for i := 0; i < 0xF; i++ {
	// 	bv.Set(uint(i))
	// 	bv.Print()
	// }
	bv.SizeForPosition(16)
	bv.SetN(0x05>>1, 8, 0)
	bv.SetN(0x05<<7, 8, 8)

	// 9 - 12 are reserved
	// command 2
	bv.SetN(0x3F, 4, 8)
	// // source id
	bv.SetN(0x1234>>8, 8, 2*8)
	bv.SetN(0x1234, 8, 3*8)
	bv.SetN(0x08, 8, 4*8)
	t.Error("TestUnSet not implemented.")
}

func TestValue(t *testing.T) {
	t.SkipNow()
	t.Error("TestValue not implemented.")
}

func TestValueN(t *testing.T) {
	bv := new(BitVector)
	var expected, size, index uint = 0x0102, 16, 3
	bv.SetN(expected, size, index)
	var actual = bv.ValueN(size, index)
	if expected != actual {
		t.Errorf("BitVector.ValueN(%d, %d) = %d, Expected %d", size, index,
			actual, expected)
	}
}

func TestSizeForPosition(t *testing.T) {
	t.SkipNow()
	t.Error("TestSizeForPosition not implemented.")
}
