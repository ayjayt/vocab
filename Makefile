export GOOGLE_APPLICATION_CREDENTIALS=$(PWD)/keys/vocab-32f12887b205.json

all:
	go build && ./vocab -input test_input/gibberish.txt
