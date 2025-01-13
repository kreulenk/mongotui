build:
	@go build -v -o bin/mongotui

install: build
	 @cp ./bin/mongotui /usr/local/bin/

licenses:
	@go-licenses report ./... --template=third_party/template.tmpl > ./ACKNOWLEDGEMENTS