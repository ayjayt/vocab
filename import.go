// This will read files, find all of the words, and translate them
package main

import (
	"io/ioutil"
	"os"
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
var inputLanguageString = flag.String("input_languages", "es,en", "Specify the languages int he document")
var wordRegexp = regexp.MustCompile(`\s*['\-\pL]['\-\pL]['\-\pL]+\s*`)

func init() {
	flag.Parse()
}

// GoogleWord will translate a word TODO: can batch
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
			fmt.Printf("Returned empty on translation\n")
		} else {
			fmt.Printf("resp:\n%+v\n", TranslateResp)
		}
		return strings.ToLower(TranslateResp[0].Text)
}

type OxfordClient struct {
	http.Client
}

// OxfordDefine gets a response about definitions from an oxford dictionary
func (client OxfordClient) OxfordDefine(word string) []string {
	fmt.Printf("Requesting:https://od-api.oxforddictionaries.com/api/v2/lemma/es/" + word + "\n")
	req, err := http.NewRequest("GET", "https://od-api.oxforddictionaries.com/api/v2/entries/es/" + word, nil)
	OXFORD_ID := os.Getenv("OXFORD_ID")
	OXFORD_PASS := os.Getenv("OXFORD_PASS")
	if (len(OXFORD_ID) == 0 || len(OXFORD_PASS) == 0) {
		panic("I need key headers OXFORD_ID and OXFORD_PASS")
	}
	req.Header.Add("app_id", OXFORD_ID)
	req.Header.Add("app_key", OXFORD_PASS) // TODO: fix get keys out
	resp, err := client.Do(req)
	check(err)
	jsonData := make(map[string]interface{})
	if resp.StatusCode == http.StatusOK{
		json.NewDecoder(resp.Body).Decode(&jsonData)
		b, err := json.MarshalIndent(jsonData, "", "  ")
		check(err)
		return []string{string(b)} // TODO: still returning whole response
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
	dictionaries["es"] = hunspell.Hunspell("resources/hunspell_dictionaries/es_ES.aff", "resources/hunspell_dictionaries/es_ES.dic")
	dictionaries["en"] = hunspell.Hunspell("resources/hunspell_dictionaries/en_US.aff", "resources/hunspell_dictionaries/en_US.dic")

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
		fmt.Printf("--%v\n", index)
		// Convert all input languages to the away language
		for _, lang := range inputLanguages {
			if (dictionaries[lang].Spell(index)) {
				if (lang != *awayLanguageString) {
					fmt.Printf("Translating\n")
					translation := GoogleWord(googleClient, ctx, index, lang, *awayLanguageString)
					fmt.Printf("Translated\n")
					if (len(translation) > 0) {
						index = translation
					} else {
						continue
					}
				}
				fmt.Printf("Going to lemmatize and add it to the wordMap\n")
				if _, ok := wordMap[index]; !ok { // TODO maybe we should lookup the word
					// Now we need to lemmatize it TODO
					// check to see if the lemmatize version is in wordmap TODO
					fmt.Printf("Adding: %v\n", index)
					wordMap[index] = make([]Vocab,0) // TODO do we even need to create a wordMap like this?
					// We're checking dictionary TODO
					// Adding it if it's not there TODO
					// And adding it to the deck if it is now or was TODO
				} else {
					fmt.Printf("Actually, it was already present\n")
				}
			}
		}

	}
	// We're going to get the base for each word now
		// Oxford Translate and Dump
	//	definition := oxfordClient.OxfordDefine("") // TODO: this should be a base
	_ = oxfordClient

}

func check (e error) {
	if e != nil {
		panic(e)
	}
}

