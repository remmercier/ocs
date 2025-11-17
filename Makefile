.PHONY: build install clean

build:
	go build -o ocs .

install: build
	mkdir -p ~/.local/bin
	install -m 755 ocs ~/.local/bin/

clean:
	rm -f ocs