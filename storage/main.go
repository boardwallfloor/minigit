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

// create tree
func createGitTree() {
	// format : file_permission file_type hash file_name
	fileInfo, err := os.Stat("./main.go")
	if err != nil {
		log.Fatal("skill issue")
	}
	mode := fileInfo.Mode() & 0777
	var fileType string
	if fileInfo.IsDir() {
		fileType = "tree"
	} else {
		fileType = "blob"
	}
	var entries []string
	for i := 0; i < 10; i++ {
		entries = append(entries, fmt.Sprintf("%04o %s %s %s", mode, fileType, "hash", fileInfo.Name()))
	}
	filename := "git_tree_object.txt"

	file, err := os.Create(filename)
	if err != nil {
		log.Fatal("Error creating file:", err)
	}
	defer file.Close()

	for _, entry := range entries {
		_, err := fmt.Fprintln(file, entry)
		if err != nil {
			log.Fatal("Error inserting entry:", err)
		}
	}
}

// create blob
// create commit

func main() {
	// testFilename := "./test/test.txt"
	// file, err := os.Create(testFilename)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer file.Close()
	//
	// ft := FileTree{root: "./test/"}
	// ft.walkDir(ft.root, 0)
	// err = os.Remove(testFilename)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	createGitTree()
}
