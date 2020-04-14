# nanoDB
a simple, easy, and stupid database for prototyping and hackathons

## endpoints
* `get /` -- healthcheck
* `get /index` -- lists all current files in index
* `get /get/:key` -- attempts to get file with key
* `post /regenerate` -- rebuilds file index using cli dir