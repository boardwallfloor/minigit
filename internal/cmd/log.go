package cmd

import (
	"boardwallfloor/minigit/internal/object"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

func RunLog(args []string) {
	if len(args) != 0 {
		slog.Error("Usage: minigit log")
		fmt.Println("Usage: minigit log")
		return
	}

	headPath := filepath.Join(".minigit", "HEAD")
	headContentBytes, err := os.ReadFile(headPath)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Info("No HEAD file found, repository likely not initialized or no commits yet.")
			return
		}
		slog.Error("Failed to read HEAD", "path", headPath, "error", err)
		return
	}

	headRef := strings.TrimSpace(string(headContentBytes))
	if !strings.HasPrefix(headRef, "ref: ") {
		slog.Error("HEAD is not pointing to a reference (detached HEAD not supported)", "content", headRef)
		return
	}
	refPath := strings.TrimPrefix(headRef, "ref: ")
	fullRefPath := filepath.Join(".minigit", refPath)

	commitHashBytes, err := os.ReadFile(fullRefPath)
	if err != nil {
		if os.IsNotExist(err) {
			slog.Info("No commits found on current branch yet.", "ref", refPath)
			return
		}
		slog.Error("Failed to read current branch ref", "path", fullRefPath, "error", err)
		return
	}
	currentCommitHash := strings.TrimSpace(string(commitHashBytes))
	slog.Debug("Starting log from", "commit", currentCommitHash)

	commitCount := 0
	for currentCommitHash != "" {
		// Optional limit
		// if commitCount >= 20 {
		// 	slog.Debug("Reached log limit")
		// 	break
		// }

		objType, content, err := object.ReadObject(currentCommitHash)
		if err != nil {
			slog.Error("Failed to read commit object", "hash", currentCommitHash, "error", err)
			break
		}
		if objType != "commit" {
			slog.Error("Object is not a commit", "hash", currentCommitHash, "type", objType)
			break
		}

		_, parentHash, authorInfo, committerInfo, commitMessage, err := object.ParseCommit(content) // Assumes ParseCommit works
		if err != nil {
			slog.Error("Failed to parse commit object", "hash", currentCommitHash, "error", err)
			break
		}

		fmt.Printf("commit %s\n", currentCommitHash)
		fmt.Printf("Author:    %s\n", authorInfo)
		fmt.Printf("Committer: %s\n", committerInfo)
		fmt.Printf("\n\t%s\n\n", strings.ReplaceAll(commitMessage, "\n", "\n\t"))

		// 7. Move to Parent Commit
		if len(parentHash) > 0 {
			currentCommitHash = parentHash[0]
			slog.Debug("Moving to parent", "commit", currentCommitHash)
		} else {
			// No more parents, this was the root commit
			currentCommitHash = ""
			slog.Debug("Reached root commit")
		}
		commitCount++

	} // End history loop

	if commitCount == 0 {
		slog.Info("No commit history to display.")
	}
}
