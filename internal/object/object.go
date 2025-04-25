package object

import (
	"boardwallfloor/minigit/internal/index"
	"bufio"
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const EmptyTreeHash = "4b825dc642cb6eb9a060e54bf8d69288fbee4904" // SHA1 hash for empty tree

type TreeEntry struct {
	Mode string
	Type string
	Name string
	Hash string
}

func HashBytes(content []byte) string {
	hasher := sha1.New()
	hasher.Write(content)
	fileHash := hasher.Sum(nil)
	return fmt.Sprintf("%x", fileHash)
}

func StoreBlob(content []byte) (string, error) {
	header := fmt.Sprintf("blob %d\x00", len(content))
	fullContent := append([]byte(header), content...)
	hash := HashBytes(fullContent)

	err := storeObjectInternal(hash, fullContent)
	if err != nil {
		return "", fmt.Errorf("Failed to store object %w", err)
	}
	return hash, nil
}

func StoreTree(content []byte) (string, error) {
	header := fmt.Sprintf("tree %d\x00", len(content))
	fullContent := append([]byte(header), content...)
	hash := HashBytes(fullContent)

	err := storeObjectInternal(hash, fullContent)
	if err != nil {
		return "", fmt.Errorf("Failed to store object %w", err)
	}
	return hash, nil
}

func StoreCommit(content []byte) (string, error) {
	header := fmt.Sprintf("commit %d\x00", len(content))
	fullContent := append([]byte(header), content...)
	hash := HashBytes(fullContent)

	err := storeObjectInternal(hash, fullContent)
	if err != nil {
		return "", fmt.Errorf("Failed to store object %w", err)
	}
	return hash, nil
}

func storeObjectInternal(hash string, fullContent []byte) error {
	dirPath := filepath.Join(".minigit", "objects", hash[:2])
	filePath := filepath.Join(dirPath, hash[2:])

	if _, err := os.Stat(filePath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("Failed to check file %w", err)
	}

	var buffer bytes.Buffer
	zCompressor := zlib.NewWriter(&buffer)
	_, err := zCompressor.Write([]byte(fullContent))
	if err != nil {
		return fmt.Errorf("Failed to compress data %w", err)
	}
	err = zCompressor.Close()
	if err != nil {
		return fmt.Errorf("Failed to close compressor %w", err)
	}

	err = os.MkdirAll(dirPath, 0755)
	if err != nil {
		return fmt.Errorf("Failed to create directory %w", err)
	}

	err = os.WriteFile(filePath, buffer.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("Failed to write compressed data %w", err)
	}
	return nil
}

func groupEntriesByDirectory(allEntries map[string]index.IndexEntry) map[string][]index.IndexEntry { //
	groupedEntries := make(map[string][]index.IndexEntry)
	for _, v := range allEntries {
		currDir := filepath.Dir(v.Path)
		currDir = filepath.Clean(currDir)
		groupedEntries[currDir] = append(groupedEntries[currDir], v)
	}
	return groupedEntries
}

func buildAndStoreTree(dirPath string, entriesByDir map[string][]index.IndexEntry) (treeHash string, err error) {
	slog.Debug("Building tree for", "dirPath", dirPath)

	directEntriesInThisDir := entriesByDir[dirPath]

	var currentTreeEntries []TreeEntry
	subdirsToProcess := make(map[string]bool)

	for _, entry := range directEntriesInThisDir {
		baseName := filepath.Base(entry.Path)
		currentTreeEntries = append(currentTreeEntries, TreeEntry{
			Mode: fmt.Sprintf("%0o", entry.Mode),
			Type: "blob",
			Hash: entry.Hash,
			Name: baseName,
		})

	}

	cleanedDirPath := filepath.Clean(dirPath)
	for potentialSubdirPath := range entriesByDir {
		if potentialSubdirPath == cleanedDirPath || potentialSubdirPath == "." {
			continue
		}
		parentOfPotential := filepath.Clean(filepath.Dir(potentialSubdirPath))
		if parentOfPotential == cleanedDirPath || (cleanedDirPath == "." && !strings.Contains(potentialSubdirPath, string(filepath.Separator))) {
			subDirName := filepath.Base(potentialSubdirPath)
			if cleanedDirPath == "." {
				subDirName = potentialSubdirPath
				subDirName = strings.Split(subDirName, string(filepath.Separator))[0]
			}
			subdirsToProcess[subDirName] = true
			slog.Debug("Identified actual subdir from map keys", "parent", dirPath, "subdir", subDirName)
		}
	}

	for subDirName := range subdirsToProcess {
		slog.Debug("Recursively processing subdir", "parent", dirPath, "subdir", subDirName)
		var subdirFullPath string
		if dirPath == "." {
			subdirFullPath = subDirName
		} else {
			subdirFullPath = filepath.Join(dirPath, subDirName)
		}

		subtreeHash, err := buildAndStoreTree(subdirFullPath, entriesByDir)
		if err != nil {
			return "", fmt.Errorf("failed to build subtree %s: %w", subdirFullPath, err)
		}
		slog.Debug("Received subtree hash", "subdir", subDirName, "hash", subtreeHash)

		currentTreeEntries = append(currentTreeEntries, TreeEntry{
			Mode: "40000",
			Type: "tree",
			Hash: subtreeHash,
			Name: subDirName,
		})
	}

	sort.Slice(currentTreeEntries, func(i, j int) bool {
		return currentTreeEntries[i].Name < currentTreeEntries[j].Name
	})

	var treeContent bytes.Buffer
	for _, entry := range currentTreeEntries {
		_, err := fmt.Fprintf(&treeContent, "%s %s %s\t%s\n", entry.Mode, entry.Type, entry.Hash, entry.Name)
		if err != nil {
			return "", fmt.Errorf("failed to format tree entry for %s in %s: %w", entry.Name, dirPath, err)
		}
	}

	if treeContent.Len() == 0 {
		slog.Debug("Directory is empty, storing empty tree", "dirPath", dirPath)
	}

	treeHash, err = StoreTree(treeContent.Bytes())
	if err != nil {
		return "", fmt.Errorf("failed to store tree object for %s: %w", dirPath, err)
	}
	slog.Debug("Stored tree", "dirPath", dirPath, "hash", treeHash)

	return treeHash, nil
}

func WriteTreeFromIndex(indexPath *index.Index) (string, error) {
	if len(indexPath.Entries) == 0 {
		return "4b825dc642cb6eb9a060e54bf8d69288fbee4904", nil
	}

	entriesByDir := groupEntriesByDirectory(indexPath.Entries)
	rootHash, err := buildAndStoreTree(".", entriesByDir)
	if err != nil {
		return "", fmt.Errorf("failed to write tree from index: %w", err)
	}

	return rootHash, nil
}

func ReadObject(hash string) (string, []byte, error) {
	if len(hash) != 40 {
		return "", nil, fmt.Errorf("invalid hash length: %s", hash)
	}
	filePath := filepath.Join(".minigit", "objects", hash[:2], hash[2:])
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil, fmt.Errorf("object %s not found", hash)
		} else {
			return "", nil, fmt.Errorf("failed to open object %s: %w", hash, err)
		}
	}
	defer file.Close()

	zlibReader, err := zlib.NewReader(file)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create zlib reader for object %s: %w", hash, err)
	}
	defer zlibReader.Close()

	res, err := io.ReadAll(zlibReader)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read object %s: %w", hash, err)
	}

	nullIndex := bytes.IndexByte(res, 0)
	if nullIndex == -1 {
		return "", nil, fmt.Errorf("invalid object format for %s", hash)
	}

	headerBytes := res[:nullIndex]
	content := res[nullIndex+1:] // Assign to return variable

	headerString := string(headerBytes)
	headerParts := strings.SplitN(headerString, " ", 2) // Use SplitN
	if len(headerParts) != 2 {
		return "", nil, fmt.Errorf("invalid object header format for %s: '%s'", hash, headerString)
	}

	objectType := headerParts[0] // Assign to return variable
	sizeStr := headerParts[1]

	// Validate object type using switch
	switch objectType {
	case "blob", "tree", "commit":
		// Valid type
	default:
		return "", nil, fmt.Errorf("unknown object type '%s' for hash %s", objectType, hash)
	}

	// Parse and validate size
	expectedSize, err := strconv.Atoi(sizeStr)
	if err != nil {
		return "", nil, fmt.Errorf("invalid object size '%s' for hash %s: %w", sizeStr, hash, err)
	}

	if expectedSize != len(content) {
		return "", nil, fmt.Errorf("object size mismatch for %s: expected %d, got %d", hash, expectedSize, len(content))
	}

	// Return the correct type, the actual content bytes, and nil error
	return objectType, content, nil
}

