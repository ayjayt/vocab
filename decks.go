package main

// Deck 
type Deck struct {
	id int
	words []int
	name string
	description string
	tags string
	homeLanguage string
	awayLanguage string
}

type decks map[string]Deck

// Decks is all the decks available- okay for now since we're only doing spanish decks
var Decks = make(decks, 0)

// List names of all decks
func (d decks) ListDecks() []string {
	return nil
}

// Create a new deck
func (d decks) NewDeck(name, description, tags, language string) error {

	return nil
}

func (d decks) DoesDeckExist(name string) bool {

	return false
}

func (d decks) LoadFromFile() { // Uses flag for now

}

func (d decks) SaveToFile() {

}

func (d decks) AddWord(words ...int) error {

	return nil
}

func (d decks) DeleteWord(words ...int) error {

	return nil
}

func (d decks) GetWords() []int {

	return nil
}
