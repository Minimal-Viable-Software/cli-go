/*
Package cli implements a minimal viable CLI library.

All flags are optional. All arguments are required.

Flags and argument functions follow the same patterns and support
the same types as the standard library `flag`, with additions.
*/
package cli

import (
	"encoding"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"
)

// Sentinel errors.
var ErrHelp = errors.New("cli: help requested")
var errParse = errors.New("parse error")
var errRange = errors.New("value out of range")

type Flag struct {
	Name     string     // name as it appears on command line
	Usage    string     // help message
	Value    flag.Value // value as set
	DefValue string     // default value (as text); for usage message
}

type Argument struct {
	Name     string     // name as it appears on command line
	Usage    string     // help message
	Value    flag.Value // value as set
	Position []int      // one integer is abs position, two integers is a range, empty is rest
}

type RunFunc func() error

type Command struct {
	Name      string
	Usage     string
	flags     map[string]*Flag
	arguments map[string]*Argument
	run       RunFunc
}

type Root struct {
	Command
	commands map[string]*Command
	Help     string
	Output   io.Writer
}

// NewRoot creates a new Root command.
func NewRoot() *Root {
	return &Root{
		Command: Command{
			flags:     make(map[string]*Flag),
			arguments: make(map[string]*Argument),
		},
		commands: make(map[string]*Command),
	}
}

// Run sets the function to execute when this command is invoked.
func (c *Command) Run(fn RunFunc) {
	c.run = fn
}

