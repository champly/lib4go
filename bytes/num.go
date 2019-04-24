package bytes

import "unsafe"

// Int8ToSliceByte int8 to byte
func Int8ToSliceByte(s int8) byte {
	return *(*byte)(unsafe.Pointer(&s))
}

// Uint16ToSliceByte uint16 to [2]byte
func Uint16ToSliceByte(s uint16) [2]byte {
	return *(*[2]byte)(unsafe.Pointer(&s))
}

// Int16ToSliceByte int16 to [2]byte
func Int16ToSliceByte(s int16) [2]byte {
	return *(*[2]byte)(unsafe.Pointer(&s))
}

// Uint32ToSliceByte uint32 to [4]byte
func Uint32ToSliceByte(s uint32) [4]byte {
	return *(*[4]byte)(unsafe.Pointer(&s))
}

// Int32ToSliceByte int32 to [4]byte
func Int32ToSliceByte(s int32) [4]byte {
	return *(*[4]byte)(unsafe.Pointer(&s))
}

// Uint64ToSliceByte uint64 to [8]byte
func Uint64ToSliceByte(s uint64) [8]byte {
	return *(*[8]byte)(unsafe.Pointer(&s))
}

// Int64ToSliceByte int64 to [8]byte
func Int64ToSliceByte(s int64) [8]byte {
	return *(*[8]byte)(unsafe.Pointer(&s))
}
