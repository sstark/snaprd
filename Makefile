BIN=	snaprd
PREFIX=	/usr/local

${BIN}: *.go Makefile
	go get github.com/daviddengcn/go-colortext
	go build -o ${BIN}

checkfmt:
	@gofmt -d *.go

test:
	env TZ=Europe/Berlin go test

install: ${BIN}
	install ${BIN} ${PREFIX}/bin

clean:
	rm -f ${BIN}
