package cli_test

import (
	"fmt"

	"github.com/Minimal-Viable-Software/cli-go"
)

type name string

func (n *name) Set(s string) error {
	*n = name(s)
	return nil
}

func (n *name) String() string {
	return string(*n)
}

func ExampleCommand_EnumArg() {
	app := cli.NewApplication()

	var player name
	app.EnumArg(&player, "name", "Choose your name", "Jack", "Will", "Barbarossa")

	app.Run(func() error {
		fmt.Printf("Hello, %s\n", player)
		return nil
	})

	app.Parse([]string{"Jack"})

	err := app.Parse([]string{})
	fmt.Println(err)

	err = app.Parse([]string{"Elizabeth"})
	fmt.Println(err)

	// Output:
	// Hello, Jack
	// missing required argument: name
	// invalid value "Elizabeth" for argument name: enum must be one of: Jack, Will, Barbarossa
}
