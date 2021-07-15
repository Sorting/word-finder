package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"
	"unicode/utf8"
	"github.com/fatih/color"
)

type WordLetterKey struct {
	Letter     rune
	Position   int
	WordLength int
}

func getWords() []string {
	resp, err := http.Get("https://raw.githubusercontent.com/almgru/svenska-ord.txt/master/svenska-ord.json")

	if err != nil {
		fmt.Errorf("Unable to load the list of words ", err)
	}

	payload, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}
	var data []string

	err = json.Unmarshal(payload, &data)

	if err != nil {
		panic(err)
	}

	return data
}

func buildGraph(words []string) (map[WordLetterKey]map[string]bool, map[string]map[rune]bool, map[int][]string) {
	graph := make(map[WordLetterKey]map[string]bool)
	wordCharMap := make(map[string]map[rune]bool)
	wordsLengthMap := make(map[int][]string)
	for _, word := range words {
		charMap := make(map[rune]bool)
		wordLength := utf8.RuneCountInString(word)
		if m, ok := wordsLengthMap[wordLength]; ok {
			m = append(m, word)
			wordsLengthMap[wordLength] = m
		} else {
			wordsLengthMap[wordLength] = []string{word}
		}
		pos := 0
		for _, char := range word {
			charMap[char] = true
			key := WordLetterKey{Letter: char, WordLength: wordLength, Position: pos}
			if value, ok := graph[key]; ok {
				value[word] = true
				pos++
				continue
			}
			graph[key] = map[string]bool{word: true}
			pos++
		}
		wordCharMap[word] = charMap
	}
	return graph, wordCharMap, wordsLengthMap
}

func getKeys(s string) []WordLetterKey {
	var keys []WordLetterKey
	wordLength := utf8.RuneCountInString(s)
	pos := 0
	for _, c := range s {
		if c == '_' {
			pos++
			continue
		}
		keys = append(keys, WordLetterKey{WordLength: wordLength, Letter: c, Position: pos})
		pos++
	}
	return keys
}

func charsInWord(wordChars map[string]map[rune]bool, candidates []string, chars string) []string {
	var res []string
	var aux func(string, int) *string
	aux = func(candidate string, i int) *string {
		if i == utf8.RuneCountInString(chars) {
			return &candidate
		}
		var has bool
		for _, c := range chars {
			if m, ok := wordChars[candidate]; ok {
				if _, ok := m[c]; ok {
					has = true
				} else {
					has = false
					break
				}
			}
		}
		if has {
			return aux(candidate, i+1)
		}
		return nil
	}

	for _, candidate := range candidates {
		s := aux(candidate, 0)
		if s != nil {
			res = append(res, *s)
		}
	}

	return res
}

func getBestWord(graph map[WordLetterKey]map[string]bool, wordChars map[string]map[rune]bool, wordLengthsMap map[int][]string, s string) []string {
	if words, ok := wordLengthsMap[utf8.RuneCountInString(s)]; ok {
		res := charsInWord(wordChars, words, s)
		return res
	}
	return nil
}

func getCandidates(graph map[WordLetterKey]map[string]bool, wordChars map[string]map[rune]bool, wordLengthsMap map[int][]string, s string) []string {
	sp := strings.Split(s, " ")
	s = sp[0]
	keys := getKeys(s)
	var res []string
	var aux func(string, int) *string
	aux = func(candidate string, i int) *string {
		if i == len(keys) {
			return &candidate
		}

		words := graph[keys[i]]
		if _, ok := words[candidate]; ok {
			return aux(candidate, i+1)
		}
		return nil
	}
	var candidates map[string]bool
	if keys != nil {
		candidates = graph[keys[0]]
	} else {
		if words, ok := wordLengthsMap[utf8.RuneCountInString(s)]; ok {
			if len(sp) > 1 {
				return charsInWord(wordChars, words, sp[1])
			}
			return words
		}
		return nil
	}

	for candidate, _ := range candidates {
		s := aux(candidate, 1)
		if s != nil {
			res = append(res, *s)
		}
	}

	if len(sp) > 1 {
		res = charsInWord(wordChars, res, sp[1])
	}

	sort.Strings(res)
	return res
}

func main() {
	words := getWords()

	graph, wordChars, wordLengthsMap := buildGraph(words)

	color.Green("Antal indexerade ord: %d", len(words))

	reader := bufio.NewReader(os.Stdout)

	for {
		fmt.Println(color.HiYellowString("\nVänligen ange ordmask:"))
		s, _ := reader.ReadString('\n')
		s = strings.Trim(s, "\n")
		fmt.Println()
		if len(s) == 0 {
			color.HiRed("Du måste ange ett ord 😂\n")
			continue
		}

		candidates := getCandidates(graph, wordChars, wordLengthsMap, s)

		fmt.Println(
			"Det finns",
			color.HiGreenString("%d", len(candidates)),
			"möjliga ord som matchar den angivna ordmasken:")

		for _, candidate := range candidates {
			fmt.Println("\t*", color.HiGreenString("%s", candidate))

		}
	}
}
