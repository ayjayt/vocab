all:
	go build && ./vocab -map maps/ayjay_t -input raw/tacosubtitles.txt
