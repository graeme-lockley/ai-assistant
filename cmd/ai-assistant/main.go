package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/graemelockley/ai-assistant/internal/ask"
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
	case "ask":
		runAsk()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <server|repl|ask>\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  server  Run the HTTP server (streamed responses, session per client).\n")
	fmt.Fprintf(os.Stderr, "  repl    Run the REPL client (HTTP, streamed replies).\n")
	fmt.Fprintf(os.Stderr, "  ask     Run a single instruction and return JSON result.\n")
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

func runAsk() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s ask <instruction> [--model <model>]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  instruction  The instruction to send to the AI assistant.\n")
		fmt.Fprintf(os.Stderr, "  --model       Optional model to use (default: server default).\n")
		os.Exit(1)
	}

	cfg := config.AskFromEnv()
	var instruction string
	model := ""

	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		if args[i] == "--model" && i+1 < len(args) {
			model = args[i+1]
			i++
		} else {
			if instruction == "" {
				instruction = args[i]
			} else {
				instruction = instruction + " " + args[i]
			}
		}
	}

	if model != "" {
		cfg.Model = model
	}

	if instruction == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s ask <instruction> [--model <model>]\n", os.Args[0])
		os.Exit(1)
	}

	ctx := context.Background()
	result, err := ask.Run(ctx, cfg, instruction)
	if err != nil {
		errJSON, _ := json.Marshal(ask.ErrorResult{
			Error:   err.Error(),
			Details: "failed to execute ask request",
		})
		fmt.Println(string(errJSON))
		os.Exit(1)
	}

	output, err := json.Marshal(result)
	if err != nil {
		errJSON, _ := json.Marshal(ask.ErrorResult{
			Error:   err.Error(),
			Details: "failed to marshal result",
		})
		fmt.Println(string(errJSON))
		os.Exit(1)
	}
	fmt.Println(string(output))
}