// SubCommand registers a subcommand under the root.
func (r *Root) SubCommand(name, usage string) *Command {
	cmd := &Command{
		Name:      name,
		Usage:     usage,
		flags:     make(map[string]*Flag),
		arguments: make(map[string]*Argument),
	}
	if r.commands == nil {
		r.commands = make(map[string]*Command)
	}
	if _, exists := r.commands[name]; exists {
		panic(fmt.Sprintf("command already defined: %s", name))
	}
	r.commands[name] = cmd
	return cmd
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

// Bool — default always false
func (c *Command) BoolFlag(name string, usage string) *bool {
	p := new(bool)
	c.BoolVarFlag(p, name, usage, false)
	return p
}

func (c *Command) BoolVarFlag(p *bool, name string, usage string, value bool) {
	c.varFlag(newBoolValue(value, p), name, usage)
}

func (c *Command) BoolFuncFlag(name, usage string, fn func(string) error) {
	c.varFlag(boolFuncValue(fn), name, usage)
}

// Int
func (c *Command) IntFlag(name string, usage string, value int) *int {
	p := new(int)
	c.IntVarFlag(p, name, usage, value)
	return p
}

func (c *Command) IntVarFlag(p *int, name string, usage string, value int) {
	c.varFlag(newIntValue(value, p), name, usage)
}

// Int64
func (c *Command) Int64Flag(name string, usage string, value int64) *int64 {
	p := new(int64)
	c.Int64VarFlag(p, name, usage, value)
	return p
}

func (c *Command) Int64VarFlag(p *int64, name string, usage string, value int64) {
	c.varFlag(newInt64Value(value, p), name, usage)
}

// Uint
func (c *Command) UintFlag(name string, usage string, value uint) *uint {
	p := new(uint)
	c.UintVarFlag(p, name, usage, value)
	return p
}

func (c *Command) UintVarFlag(p *uint, name string, usage string, value uint) {
	c.varFlag(newUintValue(value, p), name, usage)
}

// Uint64
func (c *Command) Uint64Flag(name string, usage string, value uint64) *uint64 {
	p := new(uint64)
	c.Uint64VarFlag(p, name, usage, value)
	return p
}

func (c *Command) Uint64VarFlag(p *uint64, name string, usage string, value uint64) {
	c.varFlag(newUint64Value(value, p), name, usage)
}

// Float64
func (c *Command) Float64Flag(name string, usage string, value float64) *float64 {
	p := new(float64)
	c.Float64VarFlag(p, name, usage, value)
	return p
}

func (c *Command) Float64VarFlag(p *float64, name string, usage string, value float64) {
	c.varFlag(newFloat64Value(value, p), name, usage)
}

// String
func (c *Command) StringFlag(name string, usage string, value string) *string {
	p := new(string)
	c.StringVarFlag(p, name, usage, value)
	return p
}

func (c *Command) StringVarFlag(p *string, name string, usage string, value string) {
	c.varFlag(newStringValue(value, p), name, usage)
}

// Duration
func (c *Command) DurationFlag(name string, usage string, value time.Duration) *time.Duration {
	p := new(time.Duration)
	c.DurationVarFlag(p, name, usage, value)
	return p
}

func (c *Command) DurationVarFlag(p *time.Duration, name string, usage string, value time.Duration) {
	c.varFlag(newDurationValue(value, p), name, usage)
}

// TextVar
func (c *Command) TextVarFlag(p encoding.TextUnmarshaler, name string, usage string, value encoding.TextMarshaler) {
	c.varFlag(newTextValue(value, p), name, usage)
}

// Var — generic, user provides flag.Value
func (c *Command) VarFlag(value flag.Value, name string, usage string) {
	c.varFlag(value, name, usage)
}

// Func
func (c *Command) FuncFlag(name, usage string, fn func(string) error) {
	c.varFlag(funcValue(fn), name, usage)
}

// Enum — first value is default
func (c *Command) EnumFlag(name string, usage string, values ...string) *string {
	p := new(string)
	c.varFlag(newEnumValue(p, values), name, usage)
	return p
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

func (c *Command) StringArg(position int, name, usage string) *string {
	p := new(string)
	c.varArg(newStringValue("", p), []int{position}, name, usage)
	return p
}

func (c *Command) IntArg(position int, name, usage string) *int {
	p := new(int)
	c.varArg(newIntValue(0, p), []int{position}, name, usage)
	return p
}

func (c *Command) Float64Arg(position int, name, usage string) *float64 {
	p := new(float64)
	c.varArg(newFloat64Value(0, p), []int{position}, name, usage)
	return p
}

func (c *Command) BoolArg(position int, name, usage string) *bool {
	p := new(bool)
	c.varArg(newBoolValue(false, p), []int{position}, name, usage)
	return p
}

// Range argument methods — inclusive [start, end]

func (c *Command) StringArgs(start, end int, name, usage string) *[]string {
	p := &[]string{}
	c.varArg(&stringSliceValue{p: p}, []int{start, end}, name, usage)
	return p
}

func (c *Command) IntArgs(start, end int, name, usage string) *[]int {
	p := &[]int{}
	c.varArg(&intSliceValue{p: p}, []int{start, end}, name, usage)
	return p
}

func (c *Command) Float64Args(start, end int, name, usage string) *[]float64 {
	p := &[]float64{}
	c.varArg(&float64SliceValue{p: p}, []int{start, end}, name, usage)
	return p
}

func (c *Command) BoolArgs(start, end int, name, usage string) *[]bool {
	p := &[]bool{}
	c.varArg(&boolSliceValue{p: p}, []int{start, end}, name, usage)
	return p
}

// Rest argument methods — consume all remaining positional args

func (c *Command) StringRest(name, usage string) *[]string {
	p := &[]string{}
	c.varArg(&stringSliceValue{p: p}, []int{}, name, usage)
	return p
}

func (c *Command) IntRest(name, usage string) *[]int {
	p := &[]int{}
	c.varArg(&intSliceValue{p: p}, []int{}, name, usage)
	return p
}

func (c *Command) Float64Rest(name, usage string) *[]float64 {
	p := &[]float64{}
	c.varArg(&float64SliceValue{p: p}, []int{}, name, usage)
	return p
}

func (c *Command) BoolRest(name, usage string) *[]bool {
	p := &[]bool{}
	c.varArg(&boolSliceValue{p: p}, []int{}, name, usage)
	return p
}

// Parse parses the given arguments, dispatches to the appropriate command,
// and runs its RunFunc if set.
func (r *Root) Parse(args []string) error {
	// Step 1 — Help check
	if len(args) > 0 && args[0] == "help" {
		return r.help(args[1:])
	}

	// Step 2 — Command dispatch
	cmd := &r.Command
	if len(args) > 0 {
		if c, ok := r.commands[args[0]]; ok {
			cmd = c
			args = args[1:]
		}
	}

	// Step 3 — Flag parsing + positional collection
	flagsDone := false
	var positionals []string
	for _, token := range args {
		if token == "--" {
			flagsDone = true
			continue
		}
		if flagsDone {
			positionals = append(positionals, token)
			continue
		}

		// Check for name=value
		if eqIdx := strings.IndexByte(token, '='); eqIdx >= 0 {
			name := token[:eqIdx]
			value := token[eqIdx+1:]
			f, ok := cmd.flags[name]
			if !ok {
				return fmt.Errorf("unknown flag: %s", name)
			}
			if err := f.Value.Set(value); err != nil {
				return fmt.Errorf("invalid value %q for flag %s: %v", value, name, err)
			}
			continue
		}

		// Check for bare bool flag
		if f, ok := cmd.flags[token]; ok {
			if bf, isBool := f.Value.(boolFlag); isBool && bf.IsBoolFlag() {
				if err := f.Value.Set("true"); err != nil {
					return fmt.Errorf("invalid boolean flag %s: %v", token, err)
				}
				continue
			}
			// Non-bool flag used without =value
			return fmt.Errorf("flag needs a value: %s=<value>", token)
		}

		// Otherwise it's a positional
		positionals = append(positionals, token)
	}

	// Step 4 — Assign positionals
	if err := assignPositionals(cmd, positionals); err != nil {
		return err
	}

	// Step 5 — Run
	if cmd.run != nil {
		return cmd.run()
	}
	return nil
}

func assignPositionals(cmd *Command, positionals []string) error {
	// Compute maxFixedPos from all arguments with explicit positions
	maxFixedPos := -1
	for _, arg := range cmd.arguments {
		switch len(arg.Position) {
		case 1: // single
			if arg.Position[0] > maxFixedPos {
				maxFixedPos = arg.Position[0]
			}
		case 2: // range [start, end] inclusive
			if arg.Position[1] > maxFixedPos {
				maxFixedPos = arg.Position[1]
			}
		}
	}

	// Assign single-position args
	for _, arg := range cmd.arguments {
		if len(arg.Position) == 1 {
			idx := arg.Position[0]
			if idx >= len(positionals) {
				return fmt.Errorf("missing required argument: %s", arg.Name)
			}
			if err := arg.Value.Set(positionals[idx]); err != nil {
				return fmt.Errorf("invalid value %q for argument %s: %v", positionals[idx], arg.Name, err)
			}
		}
	}

	// Assign range args
	for _, arg := range cmd.arguments {
		if len(arg.Position) == 2 {
			start, end := arg.Position[0], arg.Position[1]
			for i := start; i <= end; i++ {
				if i >= len(positionals) {
					return fmt.Errorf("missing required argument: %s", arg.Name)
				}
				if err := arg.Value.Set(positionals[i]); err != nil {
					return fmt.Errorf("invalid value %q for argument %s: %v", positionals[i], arg.Name, err)
				}
			}
		}
	}

	// Assign rest args (Position is empty slice)
	for _, arg := range cmd.arguments {
		if len(arg.Position) == 0 {
			restStart := maxFixedPos + 1
			for i := restStart; i < len(positionals); i++ {
				if err := arg.Value.Set(positionals[i]); err != nil {
					return fmt.Errorf("invalid value %q for argument %s: %v", positionals[i], arg.Name, err)
				}
			}
			// Rest with zero values is OK — no error
		}
	}

	return nil
}

func (r *Root) output() io.Writer {
	if r.Output != nil {
		return r.Output
	}
	return os.Stderr
}

func typeName(v flag.Value) string {
	switch v.(type) {
	case *boolValue, boolFuncValue:
		return ""
	case *intValue:
		return "int"
	case *int64Value:
		return "int64"
	case *uintValue:
		return "uint"
	case *uint64Value:
		return "uint64"
	case *stringValue:
		return "string"
	case *float64Value:
		return "float64"
	case *durationValue:
		return "duration"
	case textValue:
		return "text"
	case *enumValue:
		return ""
	case funcValue:
		return "string"
	default:
		return "value"
	}
}

func flagDisplayName(f *Flag) string {
	if ev, ok := f.Value.(*enumValue); ok {
		return f.Name + "=" + strings.Join(ev.allowed, "|")
	}
	tn := typeName(f.Value)
	if tn == "" {
		return f.Name
	}
	return f.Name + "=<" + tn + ">"
}

func argDisplayName(a *Argument) string {
	tn := typeName(a.Value)
	if tn == "" {
		tn = "value"
	}
	switch a.Value.(type) {
	case *stringSliceValue:
		tn = "string"
	case *intSliceValue:
		tn = "int"
	case *float64SliceValue:
		tn = "float64"
	case *boolSliceValue:
		tn = "bool"
	}

	switch len(a.Position) {
	case 1:
		return "<" + tn + ":" + a.Name + ">"
	case 2:
		count := a.Position[1] - a.Position[0] + 1
		return fmt.Sprintf("<%s:%s{%d}>", tn, a.Name, count)
	default:
		return "<" + tn + ":" + a.Name + "...>"
	}
}

func (r *Root) help(args []string) error {
	var buf strings.Builder
	if len(args) == 0 {
		r.writeRootHelp(&buf)
	} else {
		cmdName := args[0]
		cmd, ok := r.commands[cmdName]
		if !ok {
			return fmt.Errorf("unknown command: %s", cmdName)
		}
		writeCommandHelp(&buf, cmdName, cmd, "")
	}
	fmt.Fprint(r.output(), buf.String())
	return ErrHelp
}

func (r *Root) writeRootHelp(buf *strings.Builder) {
	if r.Help != "" {
		buf.WriteString(r.Help)
		buf.WriteString("\n")
	}

	if len(r.Command.flags) > 0 || len(r.Command.arguments) > 0 {
		buf.WriteString("\nOptions:\n")
		writeOptionsHelp(buf, &r.Command, "  ")
	}

	if len(r.commands) > 0 {
		buf.WriteString("\nCommands:\n")
		names := make([]string, 0, len(r.commands))
		for name := range r.commands {
			names = append(names, name)
		}
		sort.Strings(names)
		for i, name := range names {
			if i > 0 {
				buf.WriteString("\n")
			}
			writeCommandHelp(buf, name, r.commands[name], "  ")
		}
	}
}

func writeCommandHelp(buf *strings.Builder, name string, cmd *Command, indent string) {
	type entry struct {
		display string
		usage   string
	}
	var entries []entry

	entries = append(entries, entry{indent + name, cmd.Usage})

	flagNames := make([]string, 0, len(cmd.flags))
	for n := range cmd.flags {
		flagNames = append(flagNames, n)
	}
	sort.Strings(flagNames)
	for _, n := range flagNames {
		f := cmd.flags[n]
		entries = append(entries, entry{indent + "  " + flagDisplayName(f), f.Usage})
	}

	args := sortedArgs(cmd)
	for _, a := range args {
		entries = append(entries, entry{indent + "  " + argDisplayName(a), a.Usage})
	}

	maxWidth := 0
	for _, e := range entries {
		if len(e.display) > maxWidth {
			maxWidth = len(e.display)
		}
	}

	for _, e := range entries {
		fmt.Fprintf(buf, "%-*s  %s\n", maxWidth, e.display, e.usage)
	}
}

func writeOptionsHelp(buf *strings.Builder, cmd *Command, indent string) {
	type entry struct {
		display string
		usage   string
	}
	var entries []entry

	flagNames := make([]string, 0, len(cmd.flags))
	for n := range cmd.flags {
		flagNames = append(flagNames, n)
	}
	sort.Strings(flagNames)
	for _, n := range flagNames {
		f := cmd.flags[n]
		entries = append(entries, entry{indent + flagDisplayName(f), f.Usage})
	}

	args := sortedArgs(cmd)
	for _, a := range args {
		entries = append(entries, entry{indent + argDisplayName(a), a.Usage})
	}

	maxWidth := 0
	for _, e := range entries {
		if len(e.display) > maxWidth {
			maxWidth = len(e.display)
		}
	}
	for _, e := range entries {
		fmt.Fprintf(buf, "%-*s  %s\n", maxWidth, e.display, e.usage)
	}
}

func sortedArgs(cmd *Command) []*Argument {
	args := make([]*Argument, 0, len(cmd.arguments))
	for _, a := range cmd.arguments {
		args = append(args, a)
	}
	sort.Slice(args, func(i, j int) bool {
		pi, pj := args[i].Position, args[j].Position
		if len(pi) == 0 {
			return false
		}
		if len(pj) == 0 {
			return true
		}
		return pi[0] < pj[0]
	})
	return args
}
