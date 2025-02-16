# Mongotui

Mongotui is a terminal user interface MongoDB client that is designed to be easy to use and fast.

![demo.gif](./docs/demo/demo.gif)

## Usage
Mongotui aims to make switching from mongosh easy as it aims to have a similar flags and commands when first connecting
to the MongoDB server.

If you have a local MongoDB server running on the default port and no authentication, you can run the following command to get up and running:
```bash
mongotui localhost
```

Mongotui also accepts full MongoDB connection strings.

```bash
mongodb://user:password@localhost:27017
```


Explore the help menu if additional connection information is required.
```bash
mongotui --help
```


## Features
- 🔗 Similar connection flags/options to mongosh
- 📂 Navigate between databases/collections/documents
- 🔍 Query for specific documents
- 📄 Pagination of results
- 👁️ View an entire document
- ✏️ Edit a document using your `$EDITOR` of choice
- 🗑️ Delete a database/collection/document

## Installation

### MacOS
```bash
brew tap kreulenk/mongotui https://github.com/kreulenk/mongotui.git
brew install mongotui
```

### Linux
Navigate to the Releases section of mongotui's GitHub repository and download the latest tar for your
processor architecture. Then, untar the executable and move it to `/usr/local/bin/mongotui`.

E.g.
```
curl -OL https://github.com/kreulenk/mongotui/releases/download/v1.0.0/mongotui-linux-amd64.tar.gz
tar -xzvf mongotui-linux-amd64.tar.gz
mv ./mongotui /usr/local/bin/mongotui
```

### Build From Source

Please ensure that you have at least Go 1.23 installed on your system.

Then, install mongotui by running
```bash
make install
```
