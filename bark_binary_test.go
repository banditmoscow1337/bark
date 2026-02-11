package bark

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"strings"
	"testing"
)

func TestBinaryLoggerAllTypes(t *testing.T) {
	var buf bytes.Buffer
	l := NewBinaryLogger(&buf)

	l.Info().
		Int("int", -1).
		Int8("int8", -8).
		Int16("int16", -16).
		Int32("int32", -32).
		Int64("int64", -64).
		Uint("uint", 1).
		Uint8("uint8", 8).
		Uint16("uint16", 16).
		Uint32("uint32", 32).
		Uint64("uint64", 64).
		Uintptr("uintptr", 128).
		Float32("float32", 1.5).
		Float64("float64", 2.5).
		Complex64("complex64", 1+2i).
		Complex128("complex128", 3+4i).
		Bool("bool_t", true).
		Bool("bool_f", false).
		Bytes("bytes", []byte{0xDE, 0xAD}).
		Str("str", "hi").
		Error(errors.New("err")).
		Msg("bin_msg")

	data := buf.Bytes()

	if len(data) < 6 {
		t.Fatal("data too short")
	}
	offset := 14

	readField := func() (string, byte, []byte) {
		if offset >= len(data) {
			t.Fatal("unexpected EOF")
		}

		kLen := int(data[offset])
		offset++
		key := string(data[offset : offset+kLen])
		offset += kLen

		tag := data[offset]
		offset++

		var val []byte
		switch tag {
		case BinTagInt, BinTagInt64, BinTagUint, BinTagUint64, BinTagUintptr, BinTagFloat64, BinTagComplex128:
			sz := 8
			if tag == BinTagComplex128 {
				sz = 16
			}
			val = data[offset : offset+sz]
			offset += sz
		case BinTagInt32, BinTagUint32, BinTagFloat32, BinTagComplex64:
			sz := 4
			if tag == BinTagComplex64 {
				sz = 8
			}
			val = data[offset : offset+sz]
			offset += sz
		case BinTagInt16, BinTagUint16:
			val = data[offset : offset+2]
			offset += 2
		case BinTagInt8, BinTagUint8, BinTagBool:
			val = data[offset : offset+1]
			offset += 1
		case BinTagString, BinTagErr, BinTagBytes:
			vLen := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
			offset += 2
			val = data[offset : offset+vLen]
			offset += vLen
		default:
			t.Fatalf("unknown tag %d at offset %d", tag, offset)
		}
		return key, tag, val
	}

	expected := []struct {
		key string
		tag uint8
		fn  func([]byte) bool
	}{
		{"int", BinTagInt, func(b []byte) bool { return int64(binary.LittleEndian.Uint64(b)) == -1 }},
		{"int8", BinTagInt8, func(b []byte) bool { return int8(b[0]) == -8 }},
		{"int16", BinTagInt16, func(b []byte) bool { return int16(binary.LittleEndian.Uint16(b)) == -16 }},
		{"int32", BinTagInt32, func(b []byte) bool { return int32(binary.LittleEndian.Uint32(b)) == -32 }},
		{"int64", BinTagInt64, func(b []byte) bool { return int64(binary.LittleEndian.Uint64(b)) == -64 }},
		{"uint", BinTagUint, func(b []byte) bool { return binary.LittleEndian.Uint64(b) == 1 }},
		{"uint8", BinTagUint8, func(b []byte) bool { return b[0] == 8 }},
		{"uint16", BinTagUint16, func(b []byte) bool { return binary.LittleEndian.Uint16(b) == 16 }},
		{"uint32", BinTagUint32, func(b []byte) bool { return binary.LittleEndian.Uint32(b) == 32 }},
		{"uint64", BinTagUint64, func(b []byte) bool { return binary.LittleEndian.Uint64(b) == 64 }},
		{"uintptr", BinTagUintptr, func(b []byte) bool { return binary.LittleEndian.Uint64(b) == 128 }},
		{"float32", BinTagFloat32, func(b []byte) bool { return math.Float32frombits(binary.LittleEndian.Uint32(b)) == 1.5 }},
		{"float64", BinTagFloat64, func(b []byte) bool { return math.Float64frombits(binary.LittleEndian.Uint64(b)) == 2.5 }},
		{"complex64", BinTagComplex64, func(b []byte) bool {
			r := math.Float32frombits(binary.LittleEndian.Uint32(b[:4]))
			i := math.Float32frombits(binary.LittleEndian.Uint32(b[4:]))
			return r == 1 && i == 2
		}},
		{"complex128", BinTagComplex128, func(b []byte) bool {
			r := math.Float64frombits(binary.LittleEndian.Uint64(b[:8]))
			i := math.Float64frombits(binary.LittleEndian.Uint64(b[8:]))
			return r == 3 && i == 4
		}},
		{"bool_t", BinTagBool, func(b []byte) bool { return b[0] == 1 }},
		{"bool_f", BinTagBool, func(b []byte) bool { return b[0] == 0 }},
		{"bytes", BinTagBytes, func(b []byte) bool { return bytes.Equal(b, []byte{0xDE, 0xAD}) }},
		{"str", BinTagString, func(b []byte) bool { return string(b) == "hi" }},
		{"error", BinTagErr, func(b []byte) bool { return string(b) == "err" }},
		{"message", BinTagString, func(b []byte) bool { return string(b) == "bin_msg" }},
	}

	for _, exp := range expected {
		k, tag, val := readField()
		if k != exp.key {
			t.Errorf("expected key %q, got %q", exp.key, k)
		}
		if tag != exp.tag {
			t.Errorf("key %s: expected tag %d, got %d", k, exp.tag, tag)
		}
		if !exp.fn(val) {
			t.Errorf("key %s: value verification failed", k)
		}
	}
}

