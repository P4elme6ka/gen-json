package runtime

import (
	"encoding"
	"fmt"
	"math"
	"reflect"
	"strconv"
)

type UnknownFieldError struct {
	TypeName string
	Field    string
}

func (e UnknownFieldError) Error() string {
	return fmt.Sprintf("%s: unknown field %q", e.TypeName, e.Field)
}

type MissingFieldError struct {
	TypeName string
	Field    string
}

func (e MissingFieldError) Error() string {
	return fmt.Sprintf("%s: missing required field %q", e.TypeName, e.Field)
}

type UnsupportedFieldTypeError struct {
	TypeName string
	Field    string
	Op       string
}

func (e UnsupportedFieldTypeError) Error() string {
	return fmt.Sprintf("%s.%s: unsupported field type for %s", e.TypeName, e.Field, e.Op)
}

type Marshaler interface {
	MarshalJSON() ([]byte, error)
}

type Unmarshaler interface {
	UnmarshalJSON([]byte) error
}

type TextMarshaler interface {
	encoding.TextMarshaler
}

type TextUnmarshaler interface {
	encoding.TextUnmarshaler
}

func IsZero(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	return rv.IsZero()
}

type parser struct {
	data []byte
	i    int
}

func DecodeObject(data []byte, visit func(key []byte, value []byte) error) error {
	p := parser{data: data}
	p.skipWS()
	if !p.consume('{') {
		return p.errf("expected '{'")
	}

	p.skipWS()
	if p.consume('}') {
		p.skipWS()
		if p.i != len(p.data) {
			return p.errf("unexpected trailing data")
		}
		return nil
	}

	for {
		key, err := p.parseStringRaw()
		if err != nil {
			return err
		}
		p.skipWS()
		if !p.consume(':') {
			return p.errf("expected ':' after object key")
		}
		p.skipWS()

		start := p.i
		if err := p.skipValue(); err != nil {
			return err
		}

		if err := visit(key, p.data[start:p.i]); err != nil {
			return err
		}

		p.skipWS()
		if p.consume('}') {
			break
		}
		if !p.consume(',') {
			return p.errf("expected ',' or '}' in object")
		}
		p.skipWS()
	}

	p.skipWS()
	if p.i != len(p.data) {
		return p.errf("unexpected trailing data")
	}
	return nil
}

func (p *parser) skipValue() error {
	p.skipWS()
	if p.i >= len(p.data) {
		return p.errf("unexpected end of input")
	}

	switch p.data[p.i] {
	case '{':
		return p.skipObject()
	case '[':
		return p.skipArray()
	case '"':
		return p.skipStringRaw()
	case 't':
		return p.consumeLiteral("true")
	case 'f':
		return p.consumeLiteral("false")
	case 'n':
		return p.consumeLiteral("null")
	default:
		return p.skipNumber()
	}
}

func (p *parser) skipObject() error {
	if !p.consume('{') {
		return p.errf("expected '{'")
	}
	p.skipWS()
	if p.consume('}') {
		return nil
	}

	for {
		if _, err := p.parseString(); err != nil {
			return err
		}
		p.skipWS()
		if !p.consume(':') {
			return p.errf("expected ':' after object key")
		}
		if err := p.skipValue(); err != nil {
			return err
		}
		p.skipWS()
		if p.consume('}') {
			return nil
		}
		if !p.consume(',') {
			return p.errf("expected ',' or '}' in object")
		}
		p.skipWS()
	}
}

func (p *parser) skipArray() error {
	if !p.consume('[') {
		return p.errf("expected '['")
	}
	p.skipWS()
	if p.consume(']') {
		return nil
	}

	for {
		if err := p.skipValue(); err != nil {
			return err
		}
		p.skipWS()
		if p.consume(']') {
			return nil
		}
		if !p.consume(',') {
			return p.errf("expected ',' or ']' in array")
		}
		p.skipWS()
	}
}

func (p *parser) skipStringRaw() error {
	if !p.consume('"') {
		return p.errf("expected string")
	}
	for p.i < len(p.data) {
		c := p.data[p.i]
		p.i++
		if c == '\\' {
			if p.i >= len(p.data) {
				return p.errf("invalid escape sequence")
			}
			p.i++
			continue
		}
		if c == '"' {
			return nil
		}
	}
	return p.errf("unterminated string")
}

