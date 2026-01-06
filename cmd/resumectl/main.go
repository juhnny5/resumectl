package main

import "resumectl/internal/cli"

// Version is set at build time via ldflags
var version = "dev"

func main() {
	cli.SetVersion(version)
	cli.Execute()
}
