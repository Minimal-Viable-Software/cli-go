package cli

import (
	"encoding"
	"flag"
	"fmt"
	"strings"
	"time"
)

type Command struct {
	Name      string
	Usage     string
	flags     map[string]*Flag
	arguments map[string]*Argument
	run       RunFunc
}

// Run sets the function to execute when this command is invoked.
func (c *Command) Run(fn RunFunc) {
	c.run = fn
}

// varFlag registers a flag.Value under the given name.
func (c *Command) varFlag(value flag.Value, name string, usage string) {
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

func (c *Command) BoolFlag(p *bool, name string, usage string, value bool) {
	c.varFlag(newBoolValue(value, p), name, usage)
}

func (c *Command) BoolFuncFlag(name, usage string, fn func(string) error) {
	c.varFlag(boolFuncValue(fn), name, usage)
}

func (c *Command) IntFlag(p *int, name string, usage string, value int) {
	c.varFlag(newIntValue(value, p), name, usage)
}

func (c *Command) Int64Flag(p *int64, name string, usage string, value int64) {
	c.varFlag(newInt64Value(value, p), name, usage)
}

func (c *Command) UintFlag(p *uint, name string, usage string, value uint) {
	c.varFlag(newUintValue(value, p), name, usage)
}

func (c *Command) Uint64Flag(p *uint64, name string, usage string, value uint64) {
	c.varFlag(newUint64Value(value, p), name, usage)
}

func (c *Command) Float64Flag(p *float64, name string, usage string, value float64) {
	c.varFlag(newFloat64Value(value, p), name, usage)
}

func (c *Command) StringFlag(p *string, name string, usage string, value string) {
	c.varFlag(newStringValue(value, p), name, usage)
}

func (c *Command) DurationFlag(p *time.Duration, name string, usage string, value time.Duration) {
	c.varFlag(newDurationValue(value, p), name, usage)
}

func (c *Command) TextFlag(p encoding.TextUnmarshaler, name string, usage string, value encoding.TextMarshaler) {
	c.varFlag(newTextValue(value, p), name, usage)
}

func (c *Command) VarFlag(value flag.Value, name string, usage string) {
	c.varFlag(value, name, usage)
}

func (c *Command) FuncFlag(name, usage string, fn func(string) error) {
	c.varFlag(funcValue(fn), name, usage)
}

func (c *Command) EnumFlag(p *string, name string, usage string, values ...string) {
	c.varFlag(newEnumValue(p, values), name, usage)
}

// varArg registers a flag.Value as a positional argument.
func (c *Command) varArg(value flag.Value, position []int, name string, usage string) {
	a := &Argument{Name: name, Usage: usage, Value: value, Position: position}
	if _, exists := c.arguments[name]; exists {
		panic(fmt.Sprintf("argument already defined: %s", name))
	}
	// Check for multiple rest args
	if len(position) == 0 {
		for _, existing := range c.arguments {
			if len(existing.Position) == 0 {
				panic(fmt.Sprintf("multiple rest arguments defined: %s and %s", existing.Name, name))
			}
		}
	}
	if c.arguments == nil {
		c.arguments = make(map[string]*Argument)
	}
	c.arguments[name] = a
}

// Single-position argument methods

func (c *Command) StringArg(p *string, position int, name, usage string) {
	c.varArg(newStringValue("", p), []int{position}, name, usage)
}

func (c *Command) IntArg(p *int, position int, name, usage string) {
	c.varArg(newIntValue(0, p), []int{position}, name, usage)
}

func (c *Command) Float64Arg(p *float64, position int, name, usage string) {
	c.varArg(newFloat64Value(0, p), []int{position}, name, usage)
}

func (c *Command) BoolArg(p *bool, position int, name, usage string) {
	c.varArg(newBoolValue(false, p), []int{position}, name, usage)
}

// Range argument methods — inclusive [start, end]

func (c *Command) StringArgs(p *[]string, start, end int, name, usage string) {
	c.varArg(&stringSliceValue{p: p}, []int{start, end}, name, usage)
}

func (c *Command) IntArgs(p *[]int, start, end int, name, usage string) {
	c.varArg(&intSliceValue{p: p}, []int{start, end}, name, usage)
}

func (c *Command) Float64Args(p *[]float64, start, end int, name, usage string) {
	c.varArg(&float64SliceValue{p: p}, []int{start, end}, name, usage)
}

func (c *Command) BoolArgs(p *[]bool, start, end int, name, usage string) {
	c.varArg(&boolSliceValue{p: p}, []int{start, end}, name, usage)
}

// Rest argument methods — consume all remaining positional args

func (c *Command) StringRest(p *[]string, name, usage string) {
	c.varArg(&stringSliceValue{p: p}, []int{}, name, usage)
}

func (c *Command) IntRest(p *[]int, name, usage string) {
	c.varArg(&intSliceValue{p: p}, []int{}, name, usage)
}

func (c *Command) Float64Rest(p *[]float64, name, usage string) {
	c.varArg(&float64SliceValue{p: p}, []int{}, name, usage)
}

func (c *Command) BoolRest(p *[]bool, name, usage string) {
	c.varArg(&boolSliceValue{p: p}, []int{}, name, usage)
}