func (p *parser) parseString() (string, error) {
	raw, err := p.parseStringRaw()
	if err != nil {
		return "", err
	}
	v, err := strconv.Unquote(string(raw))
	if err != nil {
		return "", p.errf("invalid quoted string")
	}
	return v, nil
}

func (p *parser) parseStringRaw() ([]byte, error) {
	start := p.i
	if err := p.skipStringRaw(); err != nil {
		return nil, err
	}
	return p.data[start:p.i], nil
}

func KeyEq(raw []byte, want string) bool {
	// raw is a quoted JSON string slice (including quotes), want is unquoted.
	// Fast path: only supports keys without escape sequences.
	if len(raw) != len(want)+2 {
		return false
	}
	if raw[0] != '"' || raw[len(raw)-1] != '"' {
		return false
	}
	for i := 0; i < len(want); i++ {
		c := raw[i+1]
		if c == '\\' {
			return false
		}
		if c != want[i] {
			return false
		}
	}
	return true
}

func (p *parser) skipNumber() error {
	start := p.i
	if p.peek('-') {
		p.i++
	}
	if p.i >= len(p.data) {
		return p.errf("invalid number")
	}

	if p.peek('0') {
		p.i++
	} else {
		if !isDigit19(p.data[p.i]) {
			return p.errf("invalid number")
		}
		for p.i < len(p.data) && isDigit(p.data[p.i]) {
			p.i++
		}
	}

	if p.peek('.') {
		p.i++
		if p.i >= len(p.data) || !isDigit(p.data[p.i]) {
			return p.errf("invalid number")
		}
		for p.i < len(p.data) && isDigit(p.data[p.i]) {
			p.i++
		}
	}

	if p.peek('e') || p.peek('E') {
		p.i++
		if p.peek('+') || p.peek('-') {
			p.i++
		}
		if p.i >= len(p.data) || !isDigit(p.data[p.i]) {
			return p.errf("invalid number")
		}
		for p.i < len(p.data) && isDigit(p.data[p.i]) {
			p.i++
		}
	}

	if p.i == start {
		return p.errf("invalid number")
	}
	return nil
}

func (p *parser) consumeLiteral(lit string) error {
	if len(p.data)-p.i < len(lit) || string(p.data[p.i:p.i+len(lit)]) != lit {
		return p.errf("invalid literal")
	}
	p.i += len(lit)
	return nil
}

func (p *parser) skipWS() {
	for p.i < len(p.data) {
		switch p.data[p.i] {
		case ' ', '\n', '\t', '\r':
			p.i++
		default:
			return
		}
	}
}

func (p *parser) consume(ch byte) bool {
	if p.i < len(p.data) && p.data[p.i] == ch {
		p.i++
		return true
	}
	return false
}

func (p *parser) peek(ch byte) bool {
	return p.i < len(p.data) && p.data[p.i] == ch
}

