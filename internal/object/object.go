package object

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"
)

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
	dirPath := filepath.Join(".minigit", "objects", hash[:2])
	filePath := filepath.Join(dirPath, hash[2:])

	var buffer bytes.Buffer
	zCompressor := zlib.NewWriter(&buffer)
	_, err := zCompressor.Write([]byte(fullContent))
	if err != nil {
		return "", fmt.Errorf("Failed to compress data %w", err)
	}
	err = zCompressor.Close()
	if err != nil {
		return "", fmt.Errorf("Failed to close compressor %w", err)
	}

	err = os.MkdirAll(dirPath, 0755)
	if err != nil {
		return "", fmt.Errorf("Failed to create directory %w", err)
	}

	err = os.WriteFile(filePath, buffer.Bytes(), 0644)
	if err != nil {
		return "", fmt.Errorf("Failed to write compressed data %w", err)
	}
	return hash, nil
}
