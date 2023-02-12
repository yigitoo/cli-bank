run:
	go run .
build-win:
	go build && .\cli-bank.exe
build-unix:
	go build && ./cli-bank
