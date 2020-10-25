// This will read files, find all of the words, and translate them
package main

import (
	"io/ioutil"
	//"os"
	"encoding/json"
	"flag"
	"fmt"
	"regexp"
	"context"
	"strings"
	"net/http"

	"cloud.google.com/go/translate"
	"golang.org/x/text/language"
	"github.com/sthorne/go-hunspell"



)


var inputFileString = flag.String("input", "", "Specify the file you want to input")
var homeLanguageString = flag.String("home", "en", "Specify your language")
var awayLanguageString = flag.String("away", "es", "Specify desired language")
var inputLanguageString = flag.String("input", "es,en", "Specify the languages int he document")
var wordRegexp = regexp.MustCompile(`\s*['\-\pL]['\-\pL]['\-\pL]+\s*`)

func init() {
	flag.Parse()
}

// This is basically to convert your english word to spanish
// It probably needs to be an interface
func GoogleWord(client *translate.Client, ctx context.Context, word, inLanguageString, outLanguageString string) string { // convert input language to the destination language
		// Google Translate
		outLanguage, err := language.Parse(outLanguageString)
		check(err)
		inLanguage, err := language.Parse(inLanguageString)
		check(err)

		TranslateResp, err := client.Translate(ctx, []string{word}, outLanguage, &translate.Options{
			Source: inLanguage,
		})
		check(err)
		if len(TranslateResp) == 0 {
			fmt.Printf("Returned empty on translation")
		} else {
			fmt.Printf("resp:\n%+v\n", TranslateResp)
		}
		return "" // TODO: fix return, parse this
}

type OxfordClient struct {
	http.Client
}

// OxfordDefine gets a response about definitions from an oxford dictionary
func (client OxfordClient) OxfordDefine(word string) []string {
	fmt.Printf("Requesting:https://od-api.oxforddictionaries.com/api/v2/lemma/es/" + word + "\n")
	req, err := http.NewRequest("GET", "https://od-api.oxforddictionaries.com/api/v2/entries/es/" + word, nil)
	req.Header.Add("app_id", ``)
	req.Header.Add("app_key", ``) // TODO: fix get keys out
	resp, err := client.Do(req)
	check(err)
	jsonData := make(map[string]interface{})
	if resp.StatusCode == http.StatusOK{
		json.NewDecoder(resp.Body).Decode(&jsonData)
		b, err := json.MarshalIndent(jsonData, "", "  ")
		check(err)
		return []string{string(b)}
	}
	return nil
}

// ReadFile right now reads a global file
func ReadFile() map[string]bool {
	inputText, err := ioutil.ReadFile(*inputFileString)
	check(err)
	allWordsFromRaw := wordRegexp.FindAllString(string(inputText), -1)
	if allWordsFromRaw == nil {
		return nil
	}
	wordMapRaw := make(map[string]bool, len(allWordsFromRaw))
	for _, val := range allWordsFromRaw {
		wordMapRaw[strings.ToLower(strings.TrimSpace(val))] = true
	}
	return wordMapRaw
}

func main() {

	// These are essentially the results
	var wordMap = make(map[string][]Vocab, 0)

	// Setting up the dictionaries- not sure if they're thread safe
	dictionaries := make(map[string]*hunspell.Hunhandle, 2)
	// All dictionaries will be cached
	dictionaries["es"] = hunspell.Hunspell("resources/hunspell_dictionaries/es_ES.aff", "hunspell_dictionaries/es_ES.dic")
	dictionaries["en"] = hunspell.Hunspell("resources/hunspell_dictionaries/en_US.aff", "hunspell_dictionaries/en_US.dic")

	inputLanguages := strings.Split(*inputLanguageString,",")
	for _, v := range inputLanguages {
		language.MustParse(v)
	}

	// This is an http client for oxford's api. Probably going to take a context too TODO. 
	oxfordClient := &OxfordClient{}

	// This is the setup for google translate API
	ctx := context.Background()
	googleClient, err := translate.NewClient(ctx)
	if err != nil {
		panic("translate.NewClient: " + err.Error())
	}
	defer googleClient.Close()

	// Start work
	wordMapRaw := ReadFile()

	// Iterate over words
	for index, _ := range wordMapRaw {
		fmt.Printf("\n\n%v\n", index)

		// Convert all input languages to the away language
		for _, lang := range inputLanguages {
			if (dictionaries[lang].Spell(index)) {
				if (lang != *awayLanguageString) {
					GoogleWord(googleClient, ctx, index, lang, *awayLanguageString)
					// Lets store how much we're googling
				}
				if _, ok := wordMap[index]; !ok { // TODO maybe we should lookup the word
					// Now we need to lemmatize it TODO
					wordMap[index] = make([]Vocab,0) // TODO do we even need to create a wordMap like this?
					// We're checking dictionary TODO
					// Adding it if it's not there TODO
					// And adding it to the deck if it is now or was TODO
				}
			}
		}

	}
	// We're going to get the base for each word now
		// Oxford Translate and Dump
		definition := oxfordClient.OxfordDefine("") // TODO: this should be a base
		_ = definition
}

func check (e error) {
	if e != nil {
		panic(e)
	}
}

