/*
Package cli implements a minimal viable CLI library.

All flags are optional. All arguments are required.

Flags and argument functions follow the same patterns and support
the same types as the standard library `flag`, with additions.
*/
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// Sentinel errors.
var ErrHelp = errors.New("cli: help requested")

// Flag defines an optional flag.
type Flag struct {
	Name     string     // name as it appears on command line
	Usage    string     // help message
	Value    flag.Value // value as set
	DefValue string     // default value (as text); for usage message
}

// Argument defines a required argument.
type Argument struct {
	Name     string     // name as it appears on command line
	Usage    string     // help message
	Value    flag.Value // value as set
	Position int        // position index, assigned by declaration order
}

// RunFunc is the functions used to run commands.
type RunFunc func() error

// Application defines a root command of a CLI and is
// the entry point for everything in this package.
type Application struct {
	Command
	commands map[string]*Command
	Help     string
	output   io.Writer
}

// NewApplication is where everything starts.
func NewApplication() *Application {
	return &Application{
		Command: Command{
			flags:     make(map[string]*Flag),
			arguments: make(map[string]*Argument),
		},
		commands: make(map[string]*Command),
	}
}

// SubCommand registers a sub [Command] under the root.
//
// There is no programmatic nesting of subcommands.
// All subcommands is added using this function.
// Subcommands can be nested semantically using colons: `users`
// and `users:add` (the help message groups these together).
func (a *Application) SubCommand(name, usage string) *Command {
	cmd := &Command{
		Name:      name,
		Usage:     usage,
		flags:     make(map[string]*Flag),
		arguments: make(map[string]*Argument),
	}
	if _, exists := a.commands[name]; exists {
		panic(fmt.Sprintf("command already defined: %s", name))
	}
	a.commands[name] = cmd
	return cmd
}

// Parse the given arguments and execute the appropriate command.
func (a *Application) Parse(args []string) error {
	if len(args) > 0 && args[0] == "help" {
		return a.help(args[1:])
	}

	// Find command
	cmd := &a.Command
	if len(args) > 0 {
		if c, ok := a.commands[args[0]]; ok {
			cmd = c
			args = args[1:]
		}
	}

	// Flag parsing + positional collection
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
		if name, value, ok := strings.Cut(token, "="); ok {
			f, ok := cmd.flags[name]
			if !ok {
				return fmt.Errorf("unknown flag: %s", name)
			}
			if err := f.Value.Set(value); err != nil {
				return fmt.Errorf("invalid value %q for flag %q: %v", value, name, err)
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
			return fmt.Errorf("flag needs a value: %q", token)
		}

		// Otherwise it's a positional
		positionals = append(positionals, token)
	}

	// Assign positionals
	if err := assignPositionals(cmd, positionals); err != nil {
		return err
	}

	if cmd.run != nil {
		return cmd.run()
	}
	return nil
}

// Output returns the writer used for writing help text.
//
// Defaults to stderr.
func (a *Application) Output() io.Writer {
	if a.output != nil {
		return a.output
	}
	return os.Stderr
}

// SetOutput sets the writer used for writing help text.
func (a *Application) SetOutput(w io.Writer) {
	a.output = w
}

func (a *Application) help(args []string) error {
	var buf strings.Builder
	if len(args) == 0 {
		a.writeAppHelp(&buf)
	} else {
		cmdName := args[0]
		cmd, ok := a.commands[cmdName]
		if !ok {
			return fmt.Errorf("unknown command: %s", cmdName)
		}
		writeCommandHelp(&buf, cmdName, cmd, "")
	}
	fmt.Fprint(a.Output(), buf.String())
	return ErrHelp
}

func (a *Application) writeAppHelp(buf *strings.Builder) {
	if a.Help != "" {
		buf.WriteString(a.Help)
		buf.WriteString("\n")
	}

	if len(a.Command.flags) > 0 || len(a.Command.arguments) > 0 {
		buf.WriteString("\nOptions:\n")
		writeOptionsHelp(buf, &a.Command, "  ")
	}

	if len(a.commands) > 0 {
		buf.WriteString("\nCommands:\n")
		names := make([]string, 0, len(a.commands))
		for name := range a.commands {
			names = append(names, name)
		}
		sort.Strings(names)
		for i, name := range names {
			if i > 0 {
				buf.WriteString("\n")
			}
			writeCommandHelp(buf, name, a.commands[name], "  ")
		}
	}
}

// -----------------------------------------------------------------------------
// Helpers

func assignPositionals(cmd *Command, positionals []string) error {
	args := sortedArgs(cmd)
	for _, arg := range args {
		idx := arg.Position
		if idx >= len(positionals) {
			return fmt.Errorf("missing required argument: %s", arg.Name)
		}
		if err := arg.Value.Set(positionals[idx]); err != nil {
			return fmt.Errorf("invalid value %q for argument %s: %v", positionals[idx], arg.Name, err)
		}
	}
	if len(positionals) > len(args) {
		return fmt.Errorf("too many arguments: expected %d, got %d", len(args), len(positionals))
	}
	return nil
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
	case *Path:
		return "path"
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
	if ev, ok := a.Value.(*enumValue); ok {
		return "<" + strings.Join(ev.allowed, "|") + ":" + a.Name + ">"
	}
	tn := typeName(a.Value)
	if tn == "" {
		tn = "value"
	}
	return "<" + tn + ":" + a.Name + ">"
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
		return args[i].Position < args[j].Position
	})
	return args
}
