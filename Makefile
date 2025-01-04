build:
	@go build \
		-v \
		-o bin/mtui

install: build
	 @cp ./bin/mtui /usr/local/bin/
