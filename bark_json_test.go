package bark

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestLoggerAllTypes(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(&buf)

	l.Info().
		// Strings and Bytes
		Str("str", "foo").
		Bytes("bytes", []byte("bar")).
		// Integers
		Int("int", -1).
		Int8("int8", -8).
		Int16("int16", -16).
		Int32("int32", -32).
		Int64("int64", -64).
		// Unsigned Integers
		Uint("uint", 1).
		Uint8("uint8", 8).
		Uint16("uint16", 16).
		Uint32("uint32", 32).
		Uint64("uint64", 64).
		Uintptr("uintptr", 128).
		// Floats
		Float32("float32", 1.23).
		Float64("float64", 4.56).
		// Complex
		Complex64("complex64", 1+2i).
		Complex128("complex128", 3+4i).
		// Bool & Error
		Bool("bool_true", true).
		Bool("bool_false", false).
		Error(errors.New("oops")).
		Msg("done")

	got := buf.String()
	expected := []string{
		`"level":"info"`,
		`"str":"foo"`,
		`"bytes":"` + base64.StdEncoding.EncodeToString([]byte("bar")) + `"`,
		`"int":-1`,
		`"int8":-8`,
		`"int16":-16`,
		`"int32":-32`,
		`"int64":-64`,
		`"uint":1`,
		`"uint8":8`,
		`"uint16":16`,
		`"uint32":32`,
		`"uint64":64`,
		`"uintptr":128`,
		`"float32":1.23`,
		`"float64":4.56`,
		`"complex64":"(1+2i)"`,
		`"complex128":"(3+4i)"`,
		`"bool_true":true`,
		`"bool_false":false`,
		`"error":"oops"`,
		`"message":"done"`,
	}

	for _, sub := range expected {
		if !strings.Contains(got, sub) {
			t.Errorf("missing %q in output: %s", sub, got)
		}
	}
}

func TestLoggerEdgeCases(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(&buf)

	// Test Error(nil) - should not log "error" key
	l.Info().Error(nil).Msg("no error")
	if strings.Contains(buf.String(), `"error"`) {
		t.Error("logged error key when error was nil")
	}

	buf.Reset()

	// Test Escaping
	// Needs to hit every branch in appendString
	specialChars := string([]byte{
		0x00, 0x1F, // Control chars
		'"', '\\', // Quotes and backslash
		'\n', '\r', '\t', '\b', '\f', // Common escapes
	})
	l.Info().Str("special", specialChars).Msg("esc")

	got := buf.String()

	if !strings.Contains(got, `\u0000`) {
		t.Error("missing escaped 0x00")
	}
	if !strings.Contains(got, `\u001f`) {
		t.Error("missing escaped 0x1F")
	}
	if !strings.Contains(got, `\"`) {
		t.Error("missing escaped quote")
	}
	if !strings.Contains(got, `\\`) {
		t.Error("missing escaped backslash")
	}
	if !strings.Contains(got, `\n`) {
		t.Error("missing escaped newline")
	}
}

func TestTimeFormatting(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(&buf)
	l.Info().Msg("time")

	if !strings.Contains(buf.String(), `"time":"20`) {
		t.Error("time field missing or malformed")
	}

	t1 := time.Date(2023, 10, 1, 12, 0, 0, 0, time.UTC)
	out := appendTime(nil, t1)
	if string(out) != "2023-10-01T12:00:00Z" {
		t.Errorf("UTC format mismatch: %s", string(out))
	}

	loc := time.FixedZone("EST", -5*60*60)
	t2 := time.Date(2023, 10, 1, 12, 0, 0, 0, loc)
	out2 := appendTime(nil, t2)

	if !strings.Contains(string(out2), "-05:00") {
		t.Errorf("Timezone offset missing or wrong: %s", string(out2))
	}
}

func TestLoggerConcurrency(t *testing.T) {
	l := NewLogger(io.Discard)
	var wg sync.WaitGroup
	for i := range 100 {
		wg.Go(func() {
			l.Info().Int("i", i).Msg("work")
		})
	}
	wg.Wait()
}