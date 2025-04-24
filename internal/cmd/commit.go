package cmd

import (
	"boardwallfloor/minigit/internal/index"
	"boardwallfloor/minigit/internal/object"
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func RunCommit(message string, args []string) {
	index, err := index.ReadIndex()
	if err != nil {
		slog.Error("Failed to read index", "error", err)
		return
	}

	rootHash, err := object.WriteTreeFromIndex(index)
	if err != nil {
		slog.Error("Failed to write tree from index", "error", err)
		return
	}

	content, err := os.ReadFile(filepath.Join(".minigit", "HEAD"))
	if err != nil {
		slog.Error("Failed to read HEAD", "error", err)
		return
	}
	cleanedPath := strings.TrimSpace(string(content))
	if !strings.HasPrefix(cleanedPath, "ref: ") {
		slog.Error("HEAD does not point to a reference", "content", cleanedPath)
		return
	}
	refTarget := strings.TrimPrefix(cleanedPath, "ref: ")
	prefixlessPath := strings.TrimPrefix(refTarget, "ref: ")
	fullPath := filepath.Join(".minigit", prefixlessPath)

	parentHash, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			parentHash = []byte("")
		} else {
			slog.Error("Failed to read parent hash", "error", err)
			return
		}
	}
	trimmedParentHash := strings.TrimSpace(string(parentHash))

	authorName := "Your Name"        // Replace later with config/env var
	authorEmail := "you@example.com" // Replace later
	now := time.Now()
	// Git's specific timestamp format: Unix timestamp + space + timezone offset (e.g., +0700)
	offset := now.Format("-0700") // Get timezone offset like -0700 or +0700
	timestamp := fmt.Sprintf("%d %s", now.Unix(), offset)
	authorInfo := fmt.Sprintf("%s <%s> %s", authorName, authorEmail, timestamp)
	committerInfo := authorInfo // Use same for committer for simplicity
	var commitContent bytes.Buffer
	fmt.Fprintf(&commitContent, "tree %s\n", rootHash)
	if trimmedParentHash != "" {
		fmt.Fprintf(&commitContent, "parent %s\n", trimmedParentHash)
	}
	fmt.Fprintf(&commitContent, "author %s\n", authorInfo)
	fmt.Fprintf(&commitContent, "committer %s\n", committerInfo)
	fmt.Fprintf(&commitContent, "\n")            // Blank line separator
	fmt.Fprintf(&commitContent, "%s\n", message) // User's commit message

	newCommitHash, err := object.StoreCommit(commitContent.Bytes())
	if err != nil {
		slog.Error("Failed to store commit", "error", err)
		return
	}
	slog.Debug("New commit hash", "hash", newCommitHash)

	slog.Debug("Attempting to update ref", "path", fullPath, "hash", newCommitHash)
	err = os.WriteFile(fullPath, []byte(newCommitHash+"\n"), 0644)
	if err != nil {
		slog.Error("Failed to write new commit hash to HEAD", "error", err)
		return
	}
	slog.Info("Successfully updated ref", "path", fullPath) // *** ADD THIS ***

	firstLine := strings.Split(message, "\n")[0]
	parentIndicator := ""
	if string(parentHash) == "" {
		parentIndicator = "(root-commit) "
	}
	fmt.Printf("[%s %s%s] %s\n", filepath.Base(prefixlessPath), parentIndicator, newCommitHash[:7], firstLine)
}
