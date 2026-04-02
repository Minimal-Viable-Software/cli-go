package cli

import (
	"encoding"
	"flag"
	"fmt"
	"strings"
	"time"
)

// Command represents a named command with its own set of flags and arguments.
type Command struct {
	Name       string
	Usage      string
	flags      map[string]*Flag
	arguments  map[string]*Argument
	run        RunFunc
	nextArgPos int
}

// Run sets the function to execute when this command is invoked.
func (c *Command) Run(fn RunFunc) {
	c.run = fn
}

// flag registers a flag.Value under the given name.
func (c *Command) flag(value flag.Value, name string, usage string) {
	if strings.Contains(name, "=") {
		panic(fmt.Sprintf("flag name %q contains '='", name))
	}
	defValue := value.String()
	f := &Flag{Name: name, Usage: usage, Value: value, DefValue: defValue}
	if _, exists := c.flags[name]; exists {
		panic(fmt.Sprintf("flag already defined: %s", name))
	}
	if c.flags == nil {
		c.flags = make(map[string]*Flag)
	}
	c.flags[name] = f
}

// BoolFlag defines a bool flag with specified name and usage string.
func (c *Command) BoolFlag(p *bool, name string, usage string) {
	c.flag((*boolValue)(p), name, usage)
}

// BoolFuncFlag defines a bool flag with specified name and usage string.
// The flag does not require a pointer; instead fn is called with "true" when the flag is set.
func (c *Command) BoolFuncFlag(name, usage string, fn func(string) error) {
	c.flag(boolFuncValue(fn), name, usage)
}

// IntFlag defines an int flag with specified name and usage string.
func (c *Command) IntFlag(p *int, name string, usage string) {
	c.flag((*intValue)(p), name, usage)
}

// Int64Flag defines an int64 flag with specified name and usage string.
func (c *Command) Int64Flag(p *int64, name string, usage string) {
	c.flag((*int64Value)(p), name, usage)
}

// UintFlag defines a uint flag with specified name and usage string.
func (c *Command) UintFlag(p *uint, name string, usage string) {
	c.flag((*uintValue)(p), name, usage)
}

// Uint64Flag defines a uint64 flag with specified name and usage string.
func (c *Command) Uint64Flag(p *uint64, name string, usage string) {
	c.flag((*uint64Value)(p), name, usage)
}

// Float64Flag defines a float64 flag with specified name and usage string.
func (c *Command) Float64Flag(p *float64, name string, usage string) {
	c.flag((*float64Value)(p), name, usage)
}

// StringFlag defines a string flag with specified name and usage string.
func (c *Command) StringFlag(p *string, name string, usage string) {
	c.flag((*stringValue)(p), name, usage)
}

// DurationFlag defines a time.Duration flag with specified name and usage string.
func (c *Command) DurationFlag(p *time.Duration, name string, usage string) {
	c.flag((*durationValue)(p), name, usage)
}

// TextFlag defines a flag with specified name and usage string for a value
// implementing encoding.TextUnmarshaler.
func (c *Command) TextFlag(p encoding.TextUnmarshaler, name string, usage string) {
	c.flag(textValue{p}, name, usage)
}

// Flag defines a flag with specified name and usage string for a custom flag.Value.
func (c *Command) Flag(value flag.Value, name string, usage string) {
	c.flag(value, name, usage)
}

// FuncFlag defines a flag with specified name and usage string.
// The flag does not require a pointer; instead fn is called with the flag's value.
func (c *Command) FuncFlag(name, usage string, fn func(string) error) {
	c.flag(funcValue(fn), name, usage)
}

// EnumFlag defines a flag restricted to a set of allowed values.
func (c *Command) EnumFlag(value flag.Value, name string, usage string, values ...string) {
	c.flag(newEnumValue(value, values), name, usage)
}

// arg registers a flag.Value as a positional argument.
func (c *Command) arg(value flag.Value, name string, usage string) {
	pos := c.nextArgPos
	c.nextArgPos++
	a := &Argument{Name: name, Usage: usage, Value: value, Position: pos}
	if _, exists := c.arguments[name]; exists {
		panic(fmt.Sprintf("argument already defined: %s", name))
	}
	if c.arguments == nil {
		c.arguments = make(map[string]*Argument)
	}
	c.arguments[name] = a
}

// StringArg defines a required string argument at the next position.
func (c *Command) StringArg(p *string, name, usage string) {
	c.arg((*stringValue)(p), name, usage)
}

// IntArg defines a required int argument at the next position.
func (c *Command) IntArg(p *int, name, usage string) {
	c.arg((*intValue)(p), name, usage)
}

// Int64Arg defines a required int64 argument at the next position.
func (c *Command) Int64Arg(p *int64, name, usage string) {
	c.arg((*int64Value)(p), name, usage)
}

// UintArg defines a required uint argument at the next position.
func (c *Command) UintArg(p *uint, name, usage string) {
	c.arg((*uintValue)(p), name, usage)
}

// Uint64Arg defines a required uint64 argument at the next position.
func (c *Command) Uint64Arg(p *uint64, name, usage string) {
	c.arg((*uint64Value)(p), name, usage)
}

// Float64Arg defines a required float64 argument at the next position.
func (c *Command) Float64Arg(p *float64, name, usage string) {
	c.arg((*float64Value)(p), name, usage)
}

// TextArg defines a required argument at the next position for a value
// implementing encoding.TextUnmarshaler.
func (c *Command) TextArg(p encoding.TextUnmarshaler, name, usage string) {
	c.arg(textValue{p}, name, usage)
}

// FuncArg defines a required argument at the next position.
// The argument does not require a pointer; instead fn is called with the argument's value.
func (c *Command) FuncArg(name, usage string, fn func(string) error) {
	c.arg(funcValue(fn), name, usage)
}

// Arg defines a required argument at the next position for a custom flag.Value.
func (c *Command) Arg(value flag.Value, name, usage string) {
	c.arg(value, name, usage)
}

// EnumArg defines a required argument at the next position, restricted to a set of allowed values.
func (c *Command) EnumArg(value flag.Value, name, usage string, values ...string) {
	c.arg(newEnumValue(value, values), name, usage)
}
