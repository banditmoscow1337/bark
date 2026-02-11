package bark

import (
	"encoding/base64"
	"io"
	"strconv"
	"sync"
	"time"
)

const hex = "0123456789abcdef"
var escapeTable [256]uint8

func init() {
	for i := range 32 {
		escapeTable[i] = 1
	}
	escapeTable['"'] = 1
	escapeTable['\\'] = 1
}

type Logger struct {
	pool sync.Pool
	out  io.Writer
}

type Event struct {
	buf  []byte
	out  io.Writer
	pool *sync.Pool
}

func NewLogger(w io.Writer) *Logger {
	l := &Logger{
		out: w,
	}
	l.pool.New = func() any {
		return &Event{
			buf: make([]byte, 0, 512),
			out:  w,
			pool: &l.pool,
		}
	}
	return l
}

func (l *Logger) Info() *Event {
	e := l.pool.Get().(*Event)
	e.buf = e.buf[:0]
	e.buf = append(e.buf, `{"level":"info","time":"`...)
	e.buf = appendTime(e.buf, time.Now())
	e.buf = append(e.buf, '"', ',')
	return e
}

func (e *Event) appendKey(key string) {
	e.buf = append(e.buf, '"')
	e.buf = append(e.buf, key...)
	e.buf = append(e.buf, '"', ':')
}

func (e *Event) Str(key, val string) *Event {
	e.appendKey(key)
	e.buf = appendString(e.buf, val)
	e.buf = append(e.buf, ',')
	return e
}

func (e *Event) Bytes(key string, val []byte) *Event {
	e.appendKey(key)
	e.buf = append(e.buf, '"')
	encodedLen := base64.StdEncoding.EncodedLen(len(val))
	if cap(e.buf)-len(e.buf) < encodedLen {
		newBuf := make([]byte, len(e.buf), len(e.buf)+encodedLen+32)
		copy(newBuf, e.buf)
		e.buf = newBuf
	}
	currentLen := len(e.buf)
	e.buf = e.buf[:currentLen+encodedLen]
	base64.StdEncoding.Encode(e.buf[currentLen:], val)
	e.buf = append(e.buf, '"', ',')
	return e
}

func (e *Event) Int(key string, val int) *Event {
	return e.Int64(key, int64(val))
}

func (e *Event) Int8(key string, val int8) *Event {
	return e.Int64(key, int64(val))
}

func (e *Event) Int16(key string, val int16) *Event {
	return e.Int64(key, int64(val))
}

func (e *Event) Int32(key string, val int32) *Event {
	return e.Int64(key, int64(val))
}

func (e *Event) Int64(key string, val int64) *Event {
	e.appendKey(key)
	e.buf = strconv.AppendInt(e.buf, val, 10)
	e.buf = append(e.buf, ',')
	return e
}

func (e *Event) Uint(key string, val uint) *Event {
	return e.Uint64(key, uint64(val))
}

func (e *Event) Uint8(key string, val uint8) *Event {
	return e.Uint64(key, uint64(val))
}

func (e *Event) Uint16(key string, val uint16) *Event {
	return e.Uint64(key, uint64(val))
}

func (e *Event) Uint32(key string, val uint32) *Event {
	return e.Uint64(key, uint64(val))
}

func (e *Event) Uint64(key string, val uint64) *Event {
	e.appendKey(key)
	e.buf = strconv.AppendUint(e.buf, val, 10)
	e.buf = append(e.buf, ',')
	return e
}

func (e *Event) Uintptr(key string, val uintptr) *Event {
	return e.Uint64(key, uint64(val))
}

func (e *Event) Float32(key string, val float32) *Event {
	e.appendKey(key)
	e.buf = strconv.AppendFloat(e.buf, float64(val), 'f', -1, 32)
	e.buf = append(e.buf, ',')
	return e
}

func (e *Event) Float64(key string, val float64) *Event {
	e.appendKey(key)
	e.buf = strconv.AppendFloat(e.buf, val, 'f', -1, 64)
	e.buf = append(e.buf, ',')
	return e
}