func ParseCommit(content []byte) (string, []string, string, string, string, error) {
	reader := bytes.NewReader(content)
	scanner := bufio.NewScanner(reader)

	foundBlankLine := false
	messageBuilder := strings.Builder{}
	var treeHash, authorInfo, committerInfo string
	var parentHash []string
	for scanner.Scan() {
		line := scanner.Text()
		if !foundBlankLine {
			if line == "" {
				foundBlankLine = true
				continue
			}

			parts := strings.SplitN(line, " ", 2)
			if len(parts) != 2 {
				// Allow unknown headers? Or error? Let's ignore for now.
				continue
			}
			key, value := parts[0], strings.TrimSpace(parts[1])

			switch key {
			case "tree":
				treeHash = value
			case "parent":
				parentHash = append(parentHash, value)
			case "author":
				authorInfo = value
			case "committer":
				committerInfo = value
			}
		} else {
			// We are now reading the commit message part
			messageBuilder.WriteString(line)
			messageBuilder.WriteString("\n") // Re-add newline removed by scanner
		}
	}
	if err := scanner.Err(); err != nil {
		return "", nil, "", "", "", fmt.Errorf("failed to parse commit content: %w", err)
	}

	commitMessage := strings.TrimSpace(messageBuilder.String())
	if treeHash == "" || authorInfo == "" || committerInfo == "" {
		return "", nil, "", "", "", fmt.Errorf("missing required fields in commit content")
	}
	return treeHash, parentHash, authorInfo, committerInfo, commitMessage, nil
}

