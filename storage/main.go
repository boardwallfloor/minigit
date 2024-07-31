package main

import (
	"crypto/sha1"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type FileTree struct {
	root string
}

func (ft *FileTree) walkDir(path string, depth int) {
	dirs, err := os.ReadDir(path)
	if err != nil {
		log.Fatalf("skill issue at , :%s\n", err)
	}
	for _, file := range dirs {
		if !file.IsDir() {

			fmt.Printf("%s%s\n", strings.Repeat("-", depth), file.Name())
			filePath := filepath.Join(path, file.Name())

			hashses := generateHash(filePath)
			fmt.Printf("dir name : %s, file name : %s\n", hashses[:2], hashses[2:])
		} else {
			fmt.Printf("%s%s/\n", strings.Repeat("-", depth), file.Name())
			ft.walkDir(filepath.Join(path, file.Name()), depth+1)
		}
	}
}

func generateHash(name string) string {
	file, err := os.ReadFile(name)
	if err != nil {
		log.Fatal(err)
	}

	hasher := sha1.New()
	hasher.Write(file)
	fileHash := hasher.Sum(nil)
	fmt.Printf("Hash output : %x\n", fileHash)
	return fmt.Sprintf("%x", fileHash)
}

func main() {
	testFilename := "./test/test.txt"
	file, err := os.Create(testFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	ft := FileTree{root: "./test/"}
	ft.walkDir(ft.root, 0)
	err = os.Remove(testFilename)
	if err != nil {
		log.Fatal(err)
	}
}
