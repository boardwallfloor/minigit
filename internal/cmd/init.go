package cmd

import (
	"fmt"
	"log/slog"
	"os"
)

func RunInit(args []string) {
	if _, err := os.Stat(".minigit"); err == nil {
		slog.Error("Already a minigit repository.")
		os.Exit(1)
	} else if !os.IsNotExist(err) {
		slog.Error("Failed to check for .minigit directory", "error", err)
		os.Exit(1)
	}

	slog.Info("Initializing new minigit repository...")
	err := os.Mkdir(".minigit", 0755)
	if err != nil {
		slog.Error("Failed to create .minigit directory", "error", err)
		os.Exit(1)
	}

	err = os.MkdirAll(".minigit/objects", 0755)
	if err != nil {
		slog.Error("Failed to create objects directory", "error", err)
		os.Exit(1)
	}
	err = os.MkdirAll(".minigit/refs/heads", 0755)
	if err != nil {
		slog.Error("Failed to create heads directory", "error", err)
		os.Exit(1)
	}
	err = os.WriteFile(".minigit/HEAD", []byte("ref: refs/heads/main\n"), 0644)
	if err != nil {
		slog.Error("Failed to create HEAD", "error", err)
		os.Exit(1)
	}

	fmt.Println("Initialized empty minigit repository in ./.minigit")
}
