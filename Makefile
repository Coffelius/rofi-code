BINARY_NAME = rofi-code

build:
	go build -v

run: build
	./$(BINARY_NAME) 

install: build
	sudo cp ./$(BINARY_NAME) /usr/bin/

lint:
	revive -formatter friendly -config revive.toml ./... 

tidy:
	go mod tidy

# (build but with a smaller binary)
dist:
	go build -ldflags="-w -s" -gcflags=all=-l -v

# (even smaller binary)
pack: dist
	upx ./$(BINARY_NAME)