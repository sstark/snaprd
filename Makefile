BIN=	snaprd
PREFIX=	/usr/local

${BIN}: *.go Makefile
	go build -o snaprd

checkfmt:
	@gofmt -d *.go

test:
	go test

install: ${BIN}
	install ${BIN} ${PREFIX}/bin

clean:
	rm -f ${BIN}
