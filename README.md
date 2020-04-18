# nanodb
a simple, easy, and debuggable document database for prototyping and hackathons

## what is `nanoDB`

> tldr; `nanoDB` is a document database with key-based access and reference resolution served through a REST API. each document is actually a JSON file on your local machine, making for easy debugging.

`nanoDB` is designed to be a JSON document-based database that relies on key based access to achieve `O(1)` access time. In addition, fields can hold references to other documents, which are automatically resolved up to a certain depth on retrieval. All of these documents are stored as actual JSON files on the local machine, allowing developers to easily read, debug, and modify the data without the need for external tools. 

**You can think of it like `Redis` but with `MongoDB` style documents &mdash; all of which is on-disk, human-readable, and through a REST API.**

That means you can do stuff like
* make ID based authentication services
* store user data
* simple application cache
* and much much more, without the hassle of setting up an entire database schema and having to deal with drivers!

*However*, `nanoDB` does not have any aggregation frameworks, advanced queries, sharding, or support for storage distribution. It was not created with the intention of ever being a production ready database, and should not be used as such!

## motivation
`nanoDB` arose out of many frustrations that we've personally come across when prototyping.
1. it sucks when you can't actually see/modify the documents when you're working with without the use of something external like `MongoDB Explorer` or `SQLPro`
2. I don't want to install drivers just to make queries and edit a database! Why can't this just be an API call?
3. Trying to resolve references to other documents in `noSQL` databases is a pain

As a result, we've devloped `nanoDB` to adhere to 3 key principles.

#### key principles
* easy to lookup &mdash; key-based lookup in `O(1)` time
* easy to debug &mdash; all documents are JSON files which are human readable
* easy to deploy &mdash; single binary with no dependencies. no language specific drivers needed!

## endpoints
#### `/   GET`
```bash
# get all files in database index
curl localhost:3000/

# example output on 200 OK
# > {"files":["test","test2","test3"]}
```

#### `/   POST`
```bash
# manually regenerate index
# shouldn't need to be done as each operation should update index on its own
curl -X POST localhost:3000/

# example output on 200 OK
# > regenerated index
```

#### `/:key   GET`
```bash
# get document with key `key`
curl localhost:3000/key

# example output on 200 OK (found key)
# > {"example_field": "example_value"}
# example output on 404 NotFound (key not found)
# > key 'key' not found
```

#### `/:key   PUT`
```bash
# creates document `key` if it doesn't exist
# otherwise, replaces content of `key` with given
curl -X PUT -H "Content-Type: application/json" \
            -d '{"key1":"value"}' localhost:3000/key

# example output on 200 OK (create/update success)
# > create 'key' successful
```

#### `/:key   DELETE` delete document `key`
```bash
# deletes document `key`
curl -X DELETE localhost:3000/key

# example output on 200 OK (delete success)
# > delete 'key' successful
# example output on 404 NotFound (key not found)
# > key 'key' doest not exist
```

#### `/:key/:field   GET`
```bash
# get `example_field` of document `key`
curl localhost:3000/key/example_field

# example output on 200 OK (found field)
# > "example_value"
# example output on 400 BadRequest (field not found)
# > err key 'key' does not have field 'example_field'
# example output on 404 NotFound (key not found)
# > key 'key' not found
```
#### `/:key/:field   PATCH`
```bash
# update `field` of document `key` with content
# if field doesnt exist, create it
curl -X PATCH -H "Content-Type: application/json" \
              -d '{"nested":"json!"}' \
              localhost:3000/key/example_field

# example output on 200 OK (found field)
# > patch field 'example_field' of key 'key' successful
# example output on 404 NotFound (key not found)
# > key 'key' not found
```

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

## reference resolution
You can refer to other documents by using a reference of the form `REF::<key>`. For example, with the following two JSONs:
#### `ref.json`
```json
{
  "key": "REF::nested"
}
```

#### `nested.json`
```json
{
  "nestedKey": "nestedVal"
}
```
You end up with the following:
```json
{
   "key": {
      "nestedKey": "nestedVal"
   }
}
```
This can be done within arrays and maps, to any arbitrary depth for which references should be resolved! The API has a default resolving depth of 3 while the CLI has a default of 0 but this can be explicitly changed if needed. 
## running `nanoDB`
#### from source
0. `git clone https://github.com/jackyzha0/nanoDB.git`
1. `go install github.com/jackyzha0/nanoDB`
2. `nanodb`

#### via docker
0. `docker pull jzhao2k19/nanodb:latest`
1. `docker run -p 3000:3000 jzhao2k19/nanodb:latest` # change -p 3000:3000 to different port if necessary

## building `nanoDB` from source
0. `git clone https://github.com/jackyzha0/nanoDB.git`
1. `make build`
2. (optional) for cross-platform builds, run `make build-all`