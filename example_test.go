package cli_test

import (
	"errors"
	"fmt"
	"os"

	cli "github.com/Minimal-Viable-Software/cli-go"
)

func Example_rootFlagsAndArgs() {
	root := cli.NewApplication()

	age := -1
	root.IntFlag(&age, "age", "Your age.")
	var name string
	root.StringArg(&name, "name", "Your name")
	var verbose bool
	root.BoolFlag(&verbose, "verbose", "Be verbose about it")

	root.Run(func() error {
		fmt.Println("verbose", verbose)
		fmt.Println("name=" + name)
		fmt.Println("age=" + fmt.Sprint(age))
		return nil
	})

	root.Parse([]string{"verbose", "age=4", "Stewie"})
	// Output:
	// verbose true
	// name=Stewie
	// age=4
}

type sortstr string

func (s *sortstr) Set(str string) error {
	*s = sortstr(str)
	return nil
}

func (s *sortstr) String() string {
	return string(*s)
}

func Example_commandDispatch() {
	root := cli.NewApplication()
	cmd := root.SubCommand("prices", "List prices")
	var sorting sortstr
	cmd.EnumFlag(&sorting, "sort", "Sort by (default: price)", "price", "product")
	var desc bool
	cmd.BoolFlag(&desc, "descending", "Sort descending instead of ascending")

	cmd.Run(func() error {
		fmt.Println("sort=" + sorting)
		fmt.Println("descending", desc)
		return nil
	})

	root.Parse([]string{"prices", "sort=product"})
	// Output:
	// sort=product
	// descending false
}

func Example_help() {
	root := cli.NewApplication()
	root.SetOutput(os.Stdout)
	root.Help = "Tell me about your self!"

	age := -1
	root.IntFlag(&age, "age", "Your age.")
	var name string
	root.StringArg(&name, "name", "Your name")
	var verbose bool
	root.BoolFlag(&verbose, "verbose", "Be verbose about it")

	cmd := root.SubCommand("prices", "List prices")
	var sorting sortstr
	cmd.EnumFlag(&sorting, "sort", "Sort by (default: price)", "price", "product")
	var descending bool
	cmd.BoolFlag(&descending, "descending", "Sort descending instead of ascending")

	err := root.Parse([]string{"help"})
	if errors.Is(err, cli.ErrHelp) {
		fmt.Println("Help requested")
	}
	// Output:
	// Tell me about your self!
	//
	// Options:
	//   age=<int>      Your age.
	//   verbose        Be verbose about it
	//   <string:name>  Your name
	//
	// Commands:
	//   prices                List prices
	//     descending          Sort descending instead of ascending
	//     sort=price|product  Sort by (default: price)
	// Help requested
}

func Example_helpCommand() {
	root := cli.NewApplication()
	root.SetOutput(os.Stdout)

	cmd := root.SubCommand("prices", "List prices")
	var sorting sortstr
	cmd.EnumFlag(&sorting, "sort", "Sort by (default: price)", "price", "product")
	var descending bool
	cmd.BoolFlag(&descending, "descending", "Sort descending instead of ascending")

	err := root.Parse([]string{"help", "prices"})
	if errors.Is(err, cli.ErrHelp) {
		fmt.Println("Help requested")
	}
	// Output:
	// prices                List prices
	//   descending          Sort descending instead of ascending
	//   sort=price|product  Sort by (default: price)
	// Help requested
}

func Example_doubleHyphenTerminator() {
	root := cli.NewApplication()
	var verbose bool
	root.BoolFlag(&verbose, "verbose", "Be verbose about it")
	var name string
	root.StringArg(&name, "name", "Your name")

	root.Run(func() error {
		fmt.Println("name=" + name)
		return nil
	})

	root.Parse([]string{"--", "verbose"})
	// Output:
	// name=verbose
}

func Example_enumValidation() {
	root := cli.NewApplication()
	cmd := root.SubCommand("prices", "List prices")
	var sorting sortstr
	cmd.EnumFlag(&sorting, "sort", "Sort by (default: price)", "price", "product")

	err := root.Parse([]string{"prices", "sort=invalid"})
	fmt.Println(err)
	// Output:
	// invalid value "invalid" for flag "sort": enum must be one of: price, product
}

func Example_unknownFlag() {
	root := cli.NewApplication()
	var name string
	root.StringArg(&name, "name", "Your name")

	err := root.Parse([]string{"unknown=value"})
	fmt.Println(err)
	// Output:
	// unknown flag: unknown
}

func Example_missingRequiredArg() {
	root := cli.NewApplication()
	var name string
	root.StringArg(&name, "name", "Your name")

	err := root.Parse([]string{})
	fmt.Println(err)
	// Output:
	// missing required argument: name
}

func Example_extraArgs() {
	root := cli.NewApplication()
	var name string
	root.StringArg(&name, "name", "Your name")

	err := root.Parse([]string{"Alice", "extra"})
	fmt.Println(err)
	// Output:
	// too many arguments: expected 1, got 2
}

func Example_flagsAnywhere() {
	root := cli.NewApplication()
	var greeting string
	root.StringFlag(&greeting, "greeting", "A greeting")
	var a string
	root.StringArg(&a, "a", "First arg")
	var b string
	root.StringArg(&b, "b", "Second arg")

	root.Run(func() error {
		fmt.Println("a=" + a)
		fmt.Println("b=" + b)
		fmt.Println("greeting=" + greeting)
		return nil
	})

	root.Parse([]string{"hello", "greeting=hi", "world"})
	// Output:
	// a=hello
	// b=world
	// greeting=hi
}

func Example_runFuncError() {
	root := cli.NewApplication()
	var name string
	root.StringArg(&name, "name", "Your name")

	root.Run(func() error {
		return errors.New("app error")
	})

	err := root.Parse([]string{"Alice"})
	fmt.Println(err)
	// Output:
	// app error
}

func Example_boolFlagExplicit() {
	root := cli.NewApplication()
	var verbose bool
	root.BoolFlag(&verbose, "verbose", "Be verbose about it")

	root.Run(func() error {
		fmt.Println("verbose", verbose)
		return nil
	})

	root.Parse([]string{"verbose=false"})
	// Output:
	// verbose false
}

func Example_enumArg() {
	root := cli.NewApplication()
	root.SetOutput(os.Stdout)

	cmd := root.SubCommand("query", "Run a query")
	var sort sortstr
	cmd.EnumArg(&sort, "sort", "Sort field", "price", "name")

	cmd.Run(func() error {
		fmt.Println("sort=" + sort)
		return nil
	})

	root.Parse([]string{"query", "price"})

	// Show help too
	root.Parse([]string{"help", "query"})
	// Output:
	// sort=price
	// query                Run a query
	//   <price|name:sort>  Sort field
}
