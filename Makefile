BIN=	snaprd

${BIN}: *.go Makefile
	go build -race -o snaprd

checkfmt:
	@gofmt -d *.go

test:
	go test

clean:
	rm -f ${BIN}
