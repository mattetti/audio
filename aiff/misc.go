package aiff

import "math"

// IeeeFloatToInt converts a 10 byte IEEE float into an int.
func IeeeFloatToInt(b [10]byte) int {
	var i uint32
	// Negative number
	if (b[0] & 0x80) == 1 {
		return 0
	}

	// Less than 1
	if b[0] <= 0x3F {
		return 1
	}

	// Too big
	if b[0] > 0x40 {
		return 67108864
	}

	// Still too big
	if b[0] == 0x40 && b[1] > 0x1C {
		return 800000000
	}

	i = (uint32(b[2]) << 23) | (uint32(b[3]) << 15) | (uint32(b[4]) << 7) | (uint32(b[5]) >> 1)
	i >>= (29 - uint32(b[1]))

	return int(i)
}

// IntToIeeeFloat converts an int into a 10 byte IEEE float.
func IntToIeeeFloat(i int) [10]byte {
	b := [10]byte{}
	num := float64(i)

	var sign int
	var expon int
	var fMant, fsMant float64
	var hiMant, loMant uint

	if num < 0 {
		sign = 0x8000
	} else {
		sign = 0
	}

	if num == 0 {
		expon = 0
		hiMant = 0
		loMant = 0
	} else {
		fMant, expon = math.Frexp(num)
		if (expon > 16384) || !(fMant < 1) { /* Infinity or NaN */
			expon = sign | 0x7FFF
			hiMant = 0
			loMant = 0 /* infinity */
		} else { /* Finite */
			expon += 16382
			if expon < 0 { /* denormalized */
				fMant = math.Ldexp(fMant, expon)
				expon = 0
			}
			expon |= sign
			fMant = math.Ldexp(fMant, 32)
			fsMant = math.Floor(fMant)
			hiMant = uint(fsMant)
			fMant = math.Ldexp(fMant-fsMant, 32)
			fsMant = math.Floor(fMant)
			loMant = uint(fsMant)
		}
	}

	b[0] = byte(expon >> 8)
	b[1] = byte(expon)
	b[2] = byte(hiMant >> 24)
	b[3] = byte(hiMant >> 16)
	b[4] = byte(hiMant >> 8)
	b[5] = byte(hiMant)
	b[6] = byte(loMant >> 24)
	b[7] = byte(loMant >> 16)
	b[8] = byte(loMant >> 8)
	b[9] = byte(loMant)

	return b
}

// intDataSize returns the size of the data required to represent the data when encoded.
// It returns zero if the type cannot be implemented by the fast path in Read or Write.
func intDataSize(data interface{}) int {
	switch data := data.(type) {
	case int8, *int8, *uint8:
		return 1
	case []int8:
		return len(data)
	case []uint8:
		return len(data)
	case int16, *int16, *uint16:
		return 2
	case []int16:
		return 2 * len(data)
	case []uint16:
		return 2 * len(data)
	case int32, *int32, *uint32:
		return 4
	case []int32:
		return 4 * len(data)
	case []uint32:
		return 4 * len(data)
	case int64, *int64, *uint64:
		return 8
	case []int64:
		return 8 * len(data)
	case []uint64:
		return 8 * len(data)
	}
	return 0
}

// Uint24to32 converts a 3 byte uint23 into a uint32
func Uint24to32(bytes []byte) uint32 {
	var output uint32
	output |= uint32(bytes[2]) << 0
	output |= uint32(bytes[1]) << 8
	output |= uint32(bytes[0]) << 16

	return output
}

// Uint32toUint24Bytes converts a uint32 into a 3 byte uint24 representation
func Uint32toUint24Bytes(n uint32) []byte {
	bytes := make([]byte, 3)
	bytes[0] = byte(n >> 16)
	bytes[1] = byte(n >> 8)
	bytes[2] = byte(n >> 0)

	return bytes
}
