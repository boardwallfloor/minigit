package cmd

import (
	"boardwallfloor/minigit/internal/diff"
	"boardwallfloor/minigit/internal/index"
	"boardwallfloor/minigit/internal/object"
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// RunDiff handles the 'minigit diff' command.
// It shows differences between states (WorkDir vs Index, or Index vs HEAD).
// args currently unused but could filter by path later.
func RunDiff(args []string, staged bool) {
	// TODO: Handle args later to diff specific files/paths

	if staged {
		runDiffStaged()
	} else {
		runDiffWorkdirVsIndex()
	}
}

// runDiffWorkdirVsIndex compares the working directory against the index.
func runDiffWorkdirVsIndex() {
	slog.Debug("Running diff WorkDir vs Index")
	idx, err := index.ReadIndex()
	if err != nil {
		slog.Error("Failed to read index", "error", err)
		return
	}

	if len(idx.Entries) == 0 {
		slog.Info("Index is empty, nothing to diff against working directory.")
		return
	}

	foundDiff := false
	// Iterate through files currently tracked in the index
	// TODO: Need to also detect files modified in WorkDir but NOT in index? (More like status)
	// TODO: Need to detect files deleted from WorkDir that ARE in index?
	for path, indexEntry := range idx.Entries {
		// Read working directory content
		workDirContent, err := os.ReadFile(path)
		fileDeletedInWorkDir := false
		if err != nil {
			if os.IsNotExist(err) {
				// File exists in index but not in working dir -> Deleted from work dir
				fileDeletedInWorkDir = true
				workDirContent = []byte{} // Treat deleted file as empty content for diff
				slog.Debug("File deleted from workdir", "path", path)
			} else {
				slog.Error("Failed to read working directory file", "path", path, "error", err)
				continue // Skip this file if unreadable
			}
		}

		// Read index blob content
		objType, indexContent, err := object.ReadObject(indexEntry.Hash)
		if err != nil {
			slog.Error("Failed to read index object", "path", path, "hash", indexEntry.Hash, "error", err)
			continue // Skip this file
		}
		if objType != "blob" {
			slog.Error("Index entry points to non-blob object", "path", path, "hash", indexEntry.Hash, "type", objType)
			continue // Skip this file
		}

		// Compare contents only if file wasn't just deleted OR if content differs
		// Using bytes.Equal is a quick check before generating the full diff string
		if fileDeletedInWorkDir || !bytes.Equal(workDirContent, indexContent) {
			foundDiff = true
			// Generate the diff using your refactored function
			// Note the order: index content is 'a' (old), workdir content is 'b' (new)
			diffString, diffErr := diff.GenerateLinesDiff(indexContent, workDirContent)
			if diffErr != nil {
				slog.Error("Failed to generate diff", "path", path, "error", diffErr)
				continue
			}

			// Print standard diff header and the diff content
			// TODO: Improve header formatting (e.g., include hashes)
			fmt.Printf("diff --git a/%s b/%s\n", path, path)
			fmt.Printf("--- a/%s\n", path) // Represents index version
			fmt.Printf("+++ b/%s\n", path) // Represents working directory version
			fmt.Print(diffString)          // Assumes diffString includes necessary newlines
			fmt.Println()                  // Add a blank line between files
		} else {
			slog.Debug("No textual changes detected", "path", path)
		}
	} // End loop through index entries

	if !foundDiff {
		slog.Info("No differences detected between index and working directory.")
	}
	// Note: This basic diff doesn't show untracked files or files deleted from index but present in workdir.
}

// runDiffStaged compares the index against the HEAD commit.
func runDiffStaged() {
	slog.Debug("Running diff Index vs HEAD (--staged)")

	// 1. Read Index
	idx, err := index.ReadIndex()
	if err != nil {
		slog.Error("Failed to read index", "error", err)
		return
	}
	indexEntries := idx.Entries // Get the map

	// 2. Get HEAD Commit & Tree
	headCommitHash, err := getCurrentHeadCommitHash()
	if err != nil {
		slog.Error("Failed to get HEAD commit hash", "error", err)
		return // Includes "no commits yet" case
	}

	objType, commitContent, err := object.ReadObject(headCommitHash)
	if err != nil || objType != "commit" {
		slog.Error("Failed to read HEAD commit object", "hash", headCommitHash, "error", err)
		return
	}
	treeHash, _, _, _, _, err := object.ParseCommit(commitContent) // Assumes ParseCommit works
	if err != nil {
		slog.Error("Failed to parse HEAD commit object", "hash", headCommitHash, "error", err)
		return
	}
	rootTreeHash := treeHash // This is the tree hash we need to diff against

	// 3. Read HEAD Tree Entries using the helper
	headEntries, err := object.ReadTreeEntries(rootTreeHash) // Assumes ReadTreeEntries works
	if err != nil {
		slog.Error("Failed to read HEAD tree entries", "treeHash", rootTreeHash, "error", err)
		return
	}

	// 4. Compare Index vs HEAD
	foundDiff := false

	// Combine paths from both maps to ensure we check all files
	allPaths := make(map[string]bool)
	for path := range indexEntries {
		allPaths[path] = true
	}
	for path := range headEntries {
		allPaths[path] = true
	}

	// Iterate through combined paths (could sort paths here for consistent output order)
	sortedPaths := getSortedPaths(allPaths)

	for _, path := range sortedPaths {
		indexEntry, inIndex := indexEntries[path]
		headEntry, inHead := headEntries[path]

		if inIndex && !inHead {
			// File added to index (new file)
			foundDiff = true
			_, indexContent, err := object.ReadObject(indexEntry.Hash)
			if err != nil { /* handle error */
				continue
			}
			diffString, diffErr := diff.GenerateLinesDiff([]byte{}, indexContent) // Diff against empty
			if diffErr != nil {                                                   /* handle error */
				continue
			}
			printDiffHeader(path, path, indexEntry.Mode, true) // Indicate new file mode
			fmt.Print(diffString)
			fmt.Println()
		} else if !inIndex && inHead {
			// File deleted from index
			foundDiff = true
			_, headContent, err := object.ReadObject(headEntry.Hash)
			if err != nil { /* handle error */
				continue
			}
			diffString, diffErr := diff.GenerateLinesDiff(headContent, []byte{}) // Diff against empty
			if diffErr != nil {                                                  /* handle error */
				continue
			}
			printDiffHeader(path, path, headEntry.Mode, false) // Indicate deleted file mode
			fmt.Print(diffString)
			fmt.Println()
		} else if inIndex && inHead {
			// File exists in both, check for modifications
			if indexEntry.Hash != headEntry.Hash || indexEntry.Mode != headEntry.Mode {
				foundDiff = true
				_, headContent, errH := object.ReadObject(headEntry.Hash)
				_, indexContent, errI := object.ReadObject(indexEntry.Hash)
				if errH != nil || errI != nil { /* handle error */
					continue
				}

				diffString, diffErr := diff.GenerateLinesDiff(headContent, indexContent)
				if diffErr != nil { /* handle error */
					continue
				}
				printDiffHeader(path, path, indexEntry.Mode, false) // Use current mode
				// TODO: Add index line like `index <hashA>..<hashB> <mode>`
				fmt.Print(diffString)
				fmt.Println()
			}
		}
	} // End loop through paths

	if !foundDiff {
		slog.Info("No differences detected between index and HEAD.")
	}
}

// Helper function to get current HEAD commit hash (similar to log/commit)
func getCurrentHeadCommitHash() (string, error) {
	headPath := filepath.Join(".minigit", "HEAD")
	headContentBytes, err := os.ReadFile(headPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("not a minigit repository (or no HEAD file)")
		}
		return "", fmt.Errorf("failed to read HEAD: %w", err)
	}

	headRef := strings.TrimSpace(string(headContentBytes))
	if !strings.HasPrefix(headRef, "ref: ") {
		// Handle detached HEAD later if needed, for now, assume symbolic ref
		return "", fmt.Errorf("HEAD is not pointing to a symbolic ref: %s", headRef)
	}
	refPath := strings.TrimPrefix(headRef, "ref: ")
	fullRefPath := filepath.Join(".minigit", refPath)

	commitHashBytes, err := os.ReadFile(fullRefPath)
	if err != nil {
		if os.IsNotExist(err) {
			// No commits on this branch yet
			return "", fmt.Errorf("no commits found on current branch (%s)", refPath)
		}
		return "", fmt.Errorf("failed to read current branch ref %s: %w", fullRefPath, err)
	}
	return strings.TrimSpace(string(commitHashBytes)), nil
}

// Helper to print standard diff header
func printDiffHeader(pathA, pathB string, mode os.FileMode, isNew bool) {
	fmt.Printf("diff --git a/%s b/%s\n", pathA, pathB)
	// Could add mode change detection here
	if isNew {
		fmt.Printf("new file mode %0o\n", mode)
	}
	// TODO: Add index line: fmt.Printf("index %s..%s %0o\n", hashA, hashB, mode)
	fmt.Printf("--- a/%s\n", pathA)
	fmt.Printf("+++ b/%s\n", pathB)
}

// Helper to get sorted keys from a map[string]bool
func getSortedPaths(pathMap map[string]bool) []string {
	paths := make([]string, 0, len(pathMap))
	for path := range pathMap {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}
