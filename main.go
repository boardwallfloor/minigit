package main

import (
	"boardwallfloor/minigit/internal/cmd"
	"flag"
	"fmt"
	"log/slog"
	"os"
)

type FileTree struct {
	root string
}

func main() {
	// Basic logger setup (can be more sophisticated later)
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	if len(os.Args) < 2 {
		fmt.Println("Usage: minigit <command> [options]")
		// TODO: Print available commands
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:] // Arguments specific to the subcommand

	switch command {
	case "init":
		// Define flags specific to 'init' if any (none needed for basic init)
		initCmd := flag.NewFlagSet("init", flag.ExitOnError)
		initCmd.Parse(args) // Parse the remaining args for this command
		// Ensure no extra arguments were passed for basic init
		if initCmd.NArg() > 0 {
			fmt.Println("Usage: minigit init")
			initCmd.PrintDefaults()
			os.Exit(1)
		}
		cmd.RunInit(initCmd.Args()) // Pass the (empty) remaining args

	case "add":
		addCmd := flag.NewFlagSet("add", flag.ExitOnError)
		// No specific flags for basic add yet
		addCmd.Parse(args)
		if addCmd.NArg() == 0 {
			fmt.Println("Usage: minigit add <file>...")
			addCmd.PrintDefaults()
			os.Exit(1)
		}
		cmd.RunAdd(addCmd.Args())

	case "commit":
		commitCmd := flag.NewFlagSet("commit", flag.ExitOnError)
		msg := commitCmd.String("m", "", "Commit message (required)")
		commitCmd.Parse(args)
		if *msg == "" {
			fmt.Println("Usage: minigit commit -m <message>")
			commitCmd.PrintDefaults()
			os.Exit(1)
		}
		// Placeholder - create cmd.RunCommit later
		cmd.RunCommit(*msg, commitCmd.Args()) // Pass message and remaining non-flag args

		// --- Add diff command ---
	case "diff":
		diffCmd := flag.NewFlagSet("diff", flag.ExitOnError)
		// Add --staged flag (or --cached)
		staged := diffCmd.Bool("staged", false, "Show diff between index and HEAD commit")
		// You could add other flags later if needed (e.g., specific files)

		diffCmd.Parse(args) // Parse the arguments specific to diff

		// Call RunDiff, passing remaining non-flag args (potential file paths)
		// and the boolean flag value
		cmd.RunDiff(diffCmd.Args(), *staged) // Need to create cmd.RunDiff

	case "log":
		logCmd := flag.NewFlagSet("log", flag.ExitOnError)
		// Add flags later if needed (e.g., -n)
		logCmd.Parse(args)
		if logCmd.NArg() > 0 {
			fmt.Println("Usage: minigit log")
			logCmd.PrintDefaults()
			os.Exit(1)
		}
		cmd.RunLog(logCmd.Args())

		// --- Add other commands like status later ---
	default:
		fmt.Printf("Unknown command: %s\n", command)
		// TODO: Print available commands
		os.Exit(1)
	}
}
