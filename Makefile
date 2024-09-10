SRC=$(shell find -type f -name "*.go")

bin/hcvelo: $(SRC)
	go build -o bin/hcvelo main.go
	
