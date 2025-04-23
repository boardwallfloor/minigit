package index

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

type IndexEntry struct {
	Mode os.FileMode
	Hash string
	Path string
}

type Index struct {
	Entries map[string]IndexEntry
}

func ReadIndex() (*Index, error) {
	file, err := os.Open(filepath.Join("./.minigit", "index"))
	if errors.Is(err, os.ErrNotExist) {
		return &Index{Entries: map[string]IndexEntry{}}, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	index := &Index{Entries: make(map[string]IndexEntry)}

	newReader := bufio.NewScanner(file)
	for newReader.Scan() {
		line := newReader.Text()
		if line == "" {
			continue
		}
		res := strings.Fields(line)
		if len(res) != 3 {
			return nil, fmt.Errorf("invalid index entry format: %s", line)
		}
		mode, err := strconv.ParseUint(res[0], 8, 32)
		if err != nil {
			return nil, fmt.Errorf("failed to parse mode: %v", err)
		}
		hash := res[1]
		path := res[2]
		index.Entries[path] = IndexEntry{Mode: os.FileMode(mode), Hash: hash, Path: path}
	}
	scanErr := newReader.Err()
	if scanErr != nil {
		return nil, fmt.Errorf("error reading index file: %v", scanErr)
	}
	return index, nil
}

func WriteIndex(idx *Index) error {
	indexPath := filepath.Join("./.minigit", "index")
	repoDir := filepath.Dir(indexPath)

	var filePath []string
	for _, v := range idx.Entries {
		filePath = append(filePath, v.Path)
	}
	slices.Sort(filePath)

	tmp, err := os.CreateTemp(repoDir, "index_*.tmp")
	if err != nil {
		return err
	}

	renameStatus := false
	defer func() {
		tmp.Close()
		if !renameStatus {
			os.Remove(tmp.Name())
		}
	}()

	writer := bufio.NewWriter(tmp)
	for _, path := range filePath {
		entry := idx.Entries[path]
		_, err := fmt.Fprintf(writer, "%0o %s %s\n", entry.Mode, entry.Hash, entry.Path)
		if err != nil {
			return err
		}
	}

	if err = writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush writer: %v", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %v", err)
	}

	err = os.Rename(tmp.Name(), "./.minigit/index")
	if err != nil {
		_ = os.Remove(tmp.Name())
		return fmt.Errorf("failed to rename temp file: %v", err)
	}

	renameStatus = true
	return nil
}
