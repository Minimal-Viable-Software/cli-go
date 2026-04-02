package cli_test

import (
	"fmt"
	"net"

	"github.com/Minimal-Viable-Software/cli-go"
)

func ExampleCommand_TextArg() {
	app := cli.NewApplication()

	var ip net.IP
	app.TextArg(&ip, "ip", "Your IP address")

	app.Parse([]string{"127.0.0.1"})
	fmt.Printf("IP: %v\n", ip)

	// Output:
	// IP: 127.0.0.1
}
