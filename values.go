package cli

import (
	"encoding"
	"flag"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"
)

// optional interface to indicate boolean flags that can be
// supplied without "=value" text
type boolFlag interface {
	flag.Value
	IsBoolFlag() bool
}

// -- bool Value
type boolValue bool

func newBoolValue(val bool, p *bool) *boolValue {
	*p = val
	return (*boolValue)(p)
}

func (b *boolValue) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		err = errParse
	}
	*b = boolValue(v)
	return err
}

func (b *boolValue) Get() any { return bool(*b) }

func (b *boolValue) String() string { return strconv.FormatBool(bool(*b)) }

func (b *boolValue) IsBoolFlag() bool { return true }

// -- int Value
type intValue int

func newIntValue(val int, p *int) *intValue {
	*p = val
	return (*intValue)(p)
}

func (i *intValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, strconv.IntSize)
	if err != nil {
		err = numError(err)
	}
	*i = intValue(v)
	return err
}

func (i *intValue) Get() any { return int(*i) }

func (i *intValue) String() string { return strconv.Itoa(int(*i)) }

// -- int64 Value
type int64Value int64

func newInt64Value(val int64, p *int64) *int64Value {
	*p = val
	return (*int64Value)(p)
}

func (i *int64Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		err = numError(err)
	}
	*i = int64Value(v)
	return err
}

func (i *int64Value) Get() any { return int64(*i) }

func (i *int64Value) String() string { return strconv.FormatInt(int64(*i), 10) }

// -- uint Value
type uintValue uint

func newUintValue(val uint, p *uint) *uintValue {
	*p = val
	return (*uintValue)(p)
}

func (i *uintValue) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, strconv.IntSize)
	if err != nil {
		err = numError(err)
	}
	*i = uintValue(v)
	return err
}

func (i *uintValue) Get() any { return uint(*i) }

func (i *uintValue) String() string { return strconv.FormatUint(uint64(*i), 10) }

// -- uint64 Value
type uint64Value uint64

func newUint64Value(val uint64, p *uint64) *uint64Value {
	*p = val
	return (*uint64Value)(p)
}

func (i *uint64Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 64)
	if err != nil {
		err = numError(err)
	}
	*i = uint64Value(v)
	return err
}

func (i *uint64Value) Get() any { return uint64(*i) }

func (i *uint64Value) String() string { return strconv.FormatUint(uint64(*i), 10) }

// -- string Value
type stringValue string

func newStringValue(val string, p *string) *stringValue {
	*p = val
	return (*stringValue)(p)
}

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

func (s *stringValue) Get() any { return string(*s) }

func (s *stringValue) String() string { return string(*s) }

// -- float64 Value
type float64Value float64

func newFloat64Value(val float64, p *float64) *float64Value {
	*p = val
	return (*float64Value)(p)
}

func (f *float64Value) Set(s string) error {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		err = numError(err)
	}
	*f = float64Value(v)
	return err
}

func (f *float64Value) Get() any { return float64(*f) }

func (f *float64Value) String() string { return strconv.FormatFloat(float64(*f), 'g', -1, 64) }

// -- time.Duration Value
type durationValue time.Duration

func newDurationValue(val time.Duration, p *time.Duration) *durationValue {
	*p = val
	return (*durationValue)(p)
}

func (d *durationValue) Set(s string) error {
	v, err := time.ParseDuration(s)
	if err != nil {
		err = errParse
	}
	*d = durationValue(v)
	return err
}

func (d *durationValue) Get() any { return time.Duration(*d) }

func (d *durationValue) String() string { return (*time.Duration)(d).String() }

// -- encoding.TextUnmarshaler Value
type textValue struct{ p encoding.TextUnmarshaler }

func newTextValue(val encoding.TextMarshaler, p encoding.TextUnmarshaler) textValue {
	ptrVal := reflect.ValueOf(p)
	if ptrVal.Kind() != reflect.Ptr {
		panic("variable value type must be a pointer")
	}
	defVal := reflect.ValueOf(val)
	if defVal.Kind() == reflect.Ptr {
		defVal = defVal.Elem()
	}
	if defVal.Type() != ptrVal.Type().Elem() {
		panic(fmt.Sprintf("default type does not match variable type: %v != %v", defVal.Type(), ptrVal.Type().Elem()))
	}
	ptrVal.Elem().Set(defVal)
	return textValue{p}
}

func (v textValue) Set(s string) error {
	return v.p.UnmarshalText([]byte(s))
}

func (v textValue) Get() any {
	return v.p
}

func (v textValue) String() string {
	if m, ok := v.p.(encoding.TextMarshaler); ok {
		if b, err := m.MarshalText(); err == nil {
			return string(b)
		}
	}
	return ""
}

// -- func Value
type funcValue func(string) error

func (f funcValue) Set(s string) error { return f(s) }

func (f funcValue) String() string { return "" }

// -- boolFunc Value
type boolFuncValue func(string) error

func (f boolFuncValue) Set(s string) error { return f(s) }

func (f boolFuncValue) String() string { return "" }

func (f boolFuncValue) IsBoolFlag() bool { return true }

// -- enum Value
type enumValue struct {
	val     flag.Value
	allowed []string
}

func newEnumValue(val flag.Value, allowed []string) *enumValue {
	return &enumValue{val: val, allowed: allowed}
}

func (e *enumValue) Set(s string) error {
	if slices.Contains(e.allowed, s) {
		e.val.Set(s)
		return nil
	}
	return fmt.Errorf("enum must be one of: %s", strings.Join(e.allowed, ", "))
}

func (e *enumValue) String() string {
	if e.val == nil {
		return ""
	}
	return e.val.String()
}

// -- string slice Value (for multi-value args)
type stringSliceValue struct{ p *[]string }

func (s *stringSliceValue) Set(val string) error {
	*s.p = append(*s.p, val)
	return nil
}

func (s *stringSliceValue) String() string {
	if s.p == nil {
		return ""
	}
	return fmt.Sprint(*s.p)
}

// -- int slice Value
type intSliceValue struct{ p *[]int }

func (s *intSliceValue) Set(val string) error {
	v, err := strconv.ParseInt(val, 0, strconv.IntSize)
	if err != nil {
		return numError(err)
	}
	*s.p = append(*s.p, int(v))
	return nil
}

func (s *intSliceValue) String() string {
	if s.p == nil {
		return ""
	}
	return fmt.Sprint(*s.p)
}

// -- float64 slice Value
type float64SliceValue struct{ p *[]float64 }

func (s *float64SliceValue) Set(val string) error {
	v, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return numError(err)
	}
	*s.p = append(*s.p, v)
	return nil
}

func (s *float64SliceValue) String() string {
	if s.p == nil {
		return ""
	}
	return fmt.Sprint(*s.p)
}

// -- bool slice Value
type boolSliceValue struct{ p *[]bool }

func (s *boolSliceValue) Set(val string) error {
	v, err := strconv.ParseBool(val)
	if err != nil {
		return errParse
	}
	*s.p = append(*s.p, v)
	return nil
}

func (s *boolSliceValue) String() string {
	if s.p == nil {
		return ""
	}
	return fmt.Sprint(*s.p)
}

// Convert a [strconv.NumError] to a `cli` error.
func numError(err error) error {
	ne, ok := err.(*strconv.NumError)
	if !ok {
		return err
	}
	if ne.Err == strconv.ErrSyntax {
		return errParse
	}
	if ne.Err == strconv.ErrRange {
		return errRange
	}
	return err
}
