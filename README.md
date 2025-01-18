# mongotui

mongotui is a terminal based MongoDB client that is designed to be easy to use and fast.

## Usage
mongotui aims to make switching from mongosh easy as it aims to have a similar flags and commands when first connecting
to the MongoDB server. From there, you can enjoy the ease of use of a terminal based UI.

![demo.gif](./docs/demo/demo.gif)

If you have a local MongoDB server running on the default port and no authentication, you can run the following command to get up and running:
```bash
mongotui localhost
```

Please explore the help menu if additional connection information is required
```bash
mongotui --help
```

## Installation
Please ensure that you have at least Go 1.24 installed on your system.

Then, you can install mongotui by running the following command:
```bash
make install
```

## Development Roadmap
The high level plan for this project before the v1.0 release is as follows:
|  #  | Step                                                            | Status |
| :-: | --------------------------------------------------------------- | :----: |
|  1  | Simple PLAIN authentication to a MongoDB server                 |   ✅   |
|  2  | View a server's databases, collections, and documents summaries |   ✅   |
|  3  | Query within a collection                                       |   ✅   |
|  4  | View an entire document                                         |   ✅   |
|  5  | Edit a document                                                 |   ✅   |
|  6  | Delete a database/collection/document                           |   ❌   |
|  7  | Connection/authentication option feature parity with mongosh    |   ❌   |
|  8  | TBD!                                                            |   💥   |
