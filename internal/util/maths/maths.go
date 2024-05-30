package maths

func IntAbs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func Max(x, y int) int {
	if x < y {
		return y
	}
	return x
}

func Min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

func Uint32Max(x, y uint32) uint32 {
	if x > y {
		return x
	}
	return y
}

func Int64Max(x, y int64) int64 {
	if x < y {
		return y
	}
	return x
}

func SetBit(curValue, targetBit int64, bitValue bool) int64 {
	if bitValue {
		return curValue | targetBit
	} else {
		return curValue &^ targetBit
	}
}

func Xor(x, y int64) int64 {
	return x ^ y
}
