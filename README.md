# Minigit: A Minimal Git Implementation in Go

## Project Overview

Minigit is a command-line application built from scratch in Go that replicates core functionalities of the Git version control system. Developed as a learning exercise, this project demonstrates a deep dive into version control fundamentals, Git's internal object model, and the implementation of essential commands like `init`, `add`, `commit`, `log`, and `diff`.


## Motivation

Git is indispensable in modern software development, yet its internal mechanics can be opaque. Minigit was created to demystify Git by implementing its core concepts directly. This project provided hands-on experience with:

* **Version Control Principles:** Understanding how changes are tracked and history is maintained.
* **Content-Addressable Storage:** Implementing object storage using SHA-1 hashing[cite: 6].
* **Git Object Model:** Modeling file content (blobs), directory structures (trees), and commit history (commits)[cite: 6].
* **Staging Area (Index):** Managing the transition between the working directory and the repository history.
* **Go Programming:** Utilizing Go's standard libraries for file system interaction, data serialization (custom text format for index and tree objects)[cite: 6], compression (zlib)[cite: 6], and building a command-line interface (`flag` package)[cite: 1].
* **Algorithm Implementation:** Implementing a diffing algorithm (based on Longest Common Subsequence) to compare file versions[cite: 5].

## Key Features Implemented

Based on the current codebase, Minigit supports the following commands:

* **`minigit init`**: Initializes a new `.minigit` repository directory structure (`objects/`, `refs/heads/`, `HEAD` file)[cite: 7].
* **`minigit add <file>...`**: Stages one or more files. Calculates SHA-1 hashes of file content (blobs), stores compressed blob objects, and updates the `.minigit/index` file with file paths, modes, and hashes[cite: 6]. It detects if file content or mode has changed since the last add.
* **`minigit commit -m <message>`**: Records the staged changes (current state of the index) as a new commit.
    * Builds tree objects recursively from the index to represent directory structures[cite: 6].
    * Identifies the parent commit by reading the current branch reference.
    * Creates and stores a commit object containing the root tree hash, parent commit hash (if any), author/committer details (currently hardcoded, timestamped), and the commit message[cite: 6].
    * Updates the current branch reference (e.g., `.minigit/refs/heads/main`) to point to the new commit hash.
* **`minigit log`**: Displays the commit history of the current branch, walking backwards from the current commit through parent pointers[cite: 8].
* **`minigit diff`**: Shows differences between file states.
    * `minigit diff`: Compares the working directory files against the staging area (index).
    * `minigit diff --staged`: Compares the staging area (index) against the last commit (HEAD).

## Technical Highlights

* **Language:** Go
* **Core Data Structures:** Custom implementations for Git's Blobs, Trees, and Commits[cite: 6].
* **Hashing:** SHA-1 for content addressing and object identification[cite: 6].
* **Storage:** Mimics Git's object storage (`.minigit/objects/XX/YYYY...`) with zlib compression[cite: 6].
* **Index:** A custom text-based index file (`.minigit/index`) acts as the staging area, managed with atomic writes for safety.
* **References:** Uses `.minigit/HEAD` and `.minigit/refs/heads/` for branch management[cite: 7].
* **Diffing:** Implements line-based diffing using a Longest Common Subsequence (LCS) approach[cite: 5].

## How to Build and Run

```bash
# 1. Ensure Go (e.g., 1.21 or later) is installed.
#    [https://go.dev/doc/install](https://go.dev/doc/install)

# 2. Clone or download the repository.
#    git clone <repository-url>
#    cd minigit

# 3. Build the executable.
#    (From the root 'minigit' directory containing main.go)
go build -o minigit .

# 4. Example Usage (in a separate test directory):
mkdir /tmp/my_test_repo && cd /tmp/my_test_repo

# Initialize a new Minigit repository
../minigit init  # Use the correct relative path to your built 'minigit' executable

# Create a file, add it, and commit it
echo "Version 1" > my_file.txt
../minigit add my_file.txt
../minigit commit -m "Initial commit: Add my_file.txt"

# Modify the file and view the unstaged diff
echo "Version 2" > my_file.txt
../minigit diff

# Stage the change and view the staged diff
../minigit add my_file.txt
../minigit diff --staged

# Commit the change
../minigit commit -m "Update my_file.txt to Version 2"

# View the commit history
../minigit log

Challenges & Learnings

    Accurately implementing the recursive construction of tree objects from the index file was a key challenge.
    Ensuring precise formatting for Git object headers and tree entries was crucial for compatibility and correctness.
    Managing state reliably through the index file and HEAD/branch references, especially during updates, required careful handling (e.g., atomic writes for the index).
    Developing the diff logic involved understanding and implementing the LCS algorithm.
    This project significantly deepened my appreciation for the design decisions and efficiency of Git's internal architecture.

Potential Future Enhancements

    Implement minigit status for a comprehensive overview of the repository state (unstaged, staged, untracked files).
    Add support for configuration files (.gitconfig equivalent) for user details.
    Implement basic branching (minigit branch <name>) and checkout (minigit checkout <branch>).
    Allow adding entire directories recursively with minigit add <directory>.
    Improve error handling and user feedback.
