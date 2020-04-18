# nanodb
a simple, easy, and debuggable document database for prototyping and hackathons

## what is `nanoDB`

> tldr; a document database with key-based access and reference resolution. each document is actually a JSON file on your local machine, making for easy debugging.

document-based database with key-based access and reference resolution.
schemaless
think about it like a lightweight mongodb with built-in reference resolution.

that means you can do stuff like:

does not have
aggregation framework
advanced queries
sharding
eventual consistency
this is not meant to be a production ready database!

## motivation

## endpoints
#### `/   GET` get index
#### `/   POST` regenerate index
#### `/:key   GET` get document `key`
#### `/:key   PUT` create/update document `key`
#### `/:key   DELETE` delete document `key`
#### `/:key/:field   POST` get `field` of document `key`
#### `/:key/:field   PATCH` update `field` of document `key`

## commands
```bash
nanodb help  # shows a list of commands
nanodb start # start a nanodb server on :3000 using folder `db`
nanodb shell # start an interactive nanodb shell
```

#### `nanodb start`
This command starts a new `nanodb` server which listens for requests port `:3000` and uses the default folder `db`. The API endpoints are listed [here](#markdown-header-endpoints).

You can change the directory with the `--dir <value>, -d <value>` flag.
```bash
# e.g.
nanodb --dir some/folder start # start a nanodb server using folder `some/folder`
nanodb -d . start           # start a nanodb shell in the current directory
```

You can also change the port the server is hosted on with the `--p <value>, -p <value>` flag.
```bash
# e.g.
nanodb start --port 8081  # start a nanodb server on port 8081
nanodb -d . start -p 3000 # start a nanodb server on port 3000 using current directory
```

#### `nanodb shell`
This command starts a new `nanodb` interactive shell using the defailt folder `db`. The interactive shell isn't designed to do everything the API does, rather it is more like a quick tool to explore the database by allowing easy viewing of the database index, lookup of documents, and deletion of documents. 

<img src="https://user-images.githubusercontent.com/23178940/79622428-18718d00-80cc-11ea-8fe6-b0f620131b61.gif" width="400">

Similar to the `nanodb` server, you can change the directory with the `--dir <value>, -d <value>` flag.
```bash
# e.g.
nanodb -d . shell # start a nanodb shell using current directory
```

## running `nanoDB`
#### from source
0. `git clone https://github.com/jackyzha0/nanoDB.git`
1. `go install github.com/jackyzha0/nanoDB`
2. `nanodb`

#### via docker
0. TODO

## building `nanoDB` from source
0. `git clone https://github.com/jackyzha0/nanoDB.git`
1. `go build -o nanodb .`