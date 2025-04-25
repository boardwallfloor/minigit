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

func RunDiff(args []string, staged bool) {
	// TODO: Handle args later to diff specific files/paths

	if staged {
		runDiffStaged()
	} else {
		runDiffWorkdirVsIndex()
	}
}

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
	// TODO: Need to also detect files modified in WorkDir but NOT in index? (More like status)
	// TODO: Need to detect files deleted from WorkDir that ARE in index?
	for path, indexEntry := range idx.Entries {
		workDirContent, err := os.ReadFile(path)
		fileDeletedInWorkDir := false
		if err != nil {
			if os.IsNotExist(err) {
				fileDeletedInWorkDir = true
				workDirContent = []byte{}
				slog.Debug("File deleted from workdir", "path", path)
			} else {
				slog.Error("Failed to read working directory file", "path", path, "error", err)
				continue
			}
		}

		objType, indexContent, err := object.ReadObject(indexEntry.Hash)
		if err != nil {
			slog.Error("Failed to read index object", "path", path, "hash", indexEntry.Hash, "error", err)
			continue
		}
		if objType != "blob" {
			slog.Error("Index entry points to non-blob object", "path", path, "hash", indexEntry.Hash, "type", objType)
			continue
		}

		if fileDeletedInWorkDir || !bytes.Equal(workDirContent, indexContent) {
			foundDiff = true
			diffString, diffErr := diff.GenerateLinesDiff(indexContent, workDirContent)
			if diffErr != nil {
				slog.Error("Failed to generate diff", "path", path, "error", diffErr)
				continue
			}

			// TODO: Improve header formatting (e.g., include hashes)
			fmt.Printf("diff --git a/%s b/%s\n", path, path)
			fmt.Printf("--- a/%s\n", path)
			fmt.Printf("+++ b/%s\n", path)
			fmt.Print(diffString)
			fmt.Println()
		} else {
			slog.Debug("No textual changes detected", "path", path)
		}
	}

	if !foundDiff {
		slog.Info("No differences detected between index and working directory.")
	}
}

func runDiffStaged() {
	slog.Debug("Running diff Index vs HEAD (--staged)")

	idx, err := index.ReadIndex()
	if err != nil {
		slog.Error("Failed to read index", "error", err)
		return
	}
	indexEntries := idx.Entries

	headCommitHash, err := getCurrentHeadCommitHash()
	if err != nil {
		slog.Error("Failed to get HEAD commit hash", "error", err)
		return
	}

	objType, commitContent, err := object.ReadObject(headCommitHash)
	if err != nil || objType != "commit" {
		slog.Error("Failed to read HEAD commit object", "hash", headCommitHash, "error", err)
		return
	}
	treeHash, _, _, _, _, err := object.ParseCommit(commitContent)
	if err != nil {
		slog.Error("Failed to parse HEAD commit object", "hash", headCommitHash, "error", err)
		return
	}
	rootTreeHash := treeHash

	headEntries, err := object.ReadTreeEntries(rootTreeHash)
	if err != nil {
		slog.Error("Failed to read HEAD tree entries", "treeHash", rootTreeHash, "error", err)
		return
	}

	foundDiff := false

	allPaths := make(map[string]bool)
	for path := range indexEntries {
		allPaths[path] = true
	}
	for path := range headEntries {
		allPaths[path] = true
	}

	sortedPaths := getSortedPaths(allPaths)

	for _, path := range sortedPaths {
		indexEntry, inIndex := indexEntries[path]
		headEntry, inHead := headEntries[path]

		if inIndex && !inHead {
			foundDiff = true
			_, indexContent, err := object.ReadObject(indexEntry.Hash)
			if err != nil {
				continue
			}
			diffString, diffErr := diff.GenerateLinesDiff([]byte{}, indexContent)
			if diffErr != nil {
				continue
			}
			printDiffHeader(path, path, indexEntry.Mode, true)
			fmt.Print(diffString)
			fmt.Println()
		} else if !inIndex && inHead {
			foundDiff = true
			_, headContent, err := object.ReadObject(headEntry.Hash)
			if err != nil {
				continue
			}
			diffString, diffErr := diff.GenerateLinesDiff(headContent, []byte{})
			if diffErr != nil {
				continue
			}
			printDiffHeader(path, path, headEntry.Mode, false)
			fmt.Print(diffString)
			fmt.Println()
		} else if inIndex && inHead {
			if indexEntry.Hash != headEntry.Hash || indexEntry.Mode != headEntry.Mode {
				foundDiff = true
				_, headContent, errH := object.ReadObject(headEntry.Hash)
				_, indexContent, errI := object.ReadObject(indexEntry.Hash)
				if errH != nil || errI != nil {
					continue
				}

				diffString, diffErr := diff.GenerateLinesDiff(headContent, indexContent)
				if diffErr != nil {
					continue
				}
				printDiffHeader(path, path, indexEntry.Mode, false)
				// TODO: Add index line like `index <hashA>..<hashB> <mode>`
				fmt.Print(diffString)
				fmt.Println()
			}
		}
	}

	if !foundDiff {
		slog.Info("No differences detected between index and HEAD.")
	}
}

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
		return "", fmt.Errorf("HEAD is not pointing to a symbolic ref: %s", headRef)
	}
	refPath := strings.TrimPrefix(headRef, "ref: ")
	fullRefPath := filepath.Join(".minigit", refPath)

	commitHashBytes, err := os.ReadFile(fullRefPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("no commits found on current branch (%s)", refPath)
		}
		return "", fmt.Errorf("failed to read current branch ref %s: %w", fullRefPath, err)
	}
	return strings.TrimSpace(string(commitHashBytes)), nil
}

func printDiffHeader(pathA, pathB string, mode os.FileMode, isNew bool) {
	fmt.Printf("diff --git a/%s b/%s\n", pathA, pathB)
	if isNew {
		fmt.Printf("new file mode %0o\n", mode)
	}
	// TODO: Add index line: fmt.Printf("index %s..%s %0o\n", hashA, hashB, mode)
	fmt.Printf("--- a/%s\n", pathA)
	fmt.Printf("+++ b/%s\n", pathB)
}

func getSortedPaths(pathMap map[string]bool) []string {
	paths := make([]string, 0, len(pathMap))
	for path := range pathMap {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}
