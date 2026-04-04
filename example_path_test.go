package cli_test

import (
	"fmt"
	"os"

	cli "github.com/Minimal-Viable-Software/cli-go"
)

func ExamplePath_arg() {
	app := cli.NewApplication()

	cmd := app.SubCommand("read", "Read a file")
	var path cli.Path
	cmd.Arg(&path, "file", "the file path")

	cmd.Run(func() error {
		fmt.Printf("path=%s\n", path)
		return nil
	})

	app.Parse([]string{"read", "/tmp/data.txt"})

	// Output:
	// path=/tmp/data.txt
}

func ExamplePath_flag() {
	app := cli.NewApplication()

	cmd := app.SubCommand("write", "Write output")
	var path cli.Path
	cmd.Flag(&path, "output", "output path")

	cmd.Run(func() error {
		fmt.Printf("output=%s\n", path)
		return nil
	})

	app.Parse([]string{"write", "output=/tmp/out.txt"})

	// Output:
	// output=/tmp/out.txt
}

func ExamplePath_help() {
	app := cli.NewApplication()
	app.SetOutput(os.Stdout)

	cmd := app.SubCommand("read", "Read a file")
	var path cli.Path
	cmd.Arg(&path, "file", "the file path")

	app.Parse([]string{"help", "read"})

	// Output:
	// read           Read a file
	//   <path:file>  the file path
}
