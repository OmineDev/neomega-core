package big_endian

import "math"

type CanReadOutBytes interface {
	ReadOut(len int) (b []byte, err error)
}

// Uint16 ...
func Uint16(r CanReadOutBytes) (uint16, error) {
	b, err := r.ReadOut(2)
	if err != nil {
		return 0, err
	}
	return uint16(b[0])<<8 | uint16(b[1]), nil
}

// Int16 ...
func Int16(r CanReadOutBytes) (int16, error) {
	u16, err := Uint16(r)
	return int16(u16), err
}

// Uint32 ...
func Uint32(r CanReadOutBytes) (uint32, error) {
	b, err := r.ReadOut(4)
	if err != nil {
		return 0, err
	}
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3]), nil
}

// Int32 ...
func Int32(r CanReadOutBytes) (int32, error) {
	u32, err := Uint32(r)
	return int32(u32), err
}

func Uint64(r CanReadOutBytes) (uint64, error) {
	b, err := r.ReadOut(8)
	if err != nil {
		return 0, err
	}
	return uint64(b[0])<<56 | uint64(b[1])<<48 | uint64(b[2])<<40 | uint64(b[3])<<32 |
		uint64(b[4])<<24 | uint64(b[5])<<16 | uint64(b[6])<<8 | uint64(b[7]), nil
}

// Int64 ...
func Int64(r CanReadOutBytes) (int64, error) {
	u64, err := Uint64(r)
	return int64(u64), err
}

// Float32 ...
func Float32(r CanReadOutBytes) (float32, error) {
	b, err := r.ReadOut(4)
	if err != nil {
		return 0, err
	}
	return math.Float32frombits(uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])), nil
}

// Float64 ...
func Float64(r CanReadOutBytes) (float64, error) {
	b, err := r.ReadOut(8)
	if err != nil {
		return 0, err
	}
	return math.Float64frombits(uint64(b[0])<<56 | uint64(b[1])<<48 | uint64(b[2])<<40 | uint64(b[3])<<32 |
		uint64(b[4])<<24 | uint64(b[5])<<16 | uint64(b[6])<<8 | uint64(b[7])), nil
}

// String ...
func String(r CanReadOutBytes) (string, error) {
	b, err := r.ReadOut(2)
	if err != nil {
		return "", err
	}
	stringLength := int(uint16(b[0])<<8 | uint16(b[1]))
	data, err := r.ReadOut(stringLength)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
