package binutil

import (
	"fmt"
	"math"
)

const size = 8

type bits uint8

// BitVector is a convenience structure for setting inividual bits in an array
// of bytes. Create a new BitVector using 'new(BitVector)'. The BitVector will
// be automatically sized to accomadate the positions passed in, however you
// preallocate space by calling SizeForPosition with the highest bit position
// required.
type BitVector []bits

// NewBitVector from the passed bytes.
func NewBitVector(b []byte) BitVector {
	bv := BitVector{}
	l := len(b)
	if l == 0 {
		return bv
	}
	bv.SizeForPosition(uint(l*size - 1))
	for i := 0; i < l; i++ {
		bv[i] = bits(b[i])
	}
	return bv
}

// SizeForPosition resizes the BitVector to accomodate the position passed.
func (bv *BitVector) SizeForPosition(position uint) {
	requiredSize := int(position/size + 1)
	if len(*bv) < requiredSize {
		r := make([]bits, requiredSize)
		copy(r, *bv)
		*bv = r
	}
}

// Set the position passed to 1 in the BitVector
func (bv *BitVector) Set(position uint) {
	bv.SizeForPosition(position)
	(*bv)[position/size] |= 1 << (position % size)
}

// SetN sets the bits in the BitVector to the same values as the 'n' bits
// in the passed uint starting at the specified place.
func (bv *BitVector) SetN(value, n, start uint) {
	bv.SizeForPosition(n + start - 1)
	for i, j := start, uint(0); i < n+start; i, j = i+1, j+1 {
		b := value & (1 << j)
		if b == 0 {
			bv.UnSet(i)
		} else {
			bv.Set(i)
		}
	}
}

// ValueN gets the value of the first 'n' bits starts at the specified location
// and returns the result as a uint.
func (bv *BitVector) ValueN(n, start uint) uint {
	bv.SizeForPosition(n + start - 1)
	var result uint
	for i, j := start, uint(0); i < n+start; i, j = i+1, j+1 {
		v := bv.Value(i)
		if v > 0 {
			result |= (1 << j)
		}
	}
	return result
}

// UnSet the position passed in the BitVector. (Sets the position to 0.)
func (bv *BitVector) UnSet(position uint) {
	bv.SizeForPosition(position)
	(*bv)[position/size] &^= 1 << (position % size)
}

// Value of the position (0||1)
// This function will resize the underlying BitVector if the position is
// outside of the current bounds.
func (bv *BitVector) Value(position uint) int {
	bv.SizeForPosition(position)
	var val uint = 1 << (position % size)
	val &= uint((*bv)[position/size])
	r := 0
	if val > 0 {
		r = 1
	}
	return r
}

// Bytes of the BitVector
func (bv *BitVector) Bytes() []byte {
	length := len(*bv)
	r := make([]byte, length)
	for i := 0; i < length; i++ {
		b := (*bv)[i]
		r[i] = byte(b)
	}
	return r
}

// PrintHex prints the bytes in order (0->N) in their Hexadecimal
// representation.
func (bv *BitVector) PrintHex() {
	length := len(*bv)
	fmt.Print("BitVector{ ")
	for i := 0; i < length; i++ {
		fmt.Printf("%.2X ", uint((*bv)[i]))
	}
	fmt.Print("}\n")
}

// PrintN prints the size number bits in the passed uint from MSb to LSb
func PrintN(n, size uint) {
	// if size == math.MaxUint this will fail
	for i := size - 1; i < size && i != math.MaxUint; i-- {
		v := n & (1 << i)
		if v > 0 {
			fmt.Print(1)
		} else {
			fmt.Print(0)
		}
		if i%4 == 0 {
			fmt.Print(" ")
		}
	}
}

// Print the bit vector LSB to MSB and MSb to LSb within bytes.
func (bv *BitVector) Print() {
	nbytes := len(*bv)
	fmt.Printf("[%d]BitVector{ ", nbytes)
	// This is the proper way to render the binary so that it
	// reads most significant to least significant
	for i := 0; i < nbytes; i++ {
		PrintN(uint((*bv)[i]), size)
		fmt.Print(" ")
	}
	fmt.Print(" }\n")
}
