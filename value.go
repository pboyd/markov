package markov

import (
	"encoding/binary"
	"fmt"
	"math"
)

const (
	stringValue uint8 = iota
	uint64Value
	int64Value
	uintValue
	intValue
	uint32Value
	int32Value
	float32Value
	float64Value
)

func marshalValue(value interface{}) ([]byte, error) {
	switch v := value.(type) {
	case string:
		return marshalString(v)

	case uint64:
		return marshalUint64(v, uint64Value)
	case int64:
		return marshalUint64(uint64(v), int64Value)

	// uint and int are treated as 64 bit, even on 32 bit platforms.
	case uint:
		return marshalUint64(uint64(v), uintValue)
	case int:
		return marshalUint64(uint64(v), intValue)

	case uint32:
		return marshalUint32(v, uint32Value)
	case int32:
		// also covers rune
		return marshalUint32(uint32(v), int32Value)

	case float32:
		return marshalFloat32(v)
	case float64:
		return marshalFloat64(v)

	default:
		return nil, fmt.Errorf("unsupported type: %T", v)
	}

	return nil, nil
}

func unmarshalValue(buf []byte) (interface{}, error) {
	if len(buf) == 0 {
		return nil, nil
	}

	switch buf[0] {
	case stringValue:
		return unmarshalString(buf)

	case uint64Value:
		return unmarshalUint64(buf)
	case int64Value:
		v, err := unmarshalUint64(buf)
		if err != nil {
			return nil, err
		}
		return int64(v), nil

	// FIXME: If a uint or int value is set on a 64 bit
	case uintValue:
		v, err := unmarshalUint64(buf)
		if err != nil {
			return nil, err
		}
		return uint(v), nil
	case intValue:
		v, err := unmarshalUint64(buf)
		if err != nil {
			return nil, err
		}
		return int(v), nil

	case uint32Value:
		return unmarshalUint32(buf)
	case int32Value:
		v, err := unmarshalUint32(buf)
		if err != nil {
			return nil, err
		}
		return int32(v), nil

	case float32Value:
		return unmarshalFloat32(buf)
	case float64Value:
		return unmarshalFloat64(buf)

	default:
		return nil, fmt.Errorf("unsupported type id %d", buf[0])
	}

	return nil, nil
}

func marshalString(s string) ([]byte, error) {
	buf := make([]byte, len(s)+1)
	buf[0] = stringValue
	copy(buf[1:], s)
	return buf, nil
}

func unmarshalString(buf []byte) (string, error) {
	return string(buf[1:]), nil
}

func marshalUint64(i uint64, t byte) ([]byte, error) {
	buf := make([]byte, 9)
	buf[0] = t
	binary.BigEndian.PutUint64(buf[1:], i)
	return buf, nil
}

func unmarshalUint64(buf []byte) (uint64, error) {
	return binary.BigEndian.Uint64(buf[1:]), nil
}

func marshalUint32(i uint32, t byte) ([]byte, error) {
	buf := make([]byte, 5)
	buf[0] = t
	binary.BigEndian.PutUint32(buf[1:], i)
	return buf, nil
}

func unmarshalUint32(buf []byte) (uint32, error) {
	return binary.BigEndian.Uint32(buf[1:]), nil
}

func marshalFloat32(f float32) ([]byte, error) {
	return marshalUint32(math.Float32bits(f), float32Value)
}

func unmarshalFloat32(buf []byte) (float32, error) {
	bits, _ := unmarshalUint32(buf)
	return math.Float32frombits(bits), nil
}

func marshalFloat64(f float64) ([]byte, error) {
	return marshalUint64(math.Float64bits(f), float64Value)
}

func unmarshalFloat64(buf []byte) (float64, error) {
	bits, _ := unmarshalUint64(buf)
	return math.Float64frombits(bits), nil
}
