# mongotui

mongotui is a terminal based MongoDB client that is designed to be easy to use and fast.

![demo.gif](./docs/demo/demo.gif)

## Usage
mongotui aims to make switching from mongosh easy as it aims to have a similar flags and commands when first connecting
to the MongoDB server.

If you have a local MongoDB server running on the default port and no authentication, you can run the following command to get up and running:
```bash
mongotui localhost
```

Please explore the help menu if additional connection information is required
```bash
mongotui --help
```

## Installation
Please ensure that you have at least Go 1.23 installed on your system.

Then, you can install mongotui by running the following command:
```bash
make install
```

## Features
- 🔗 Similar connection flags/options to mongosh
- 📂 Navigate between databases/collections/documents
- 🔍 Query for specific documents
- 📄 Pagination of results
- 👁️ View an entire document
- ✏️ Edit a document using your `$EDITOR` of choice
- 🗑️ Delete a database/collection/document
