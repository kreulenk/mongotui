# Mongotui

<p>
    <a href="https://github.com/kreulenk/mongotui/releases"><img src="https://img.shields.io/github/release/kreulenk/mongotui.svg" alt="Latest Release"></a>
    <a href="https://goreportcard.com/report/github.com/kreulenk/mongotui"><img src="https://goreportcard.com/badge/github.com/kreulenk/mongotui"></a>
</p>

Mongotui is a terminal user interface MongoDB client that is designed to be easy to use and fast.

![demo.gif](./docs/demo/demo.gif)

## Usage
Mongotui aims to make switching from mongosh easy as it has similar flags and commands when first connecting
to a MongoDB server.

If you have a local MongoDB server running on the default port and no authentication, you can run the following command to get up and running.
```bash
mongotui localhost
```

Mongotui also accepts full MongoDB connection strings.

```bash
mongotui mongodb://user:password@localhost:27017
```


Explore the help menu if additional connection information is required.
```bash
mongotui --help
```


## Features
- 🔗 Similar connection flags/options to mongosh
- 📂 Navigate between databases/collections/documents
- 🗂 Filter displayed databases/collections
- 🔍 Query for specific documents
- 📄 Pagination of document results
- 👁️ View an entire document
- ➕ Insert a new document
- ✏️ Edit a document using your `$EDITOR` of choice
- 🗑️ Drop databases/collections and delete documents

## Installation

### MacOS
If you have HomeBrew installed, use the tap shown below.

```bash
brew tap kreulenk/brew
brew install mongotui
```

### Linux
Navigate to the Releases section of mongotui's GitHub repository and download the latest tar for your
processor architecture. Then, untar the executable and move it to `/usr/local/bin/mongotui`.

E.g.
```
curl -OL https://github.com/kreulenk/mongotui/releases/download/v1.3.0/mongotui-linux-amd64.tar.gz
tar -xzvf mongotui-linux-amd64.tar.gz
mv ./mongotui /usr/local/bin/mongotui
```

### Build From Source

Ensure that you have at least Go 1.23 installed on your system.

Then, run
```bash
make install
```
