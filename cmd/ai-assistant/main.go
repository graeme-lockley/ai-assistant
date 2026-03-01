package main

import (
	"context"
	"fmt"
	"os"

	"github.com/graemelockley/ai-assistant/internal/config"
	"github.com/graemelockley/ai-assistant/internal/repl"
	"github.com/graemelockley/ai-assistant/internal/server"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "server":
		runServer()
	case "repl":
		runREPL()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <server|repl>\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  server  Run the HTTP server (streamed responses, session per client).\n")
	fmt.Fprintf(os.Stderr, "  repl    Run the REPL client (HTTP, streamed replies).\n")
}

func runServer() {
	cfg := config.ServerFromEnv()
	ctx := context.Background()
	if err := server.Run(ctx, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "server: %v\n", err)
		os.Exit(1)
	}
}

func runREPL() {
	cfg := config.REPLFromEnv()
	ctx := context.Background()
	if err := repl.Run(ctx, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "repl: %v\n", err)
		os.Exit(1)
	}
}
