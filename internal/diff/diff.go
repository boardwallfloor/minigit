package diff

import (
	"fmt"
	"slices"
	"strings"
)

const (
	reset = "\033[0m"  // Reset text color
	red   = "\033[31m" // Set text color to red
	green = "\033[32m" // Set text color to green
)

func splitText(s string) []string {
	lines := strings.Split(s, "\n")
	return lines
}

func splitWord(s string) []string {
	words := strings.Split(s, " ")
	words = slices.Insert(words, 0, " ")
	return words
}

func GenerateLinesDiff(inputBase, inputComp []byte) (string, error) {
	linesBase := string(inputBase)
	linesComp := string(inputComp)
	base := splitText(linesBase)
	comp := splitText(linesComp)

	table := make([][]int, len(base))
	for i := range table {
		table[i] = make([]int, len(comp))
	}

	lcs := 0
	for i := 1; i < len(table); i++ {
		for j := 1; j < len(table[i]); j++ {
			if base[i] == comp[j] {
				table[i][j] = table[i-1][j-1] + 1
				lcs = table[i][j]
			} else {
				table[i][j] = max(table[i][j-1], table[i-1][j])
			}
		}
	}

	i := len(table) - 1
	j := len(table[0]) - 1
	lcsString := make([]string, lcs)
	for lcs > 0 {
		if base[i] == comp[j] {
			lcsString[lcs-1] = base[i]
			lcs -= 1
			i -= 1
			j -= 1
		} else {
			if table[i][j-1] < table[i-1][j] {
				i -= 1
			} else {
				j -= 1
			}
		}
	}

	var diffOutput strings.Builder // Use a builder for efficient string construction
	bs := 0                        // Pointer for base slice
	cmp := 0                       // Pointer for comp slice
	i = 0                          // Pointer for lcsString slice

	for i < len(lcsString) {
		// While current base line is not part of the LCS -> Deleted line
		for bs < len(base) && base[bs] != lcsString[i] {
			_, err := fmt.Fprintf(&diffOutput, "- %s\n", base[bs]) // Capture error
			if err != nil {
				return "", fmt.Errorf("failed to write deleted line: %w", err) // Return "" and error
			}
			bs++
		}

		// While current comp line is not part of the LCS -> Added line
		for cmp < len(comp) && comp[cmp] != lcsString[i] {
			_, err := fmt.Fprintf(&diffOutput, "+ %s\n", comp[cmp]) // Capture error
			if err != nil {
				return "", fmt.Errorf("failed to write added line: %w", err) // Return "" and error
			}
			cmp++
		}

		// If we've reached here, base[bs] == comp[cmp] == lcsString[i] -> Common line
		if bs < len(base) && cmp < len(comp) && base[bs] == lcsString[i] {
			_, err := fmt.Fprintf(&diffOutput, "  %s\n", lcsString[i]) // Capture error
			if err != nil {
				return "", fmt.Errorf("failed to write common line: %w", err) // Return "" and error
			}
			bs++
			cmp++
			i++
		} else {
			// Defensive break - return an error instead of just logging
			return "", fmt.Errorf("diff logic error: pointers misaligned with LCS")
		}
	}

	// Append any remaining lines from base as deletions
	for bs < len(base) {
		_, err := fmt.Fprintf(&diffOutput, "- %s\n", base[bs]) // Capture error
		if err != nil {
			return "", fmt.Errorf("failed to write remaining deleted line: %w", err) // Return "" and error
		}
		bs++
	}
	for cmp < len(comp) {
		_, err := fmt.Fprintf(&diffOutput, "+ %s\n", comp[cmp]) // Capture error
		if err != nil {
			return "", fmt.Errorf("failed to write remaining added line: %w", err) // Return "" and error
		}
		cmp++
	}

	return diffOutput.String(), nil // Return the single diff string
}
