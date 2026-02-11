package bark

import (
	"encoding/binary"
	"io"
	"math"
	"sync"
	"time"
)

const (
	BinTypeInfo = uint16(1)
	BinTagString     = uint8(1)
	BinTagInt        = uint8(2)
	BinTagInt8       = uint8(3)
	BinTagInt16      = uint8(4)
	BinTagInt32      = uint8(5)
	BinTagInt64      = uint8(6)
	BinTagUint       = uint8(7)
	BinTagUint8      = uint8(8)
	BinTagUint16     = uint8(9)
	BinTagUint32     = uint8(10)
	BinTagUint64     = uint8(11)
	BinTagFloat32    = uint8(12)
	BinTagFloat64    = uint8(13)
	BinTagBool       = uint8(14)
	BinTagErr        = uint8(15)
	BinTagComplex64  = uint8(16)
	BinTagComplex128 = uint8(17)
	BinTagUintptr    = uint8(18)
	BinTagBytes      = uint8(19)
)

type BinaryLogger struct {
	pool sync.Pool
	out  io.Writer
}

type BinaryEvent struct {
	buf  []byte
	out  io.Writer
	pool *sync.Pool
}

func NewBinaryLogger(w io.Writer) *BinaryLogger {
	l := &BinaryLogger{
		out: w,
	}
	l.pool.New = func() any {
		return &BinaryEvent{
			buf: make([]byte, 0, 512),
			out:  w,
			pool: &l.pool,
		}
	}
	return l
}

func (l *BinaryLogger) Info() *BinaryEvent {
	e := l.pool.Get().(*BinaryEvent)
	e.buf = e.buf[:0]
	e.buf = append(e.buf, 0, 0, 0, 0, 0, 0)
	e.buf = binary.LittleEndian.AppendUint64(e.buf, uint64(time.Now().UnixNano()))

	return e
}

// appendKey adds [KeyLen][KeyBytes]
func (e *BinaryEvent) appendKey(key string) {
	if len(key) > 255 {
		key = key[:255]
	}
	e.buf = append(e.buf, uint8(len(key)))
	e.buf = append(e.buf, key...)
}

func (e *BinaryEvent) Str(key, val string) *BinaryEvent {
	e.appendKey(key)
	e.buf = append(e.buf, BinTagString)
	if len(val) > 65535 {
		val = val[:65535]
	}
	e.buf = binary.LittleEndian.AppendUint16(e.buf, uint16(len(val)))
	e.buf = append(e.buf, val...)
	return e
}

func (e *BinaryEvent) Bytes(key string, val []byte) *BinaryEvent {
	e.appendKey(key)
	e.buf = append(e.buf, BinTagBytes)
	vLen := len(val)
	if vLen > 65535 {
		vLen = 65535
	}
	e.buf = binary.LittleEndian.AppendUint16(e.buf, uint16(vLen))
	e.buf = append(e.buf, val[:vLen]...)
	return e
}

// Integers

func (e *BinaryEvent) Int(key string, val int) *BinaryEvent {
	e.appendKey(key)
	e.buf = append(e.buf, BinTagInt)
	e.buf = binary.LittleEndian.AppendUint64(e.buf, uint64(val))
	return e
}

func (e *BinaryEvent) Int8(key string, val int8) *BinaryEvent {
	e.appendKey(key)
	e.buf = append(e.buf, BinTagInt8)
	e.buf = append(e.buf, uint8(val))
	return e
}

func (e *BinaryEvent) Int16(key string, val int16) *BinaryEvent {
	e.appendKey(key)
	e.buf = append(e.buf, BinTagInt16)
	e.buf = binary.LittleEndian.AppendUint16(e.buf, uint16(val))
	return e
}

func (e *BinaryEvent) Int32(key string, val int32) *BinaryEvent {
	e.appendKey(key)
	e.buf = append(e.buf, BinTagInt32)
	e.buf = binary.LittleEndian.AppendUint32(e.buf, uint32(val))
	return e
}

func (e *BinaryEvent) Int64(key string, val int64) *BinaryEvent {
	e.appendKey(key)
	e.buf = append(e.buf, BinTagInt64)
	e.buf = binary.LittleEndian.AppendUint64(e.buf, uint64(val))
	return e
}

