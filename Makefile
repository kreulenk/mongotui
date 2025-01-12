build:
	@go build \
		-v \
		-o bin/mtui

install: build
	 @cp ./bin/mtui /usr/local/bin/

licenses:
	@go-licenses report ./... --template=third_party/template.tmpl > ./ACKNOWLEDGEMENTS