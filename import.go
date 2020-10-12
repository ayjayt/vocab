// This will read files, find all of the words, and translate them
package main

import (
	//"io"
	"io/ioutil"
	"os"
	//"encoding/json"
	"flag"
	"fmt"
	"regexp"
)

var mapFileString, inputFileString *string

var wordRegexp *regexp.Regexp

func check (e error) {
	if e != nil {
		panic(e)
	}
}

func init() {
	mapFileString = flag.String("map", "", "Specify the file you use to keep track of words you know")
	inputFileString = flag.String("input", "", "Specify the file you want to input")
	flag.Parse()
	wordRegexp = regexp.MustCompile(`\s*['\-\pL]['\-\pL]['\-\pL]+\s*`)
}

type vocab struct {
	base string
	translation string
	gender string
}

var englishMap = make(map[string][]vocab, 0)
var spanishMap = make(map[string][]vocab, 0)

// How will this actually look @ end of times.
// Deck, Spanish Form
// Spanish Form, English Translation, Note, Gender, Base(string), Learning Meta
// Base, English Translation, Note, Gender, Form, Form, Form, Form, Form, Form, Form, Form, Form, Form, Form
// We can construct decks from the base, from the applied form, and we can randomize forms


func main() {
	mapFile, err := os.Open(*mapFileString)
	check(err)
	_ = mapFile
	/*
	dec := json.NewDecoder(mapFile)
	for {
		if err := dec.Decode(&vocabMap); err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
	}
	*/
	fmt.Printf("%+v\n", englishMap)
	fmt.Printf("%+v\n", spanishMap)
	fmt.Printf("%+v\n", *mapFileString)
	inputText, err := ioutil.ReadFile(*inputFileString)
	check(err)
	allWords := wordRegexp.FindAllString(string(inputText), -1)
	if allWords == nil {
		panic("No words in input")
	}
	for _, val := range allWords {
	}

	// CommandLine dictionary to a) check against map, install answer of necessary
	// Bias for spanish
}

