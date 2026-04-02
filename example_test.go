package cli_test

import (
	"errors"
	"fmt"
	"os"
	"strings"

	cli "github.com/Minimal-Viable-Software/cli-go"
)

func Example_rootFlagsAndArgs() {
	root := cli.NewRoot()

	age := root.IntFlag("age", "Your age.", -1)
	name := root.StringArg(0, "name", "Your name")
	parents := root.StringArgs(1, 2, "parents", "Your parents names")
	siblings := root.StringRest("siblings", "Your siblings names")
	verbose := root.BoolFlag("verbose", "Be verbose about it")

	root.Run(func() error {
		fmt.Println("verbose", *verbose)
		fmt.Println("name=" + *name)
		fmt.Println("age=" + fmt.Sprint(*age))
		fmt.Println("parents=" + fmt.Sprintf("[%s]", strings.Join(*parents, " ")))
		fmt.Println("siblings=" + fmt.Sprintf("[%s]", strings.Join(*siblings, " ")))
		return nil
	})

	root.Parse([]string{"verbose", "age=4", "Stewie", "Peter", "Louis", "Meg", "Chris"})
	// Output:
	// verbose true
	// name=Stewie
	// age=4
	// parents=[Peter Louis]
	// siblings=[Meg Chris]
}

func Example_commandDispatch() {
	root := cli.NewRoot()
	cmd := root.SubCommand("prices", "List prices")
	sorting := cmd.EnumFlag("sort", "Sort by (default: price)", "price", "product")
	desc := cmd.BoolFlag("descending", "Sort descending instead of ascending")

	cmd.Run(func() error {
		fmt.Println("sort=" + *sorting)
		fmt.Println("descending", *desc)
		return nil
	})

	root.Parse([]string{"prices", "sort=product"})
	// Output:
	// sort=product
	// descending false
}

func Example_help() {
	root := cli.NewRoot()
	root.Output = os.Stdout
	root.Help = "Tell me about your self!"

	root.IntFlag("age", "Your age.", -1)
	root.StringArg(0, "name", "Your name")
	root.StringArgs(1, 2, "parents", "Your parents names")
	root.StringRest("siblings", "Your siblings names")
	root.BoolFlag("verbose", "Be verbose about it")

	cmd := root.SubCommand("prices", "List prices")
	cmd.EnumFlag("sort", "Sort by (default: price)", "price", "product")
	cmd.BoolFlag("descending", "Sort descending instead of ascending")

	err := root.Parse([]string{"help"})
	if errors.Is(err, cli.ErrHelp) {
		fmt.Println("Help requested")
	}
	// Output:
	// Tell me about your self!
	//
	// Options:
	//   age=<int>             Your age.
	//   verbose               Be verbose about it
	//   <string:name>         Your name
	//   <string:parents{2}>   Your parents names
	//   <string:siblings...>  Your siblings names
	//
	// Commands:
	//   prices                List prices
	//     descending          Sort descending instead of ascending
	//     sort=price|product  Sort by (default: price)
	// Help requested
}

func Example_helpCommand() {
	root := cli.NewRoot()
	root.Output = os.Stdout

	cmd := root.SubCommand("prices", "List prices")
	cmd.EnumFlag("sort", "Sort by (default: price)", "price", "product")
	cmd.BoolFlag("descending", "Sort descending instead of ascending")

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
	root := cli.NewRoot()
	root.BoolFlag("verbose", "Be verbose about it")
	name := root.StringArg(0, "name", "Your name")

	root.Run(func() error {
		fmt.Println("name=" + *name)
		return nil
	})

	root.Parse([]string{"--", "verbose"})
	// Output:
	// name=verbose
}

func Example_enumValidation() {
	root := cli.NewRoot()
	cmd := root.SubCommand("prices", "List prices")
	cmd.EnumFlag("sort", "Sort by (default: price)", "price", "product")

	err := root.Parse([]string{"prices", "sort=invalid"})
	fmt.Println(err)
	// Output:
	// invalid value "invalid" for flag sort: invalid value "invalid", must be one of: price, product
}

func Example_unknownFlag() {
	root := cli.NewRoot()
	root.StringArg(0, "name", "Your name")

	err := root.Parse([]string{"unknown=value"})
	fmt.Println(err)
	// Output:
	// unknown flag: unknown
}

func Example_missingRequiredArg() {
	root := cli.NewRoot()
	root.StringArg(0, "name", "Your name")

	err := root.Parse([]string{})
	fmt.Println(err)
	// Output:
	// missing required argument: name
}

func Example_restArgsEmpty() {
	root := cli.NewRoot()
	name := root.StringArg(0, "name", "Your name")
	rest := root.StringRest("rest", "The rest")

	root.Run(func() error {
		fmt.Println("name=" + *name)
		fmt.Println("rest=" + fmt.Sprintf("%v", *rest))
		return nil
	})

	root.Parse([]string{"hello"})
	// Output:
	// name=hello
	// rest=[]
}

func Example_flagsAnywhere() {
	root := cli.NewRoot()
	greeting := root.StringFlag("greeting", "A greeting", "")
	a := root.StringArg(0, "a", "First arg")
	b := root.StringArg(1, "b", "Second arg")

	root.Run(func() error {
		fmt.Println("a=" + *a)
		fmt.Println("b=" + *b)
		fmt.Println("greeting=" + *greeting)
		return nil
	})

	root.Parse([]string{"hello", "greeting=hi", "world"})
	// Output:
	// a=hello
	// b=world
	// greeting=hi
}

func Example_runFuncError() {
	root := cli.NewRoot()
	root.StringArg(0, "name", "Your name")

	root.Run(func() error {
		return errors.New("app error")
	})

	err := root.Parse([]string{"Alice"})
	fmt.Println(err)
	// Output:
	// app error
}

func Example_boolFlagExplicit() {
	root := cli.NewRoot()
	verbose := root.BoolFlag("verbose", "Be verbose about it")

	root.Run(func() error {
		fmt.Println("verbose", *verbose)
		return nil
	})

	root.Parse([]string{"verbose=false"})
	// Output:
	// verbose false
}
