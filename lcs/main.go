package main

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

const testPoem1 = `Willow's hair, a golden sheen,
Dances softly in autumn's scene.
A lofty perch, where dreams take flight,
Cradled in leaves, bathed in soft light.

Tiny soul, with downy grace,
Gazes out at nature's vast space.
Pecks at the world with eager beak,
As mother bird soars beyond the peak.
Crimson leaves descend like fire,
Igniting hope, a pure desire.

Autumn's embrace, a tender hold,
Warming nest, a story untold.
Small life clings, with courage deep,
Dreaming of heights where spirits leap.`

const testPoem2 = `Willow's weep, a golden sheen,
Dances softly in autumn's scene.
A lofty perch, where dreams take flight,
Cradled in leaves, bathed in soft light.

Tiny heart, with downy grace,
Gazes out at nature's vast space.
Hungry cries, a plaintive sound,
As mother bird flits without bound.
Crimson leaves descend like fire,
Igniting hope, a pure desire.

Autumn's embrace, a tender hold,
Warming nest, a story untold.
Small life clings, with might and main,
Dreaming of skies without chain.`

func splitText(s string) []string {
	lines := strings.Split(s, "\n")
	return lines
}

func splitWord(s string) []string {
	words := strings.Split(s, " ")
	words = slices.Insert(words, 0, " ")
	return words
}

func compareString(base, comp []string) (string, string, bool) {
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
	lcsCount := lcs
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

	base = base[1:]
	comp = comp[1:]
	bs := 0
	cmp := 0
	i = 0
	var origin, compare string
	for i < len(lcsString) {
		if base[bs] == lcsString[i] && comp[cmp] == lcsString[i] {
			origin += fmt.Sprint(lcsString[i], " ")
			compare += fmt.Sprint(lcsString[i], " ")
			bs += 1
			cmp += 1
			i += 1
		} else if base[bs] != lcsString[i] {
			origin += fmt.Sprint(green, base[bs], reset, " ")
			bs += 1
		} else if comp[cmp] != lcsString[i] {
			compare += fmt.Sprint(red, comp[cmp], reset, " ")
			cmp += 1
		}
	}
	if bs < len(base) {
		origin += green + " " + strings.Join(base[bs:], " ") + reset
	}
	if cmp < len(comp) {
		compare += red + strings.Join(comp[cmp:], " ") + reset
	}

	var isDiff bool
	if lcsCount == len(base) && lcsCount == len(comp) {
		isDiff = false
	} else {
		isDiff = true
	}
	return origin, compare, isDiff
}

func main() {
	baseText := splitText(testPoem1)
	compText := splitText(testPoem2)
	for i := 0; i < len(baseText); i++ {
		baseSentence := splitWord(baseText[i])
		compSentence := splitWord(compText[i])
		base, _, isDiff := compareString(baseSentence, compSentence)
		if isDiff {
			fmt.Printf("%s%s%s\n", red, strings.Join(baseSentence[1:], " "), reset)
			fmt.Printf("%s%s%s\n", green, strings.Join(compSentence[1:], " "), reset)
		} else {
			fmt.Println(base)
		}
	}
}