func (p *parser) errf(msg string) error {
	return fmt.Errorf("json parse error at byte %d: %s", p.i, msg)
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func isDigit19(b byte) bool {
	return b >= '1' && b <= '9'
}

func IsNull(value []byte) bool {
	return len(value) == 4 && value[0] == 'n' && value[1] == 'u' && value[2] == 'l' && value[3] == 'l'
}

func DecodeString(value []byte) (string, error) {
	if len(value) < 2 || value[0] != '"' || value[len(value)-1] != '"' {
		return "", fmt.Errorf("expected JSON string")
	}
	inner := value[1 : len(value)-1]
	// Fast path: no escapes, treat as raw UTF-8 bytes.
	for i := 0; i < len(inner); i++ {
		if inner[i] == '\\' {
			v, err := strconv.Unquote(string(value))
			if err != nil {
				return "", fmt.Errorf("invalid JSON string: %w", err)
			}
			return v, nil
		}
	}
	return string(inner), nil
}

func DecodeText(value []byte) ([]byte, error) {
	// JSON string -> bytes for TextUnmarshaler
	s, err := DecodeString(value)
	if err != nil {
		return nil, err
	}
	return []byte(s), nil
}

func DecodeArray(data []byte, visit func(elem []byte) error) error {
	p := parser{data: data}
	p.skipWS()
	if !p.consume('[') {
		return p.errf("expected '['")
	}
	p.skipWS()
	if p.consume(']') {
		p.skipWS()
		if p.i != len(p.data) {
			return p.errf("unexpected trailing data")
		}
		return nil
	}
	for {
		p.skipWS()
		start := p.i
		if err := p.skipValue(); err != nil {
			return err
		}
		if err := visit(p.data[start:p.i]); err != nil {
			return err
		}
		p.skipWS()
		if p.consume(']') {
			break
		}
		if !p.consume(',') {
			return p.errf("expected ',' or ']' in array")
		}
	}
	p.skipWS()
	if p.i != len(p.data) {
		return p.errf("unexpected trailing data")
	}
	return nil
}

func DecodeMapObject(data []byte, visit func(key string, value []byte) error) error {
	// Similar to DecodeObject but returns decoded key strings (maps usually have arbitrary keys).
	p := parser{data: data}
	p.skipWS()
	if !p.consume('{') {
		return p.errf("expected '{'")
	}
	p.skipWS()
	if p.consume('}') {
		p.skipWS()
		if p.i != len(p.data) {
			return p.errf("unexpected trailing data")
		}
		return nil
	}
	for {
		key, err := p.parseString()
		if err != nil {
			return err
		}
		p.skipWS()
		if !p.consume(':') {
			return p.errf("expected ':' after object key")
		}
		p.skipWS()
		start := p.i
		if err := p.skipValue(); err != nil {
			return err
		}
		if err := visit(key, p.data[start:p.i]); err != nil {
			return err
		}
		p.skipWS()
		if p.consume('}') {
			break
		}
		if !p.consume(',') {
			return p.errf("expected ',' or '}' in object")
		}
		p.skipWS()
	}
	p.skipWS()
	if p.i != len(p.data) {
		return p.errf("unexpected trailing data")
	}
	return nil
}

func AppendCommaIfNeeded(dst []byte, wroteAny *bool) []byte {
	if *wroteAny {
		return append(dst, ',')
	}
	*wroteAny = true
	return dst
}

func AppendArrayStart(dst []byte, fieldWroteAny *bool) []byte {
	dst = AppendCommaIfNeeded(dst, fieldWroteAny)
	dst = append(dst, '[')
	return dst
}

func AppendArrayEnd(dst []byte) []byte { return append(dst, ']') }

func AppendObjectStart(dst []byte, fieldWroteAny *bool) []byte {
	dst = AppendCommaIfNeeded(dst, fieldWroteAny)
	dst = append(dst, '{')
	return dst
}

func AppendObjectEnd(dst []byte) []byte { return append(dst, '}') }

func DecodeBool(value []byte) (bool, error) {
	if len(value) == 4 && value[0] == 't' && value[1] == 'r' && value[2] == 'u' && value[3] == 'e' {
		return true, nil
	}
	if len(value) == 5 && value[0] == 'f' && value[1] == 'a' && value[2] == 'l' && value[3] == 's' && value[4] == 'e' {
		return false, nil
	}
	return false, fmt.Errorf("expected JSON boolean")
}

func DecodeInt(value []byte, bitSize int) (int64, error) {
	v, err := strconv.ParseInt(string(value), 10, bitSize)
	if err != nil {
		return 0, fmt.Errorf("invalid JSON integer: %w", err)
	}
	return v, nil
}

func DecodeUint(value []byte, bitSize int) (uint64, error) {
	v, err := strconv.ParseUint(string(value), 10, bitSize)
	if err != nil {
		return 0, fmt.Errorf("invalid JSON unsigned integer: %w", err)
	}
	return v, nil
}

func DecodeFloat(value []byte, bitSize int) (float64, error) {
	v, err := strconv.ParseFloat(string(value), bitSize)
	if err != nil {
		return 0, fmt.Errorf("invalid JSON float: %w", err)
	}
	return v, nil
}

func AppendFieldName(dst []byte, wroteAny *bool, name string) []byte {
	if *wroteAny {
		dst = append(dst, ',')
	} else {
		*wroteAny = true
	}
	dst = strconv.AppendQuote(dst, name)
	dst = append(dst, ':')
	return dst
}

// AppendFieldToken appends a precomputed JSON object field token of the form:
//
//	"name":
//
// The token MUST already be quoted and MUST end with ':'.
//
// This avoids per-field strconv.AppendQuote calls in hot encode paths.
func AppendFieldToken(dst []byte, wroteAny *bool, tok []byte) []byte {
	if *wroteAny {
		dst = append(dst, ',')
	} else {
		*wroteAny = true
	}
	return append(dst, tok...)
}

func AppendFloat64(dst []byte, v float64, bitSize int) ([]byte, error) {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return nil, fmt.Errorf("unsupported float value")
	}
	return strconv.AppendFloat(dst, v, 'g', -1, bitSize), nil
}
