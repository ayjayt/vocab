package main

type Vocab struct {
	id int
	word string
	language string
	translation string
	gender string
}

func DoesWordExist(word string) bool {

	return false
}

func AddWord(word, language, translation, gender string) error {

	return nil
}

func GetWordByWord(words ...string) []Vocab {

	return nil
}

func GetWordById(words ...int) []Vocab {

	return nil
}
