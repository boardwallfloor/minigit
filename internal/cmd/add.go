package cmd

import (
	"boardwallfloor/minigit/internal/index"
	"boardwallfloor/minigit/internal/object"
	"fmt"
	"log/slog"
	"os"
)

func RunAdd(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: minigit add <file>...")
		slog.Error("No files specified for add command.")
		os.Exit(1)
	}

	slog.Debug("Files to add:", "files", args)

	currState, err := index.ReadIndex()
	if err != nil {
		slog.Error("Failed to read index:", "error", err)
		os.Exit(1)
	}

	indexUpdateStatus := false
	for _, file := range args {
		slog.Debug("Processing file", "file", file)

		stat, err := os.Stat(file)
		if err != nil {
			if os.IsNotExist(err) {
				slog.Error("Cannot add file", "path", file, "reason", "not a regular file or does not exist")
			} else {
				slog.Error("Error checking file", "path", file, "reason", err)
			}
			continue
		}
		if stat.IsDir() {
			slog.Error("Cannot add directory", "path", file, "reason", "not a regular file")
			continue
		}

		content, err := os.ReadFile(file)
		if err != nil {
			slog.Error("Error reading file", "path", file, "reason", err)
			continue
		}

		hash, err := object.StoreBlob(content)
		if err != nil {
			slog.Error("Error storing blob", "path", file, "reason", err)
			continue
		}

		meta := index.IndexEntry{
			Mode: stat.Mode().Perm(),
			Hash: hash,
			Path: file,
		}

		entry, exists := currState.Entries[file]
		if !exists || entry.Hash != hash || entry.Mode != stat.Mode().Perm() {
			currState.Entries[file] = meta
			slog.Info("Added file to index", "file", file, "hash", hash)
			indexUpdateStatus = true
		} else {
			slog.Debug("No changes detected for file", "file", file)
		}

	}

	if indexUpdateStatus {
		err = index.WriteIndex(currState)
		if err != nil {
			slog.Error("Failed to write index:", "error", err)
			os.Exit(1)
		}
		slog.Info("Added files to index", "files", args)
	} else {
		slog.Info("No changes detected in index, nothing to add")
	}
}
