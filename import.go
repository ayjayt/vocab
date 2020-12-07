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
//var inputLanguageString = flag.String("input_languages", "es,en", "Specify the languages int he document")
var inputLanguageString = flag.String("input_languages", "es", "Specify the languages int he document")
var wordRegexp = regexp.MustCompile(`\s*['\-\pL]['\-\pL]['\-\pL]+[,.!?;]*\s*`)
var barewordRegexp = regexp.MustCompile(`\s*['\-\pL]['\-\pL]['\-\pL]+\s*`)

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

type UltraLinguaClient struct {
	http.Client
}

// UltraLinguaClient gets a response about definitions from an oxford dictionary
func (client UltraLinguaClient) ULLemma(word string) []string {
	fmt.Printf("https://api.ultralingua.com/api/2.0/lemmas/es/"+word+"?key=PLPPRTU8J8QE3BDPJAL6XDBN\n")
	req, err := http.NewRequest("GET", "https://api.ultralingua.com/api/2.0/lemmas/es/"+word+"?key=PLPPRTU8J8QE3BDPJAL6XDBN", nil)
	check(err)
	resp, err := client.Do(req)
	check(err)
	defer resp.Body.Close()
	var jsonData interface{}
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		check(err)
		//json.NewDecoder(resp.Body).Decode(&jsonData)
		//b, err := json.MarshalIndent(jsonData, "", "  ")
		err = json.Unmarshal(bodyBytes, &jsonData)
		//check(err)
		//fmt.Printf("jsonData: %+v\n", jsonData)
		// Lets go through this map and find all "root"
		fmt.Printf("Body Bytes:\n%+v\n", string(bodyBytes))
		fmt.Printf("JSON Data: \n%+v\n", jsonData)
		roots := make(map[string]bool,0)
		// TODO: needs to deeper
		for _, val := range jsonData.([]interface{}) {
			if valMap, ok := val.(map[string]interface{}); ok {
				root, ok := valMap["root"]; if ok {
					fmt.Printf("Found: %v\n", string(root.(string)))
					roots[string(root.(string))] = true
				}
			}
		}
// loop through json data
		keys := make([]string, 0, len(roots))
		for k := range roots {
				keys = append(keys, k)
		}
		return keys// TODO: still returning whole response
	}
	return nil
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

// ReadFile right now reads a global file var set by flag- for inputs
func ReadFile() (map[string]bool, []string) {
	inputText, err := ioutil.ReadFile(*inputFileString)
	check(err)
	allWordsFromRaw := wordRegexp.FindAllString(string(inputText), -1)
	if allWordsFromRaw == nil {
		return nil, nil
	}
	wordMapRaw := make(map[string]bool, len(allWordsFromRaw))
	for _, val := range allWordsFromRaw {
		wordMapRaw[strings.ToLower(strings.TrimSpace(val))] = true
	}
	return wordMapRaw, allWordsFromRaw
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
	ultraLinguaClient := &UltraLinguaClient{}

	// This is the setup for google translate API
	ctx := context.Background()
	googleClient, err := translate.NewClient(ctx)
	if err != nil {
		panic("translate.NewClient: " + err.Error())
	}
	defer googleClient.Close()

	// Start work
	wordMapRaw, wordSlice := ReadFile() // a map of potential words to a bool
	wholeTextSanitized := make([]string, 0, len(wordSlice))
	for _, word := range wordSlice {
		bareword := strings.TrimSpace(barewordRegexp.FindString(word))
		for _, lang := range inputLanguages {
			if (dictionaries[lang].Spell(bareword)) { // if it is a properly spelled word in some language
				if (lang != *awayLanguageString) {
					continue
				}
				wholeTextSanitized = append(wholeTextSanitized, word)
			}
		}
	}
	fmt.Printf("Whole text sanitized: \n%v\n", wholeTextSanitized)

	// Iterate over words
	for index, _ := range wordMapRaw {
		fmt.Printf("--%v\n", index)
		// Convert all input languages to the away language
		for _, lang := range inputLanguages {
			if (dictionaries[lang].Spell(index)) { // if it is a properly spelled word in some language
				if (lang != *awayLanguageString) { // if it's an input other than the language we're learning (then do a weak translation to what we're learning)
					fmt.Printf("Translating\n")
					translation := GoogleWord(googleClient, ctx, index, lang, *awayLanguageString) // translate it
					fmt.Printf("Translated\n")
					if (len(translation) > 0) {
						index = translation
					} else {
						continue // translation yielded nothing, continue as if if dictionaries[lang].Spell() == False
					}
				}
				fmt.Printf("Going to lemmatize and add it to the wordMap\n")
				if _, ok := wordMap[index]; !ok { // TODO maybe we should lookup the word
					// HARDCODING SPANSIH
					// HARDCODING KEY ?key=PLPPRTU8J8QE3BDPJAL6XDBN
					//definition := ultraLinguaClient.ULLemma(index) // TODO: this should be a base
					//fmt.Printf("Lemmas:\n%+v\n", definition)
					_= ultraLinguaClient
					// Now we need to lemmatize it TODnPO
					// check to see if the lemmatize version is in wordmap TODO

					// so now we have the index
					// and we have the lemmas

					// we should make a map of indices-->bases
					// bases--> definitions


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