func TestBinaryLoggerEdgeCases(t *testing.T) {
	var buf bytes.Buffer
	l := NewBinaryLogger(&buf)

	l.Info().Error(nil).Msg("no_err")
	data := buf.Bytes()

	if bytes.Contains(data, []byte{5, 'e', 'r', 'r', 'o', 'r'}) {
		t.Error("found error key in binary output when error was nil")
	}

	buf.Reset()

	hugeStr := strings.Repeat("A", 70000)
	hugeKey := strings.Repeat("K", 300)
	l.Info().Str(hugeKey, hugeStr).Msg("huge")

	data = buf.Bytes()
	offset := 14

	kLen := int(data[offset])
	if kLen != 255 {
		t.Errorf("expected key truncated to 255, got %d", kLen)
	}
	offset++
	offset += kLen

	tag := data[offset]
	if tag != BinTagString {
		t.Fatal("expected string tag")
	}
	offset++

	vLen := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
	if vLen != 65535 {
		t.Errorf("expected value truncated to 65535, got %d", vLen)
	}
	
	buf.Reset()
	hugeBytes := make([]byte, 70000)
	l.Info().Bytes("b", hugeBytes).Msg("huge_bytes")
	data = buf.Bytes()
	vLen = int(binary.LittleEndian.Uint16(data[17:19]))
	if vLen != 65535 {
		t.Errorf("expected bytes truncated to 65535, got %d", vLen)
	}
}

func BenchmarkLogger(b *testing.B) {
	l := NewLogger(io.Discard)
	b.ReportAllocs()

	for b.Loop() {
		l.Info().
			Str("key", "value").
			Int("id", 1234).
			Float64("pi", 3.14).
			Bool("enabled", true).
			Msg("benchmark")
	}
}

func BenchmarkBinaryLogger(b *testing.B) {
	l := NewBinaryLogger(io.Discard)
	b.ReportAllocs()

	for b.Loop() {
		l.Info().
			Str("key", "value").
			Int("id", 1234).
			Float64("pi", 3.14).
			Bool("enabled", true).
			Msg("benchmark")
	}
}