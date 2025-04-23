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
		// Placeholder - create cmd.RunAdd later
		cmd.RunAdd(addCmd.Args())
		// fmt.Println("TODO: Implement add command for files:", addCmd.Args())

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
		// cmd.RunCommit(*msg, commitCmd.Args()) // Pass message and remaining non-flag args
		fmt.Println("TODO: Implement commit command with message:", *msg)

	// --- Placeholder for HashObject (useful for debugging) ---
	// case "hash-object":
	// 	hashCmd := flag.NewFlagSet("hash-object", flag.ExitOnError)
	// 	write := hashCmd.Bool("w", false, "Write the object to the object database")
	// 	hashCmd.Parse(args)
	// 	if hashCmd.NArg() != 1 {
	// 		fmt.Println("Usage: minigit hash-object [-w] <file>")
	// 		os.Exit(1)
	// 	}
	// 	filePath := hashCmd.Arg(0)
	// 	// Placeholder - create cmd.RunHashObject later
	//  // cmd.RunHashObject(filePath, *write)
	//  fmt.Println("TODO: Implement hash-object command for file:", filePath, "Write:", *write)

	default:
		fmt.Printf("Unknown command: %s\n", command)
		// TODO: Print available commands
		os.Exit(1)
	}
}
