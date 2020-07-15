default: bin/kbar

.PHONY: bin/kbar

bin:
	mkdir -p $@

bin/kbar: bin
	go build -o $@

install: bin/kbar
	cp bin/kbar ${HOME}/bin/