// Unsigned Integers

func (e *BinaryEvent) Uint(key string, val uint) *BinaryEvent {
	e.appendKey(key)
	e.buf = append(e.buf, BinTagUint)
	e.buf = binary.LittleEndian.AppendUint64(e.buf, uint64(val))
	return e
}

func (e *BinaryEvent) Uint8(key string, val uint8) *BinaryEvent {
	e.appendKey(key)
	e.buf = append(e.buf, BinTagUint8)
	e.buf = append(e.buf, val)
	return e
}

func (e *BinaryEvent) Uint16(key string, val uint16) *BinaryEvent {
	e.appendKey(key)
	e.buf = append(e.buf, BinTagUint16)
	e.buf = binary.LittleEndian.AppendUint16(e.buf, val)
	return e
}

func (e *BinaryEvent) Uint32(key string, val uint32) *BinaryEvent {
	e.appendKey(key)
	e.buf = append(e.buf, BinTagUint32)
	e.buf = binary.LittleEndian.AppendUint32(e.buf, val)
	return e
}

func (e *BinaryEvent) Uint64(key string, val uint64) *BinaryEvent {
	e.appendKey(key)
	e.buf = append(e.buf, BinTagUint64)
	e.buf = binary.LittleEndian.AppendUint64(e.buf, val)
	return e
}

func (e *BinaryEvent) Uintptr(key string, val uintptr) *BinaryEvent {
	e.appendKey(key)
	e.buf = append(e.buf, BinTagUintptr)
	e.buf = binary.LittleEndian.AppendUint64(e.buf, uint64(val))
	return e
}

// Floats

func (e *BinaryEvent) Float32(key string, val float32) *BinaryEvent {
	e.appendKey(key)
	e.buf = append(e.buf, BinTagFloat32)
	e.buf = binary.LittleEndian.AppendUint32(e.buf, math.Float32bits(val))
	return e
}

func (e *BinaryEvent) Float64(key string, val float64) *BinaryEvent {
	e.appendKey(key)
	e.buf = append(e.buf, BinTagFloat64)
	e.buf = binary.LittleEndian.AppendUint64(e.buf, math.Float64bits(val))
	return e
}

// Complex

func (e *BinaryEvent) Complex64(key string, val complex64) *BinaryEvent {
	e.appendKey(key)
	e.buf = append(e.buf, BinTagComplex64)
	e.buf = binary.LittleEndian.AppendUint32(e.buf, math.Float32bits(real(val)))
	e.buf = binary.LittleEndian.AppendUint32(e.buf, math.Float32bits(imag(val)))
	return e
}

func (e *BinaryEvent) Complex128(key string, val complex128) *BinaryEvent {
	e.appendKey(key)
	e.buf = append(e.buf, BinTagComplex128)
	e.buf = binary.LittleEndian.AppendUint64(e.buf, math.Float64bits(real(val)))
	e.buf = binary.LittleEndian.AppendUint64(e.buf, math.Float64bits(imag(val)))
	return e
}

// Others

func (e *BinaryEvent) Bool(key string, val bool) *BinaryEvent {
	e.appendKey(key)
	e.buf = append(e.buf, BinTagBool)
	if val {
		e.buf = append(e.buf, 1)
	} else {
		e.buf = append(e.buf, 0)
	}
	return e
}

func (e *BinaryEvent) Error(err error) *BinaryEvent {
	if err == nil {
		return e
	}
	e.appendKey("error")
	e.buf = append(e.buf, BinTagErr)
	msg := err.Error()
	if len(msg) > 65535 {
		msg = msg[:65535]
	}
	e.buf = binary.LittleEndian.AppendUint16(e.buf, uint16(len(msg)))
	e.buf = append(e.buf, msg...)
	return e
}

func (e *BinaryEvent) Msg(msg string) {
	e.Str("message", msg)
	payloadSize := len(e.buf) - 6
	binary.LittleEndian.PutUint16(e.buf[0:2], BinTypeInfo)
	binary.LittleEndian.PutUint32(e.buf[2:6], uint32(payloadSize))

	e.out.Write(e.buf)
	e.pool.Put(e)
}