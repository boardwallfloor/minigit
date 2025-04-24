package object

import (
	"boardwallfloor/minigit/internal/index"
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

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
