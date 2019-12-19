package generator

import (
	"math/rand"
	"time"
)

var randomInt = rand.New(rand.NewSource(time.Now().UnixNano()))

// Int returns a random Int
func Int() int {
	return randomInt.Int()
}

// Int16 returns a random Int16
func Int16() int16 {
	return int16(randomInt.Int())
}

// UInt16 returns a random UInt16
func UInt16() uint16 {
	return uint16(randomInt.Int())
}

// Int32 returns a random Int32
func Int32() int32 {
	return int32(randomInt.Int())
}

// UInt32 returns a random UInt32
func UInt32() uint32 {
	return uint32(randomInt.Int())
}