func (e *Event) Complex64(key string, val complex64) *Event {
	e.appendKey(key)
	e.buf = append(e.buf, '"', '(')
	e.buf = strconv.AppendFloat(e.buf, float64(real(val)), 'f', -1, 32)
	e.buf = append(e.buf, '+')
	e.buf = strconv.AppendFloat(e.buf, float64(imag(val)), 'f', -1, 32)
	e.buf = append(e.buf, 'i', ')', '"', ',')
	return e
}

func (e *Event) Complex128(key string, val complex128) *Event {
	e.appendKey(key)
	e.buf = append(e.buf, '"', '(')
	e.buf = strconv.AppendFloat(e.buf, real(val), 'f', -1, 64)
	e.buf = append(e.buf, '+')
	e.buf = strconv.AppendFloat(e.buf, imag(val), 'f', -1, 64)
	e.buf = append(e.buf, 'i', ')', '"', ',')
	return e
}

func (e *Event) Bool(key string, val bool) *Event {
	e.appendKey(key)
	if val {
		e.buf = append(e.buf, 't', 'r', 'u', 'e', ',')
	} else {
		e.buf = append(e.buf, 'f', 'a', 'l', 's', 'e', ',')
	}
	return e
}

func (e *Event) Error(err error) *Event {
	if err == nil {
		return e
	}
	e.buf = append(e.buf, `"error":`...)
	e.buf = appendString(e.buf, err.Error())
	e.buf = append(e.buf, ',')
	return e
}

func (e *Event) Msg(msg string) {
	e.buf = append(e.buf, `"message":`...)
	e.buf = appendString(e.buf, msg)
	e.buf = append(e.buf, '}', '\n')
	e.out.Write(e.buf)
	e.pool.Put(e)
}

func appendString(dst []byte, s string) []byte {
	dst = append(dst, '"')
	start := 0
	for i := 0; i < len(s); i++ {
		if escapeTable[s[i]] != 0 {
			if start < i {
				dst = append(dst, s[start:i]...)
			}
			switch s[i] {
			case '"':
				dst = append(dst, '\\', '"')
			case '\\':
				dst = append(dst, '\\', '\\')
			case '\n':
				dst = append(dst, '\\', 'n')
			case '\r':
				dst = append(dst, '\\', 'r')
			case '\t':
				dst = append(dst, '\\', 't')
			case '\b':
				dst = append(dst, '\\', 'b')
			case '\f':
				dst = append(dst, '\\', 'f')
			default:
				dst = append(dst, '\\', 'u', '0', '0', hex[s[i]>>4], hex[s[i]&0xF])
			}
			start = i + 1
		}
	}
	if start < len(s) {
		dst = append(dst, s[start:]...)
	}
	dst = append(dst, '"')
	return dst
}

// appendTime formats the time in RFC3339 format without using time.AppendFormat
// to avoid layout string parsing overhead.
func appendTime(dst []byte, t time.Time) []byte {
	year, month, day := t.Date()
	hour, min, sec := t.Clock()

	q := year / 100
	dst = append(dst, byte(q/10+'0'), byte(q%10+'0'))
	q = year % 100
	dst = append(dst, byte(q/10+'0'), byte(q%10+'0'), '-')
	m := int(month)
	dst = append(dst, byte(m/10+'0'), byte(m%10+'0'), '-')
	dst = append(dst, byte(day/10+'0'), byte(day%10+'0'), 'T')
	dst = append(dst, byte(hour/10+'0'), byte(hour%10+'0'), ':')
	dst = append(dst, byte(min/10+'0'), byte(min%10+'0'), ':')
	dst = append(dst, byte(sec/10+'0'), byte(sec%10+'0'))

	_, offset := t.Zone()
	if offset == 0 {
		return append(dst, 'Z')
	}

	if offset < 0 {
		dst = append(dst, '-')
		offset = -offset
	} else {
		dst = append(dst, '+')
	}

	offset /= 60
	dst = append(dst, byte(offset/60/10+'0'), byte(offset/60%10+'0'), ':')
	dst = append(dst, byte(offset%60/10+'0'), byte(offset%60%10+'0'))
	return dst
}