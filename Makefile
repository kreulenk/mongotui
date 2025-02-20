build:
	@go build -v -ldflags="-X 'github.com/kreulenk/mongotui/internal/build.Version=makeFileBuild' -X 'github.com/kreulenk/mongotui/internal/build.SHA=makeFileBuild'" -o bin/mongotui

install: build
	 @cp ./bin/mongotui /usr/local/bin/

licenses:
	@go-licenses report ./... --template=third_party/template.tmpl > ./ACKNOWLEDGEMENTS

# See docs/demo/DEMO.md
demo-gif:
	@vhs ./docs/demo/demo.tape