package scte35

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// newTable creates a new table with the given parameters.
func newTable(prefix, indent string) *table {
	return &table{
		prefix: prefix,
		indent: indent,
		b:      &strings.Builder{},
	}
}

// table simplifies construction of splice_info_section tables.
type table struct {
	prefix string
	indent string
	b      *strings.Builder
}

// row0 writes a new row with 0 indents.
func (t *table) row(indents int, key string, value any) {
	_, _ = t.b.WriteString(t.prefix)
	for i := 0; i < indents; i++ {
		_, _ = t.b.WriteString(t.indent)
	}
	_, _ = t.b.WriteString(key)
	if value != nil {
		_, _ = t.b.WriteString(": ")
		_, _ = t.b.WriteString(valueString(value))
	}
	_, _ = t.b.WriteRune('\n')
}

// String returns the table string.
func (t *table) String() string {
	return t.b.String()
}

// valueString converts the given value to a string
func valueString(value any) string {
	switch vt := value.(type) {
	case string:
		return vt
	case int:
		return strconv.FormatInt(int64(vt), 10)
	case int8:
		return strconv.FormatInt(int64(vt), 10)
	case int16:
		return strconv.FormatInt(int64(vt), 10)
	case int32:
		return strconv.FormatInt(int64(vt), 10)
	case int64:
		return strconv.FormatInt(vt, 10)
	case uint:
		return strconv.FormatUint(uint64(vt), 10)
	case uintptr:
		return strconv.FormatUint(uint64(vt), 10)
	case uint8:
		return strconv.FormatUint(uint64(vt), 10)
	case uint16:
		return strconv.FormatUint(uint64(vt), 10)
	case uint32:
		return strconv.FormatUint(uint64(vt), 10)
	case uint64:
		return strconv.FormatUint(vt, 10)
	case *uint64:
		if vt == nil {
			return ""
		}
		return strconv.FormatUint(*vt, 10)
	case bool:
		if vt {
			return "true"
		}
		return "false"
	case time.Time:
		return vt.Format(time.RFC3339)
	case float32:
		return fmt.Sprintf("%.2f", vt)
	case []byte:
		return fmt.Sprintf("%#02x", vt)
	default:
		return fmt.Sprintf("%s", vt)
	}
}