func ReadTreeEntries(treeHash string) (map[string]index.IndexEntry, error) {
	if treeHash == "" || treeHash == EmptyTreeHash { // Handle empty tree case
		return make(map[string]index.IndexEntry), nil
	}
	results := make(map[string]index.IndexEntry)
	err := readTreeRecursive(treeHash, "", &results) // Start recursion with empty base path
	if err != nil {
		return nil, err
	}
	return results, nil
}

// readTreeRecursive is the helper function for ReadTreeEntries.
func readTreeRecursive(currentTreeHash string, currentBasePath string, results *map[string]index.IndexEntry) error {
	slog.Debug("Reading tree entries for", "hash", currentTreeHash, "basePath", currentBasePath)

	objType, content, err := ReadObject(currentTreeHash) // Assumes ReadObject exists and works
	if err != nil {
		return fmt.Errorf("failed reading tree object %s: %w", currentTreeHash, err)
	}
	if objType != "tree" {
		return fmt.Errorf("object %s is not a tree, type %s", currentTreeHash, objType)
	}

	reader := bytes.NewReader(content)
	scanner := bufio.NewScanner(reader)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Parse format: <mode> <space> <type> <space> <hash> <tab> <name>
		parts1 := strings.SplitN(line, " ", 3) // Split into mode, type, rest
		if len(parts1) != 3 {
			return fmt.Errorf("invalid tree entry format line %d in %s: %s", lineNumber, currentTreeHash, line)
		}
		modeStr, entryType, rest := parts1[0], parts1[1], parts1[2]

		parts2 := strings.SplitN(rest, "\t", 2) // Split rest into hash, name
		if len(parts2) != 2 {
			return fmt.Errorf("invalid tree entry format line %d in %s (missing tab?): %s", lineNumber, currentTreeHash, line)
		}
		entryHash, entryName := parts2[0], parts2[1]

		// Validate basic components
		if entryName == "" || entryHash == "" || entryType == "" || modeStr == "" {
			return fmt.Errorf("invalid tree entry format line %d in %s (empty component): %s", lineNumber, currentTreeHash, line)
		}

		fullPath := filepath.Join(currentBasePath, entryName)
		// Use Clean to normalize path separators, especially important if currentBasePath was ""
		fullPath = filepath.Clean(fullPath)
		// Remove leading "." if present after join/clean with empty base path
		fullPath = strings.TrimPrefix(fullPath, "."+string(filepath.Separator))

		switch entryType {
		case "blob":
			modeUint, err := strconv.ParseUint(modeStr, 8, 32)
			if err != nil {
				return fmt.Errorf("invalid mode '%s' on line %d in %s: %w", modeStr, lineNumber, currentTreeHash, err)
			}
			entry := index.IndexEntry{
				Mode: os.FileMode(modeUint),
				Hash: entryHash,
				Path: fullPath,
			}
			(*results)[fullPath] = entry // Add blob entry to the map
			slog.Debug("Found blob entry", "path", fullPath, "hash", entryHash)

		case "tree":
			slog.Debug("Recursing into tree", "path", fullPath, "hash", entryHash)
			err := readTreeRecursive(entryHash, fullPath, results) // Recursive call
			if err != nil {
				return fmt.Errorf("failed processing subtree %s: %w", fullPath, err)
			}
		default:
			return fmt.Errorf("unknown entry type '%s' on line %d in %s", entryType, lineNumber, currentTreeHash)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error scanning tree object %s: %w", currentTreeHash, err)
	}

	return nil // Success for this level
}
