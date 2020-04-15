package api

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"

    "github.com/jackyzha0/nanoDB/index"
    "github.com/jackyzha0/nanoDB/log"

    "github.com/julienschmidt/httprouter"
)

// Serve defines all the endpoints and starts a new http server on :3000
func Serve() {
    router := httprouter.New()

    // define endpoints
    router.GET("/", Health)
    router.GET("/index", GetIndex)
    router.POST("/regenerate", RegenerateIndex)
    router.GET("/get/:key", GetKey)
    router.GET("/get/:key/:field", GetKeyField)
    router.PUT("/put/:key", UpdateKey)
    // TODO: /{key}/{field} PATCH -- update given key's field with contents
    // TODO: /{key} UNLINK -- kind of like a soft delete, just remove it from map
    // TODO: /{key} DELETE -- hard delete, remove from map and delete actual file

    // start server
    log.Info("starting api server on port 3000")
    log.Fatal(http.ListenAndServe(":3000", router))
}

// Health is a healtcheck endpoint, always returns 200 ok 
func Health(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    log.WInfo(w, "health ok")
}

// GetIndex returns a JSON of all files in db index
func GetIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    log.Info("retrieving index")
    files := index.I.List()

    data := struct {
        Files []string `json:"files"`
    }{
        Files: files,
    }

    w.Header().Set("Content-Type", "application/json")
    jsonData, _ := json.Marshal(data)
    fmt.Fprintf(w, "%+v", string(jsonData))
}

// GetKey returns the file with that key if found, otherwise return 404
func GetKey(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    key := ps.ByName("key")
    log.Info("get key '%s'", key)

    file, ok := index.I.Lookup(key)

    // if file fetch is successful
    if ok {
        w.Header().Set("Content-Type", "application/json")
        http.ServeFile(w, r, file.ResolvePath())
        return
    }

    // otherwise write 404
    w.WriteHeader(http.StatusNotFound)
    log.WWarn(w, "key '%s' not found", key)
}

// GetKeyField
func GetKeyField(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    key := ps.ByName("key")
    field := ps.ByName("field")
    log.Info("get field '%s' in key '%s'", field, key)

    file, ok := index.I.Lookup(key)

    // if file fetch is successful
    if ok {
        // unpack bytes into map
        bytes, _ := file.GetByteArray()
        var jsonMap map[string]interface{}
        err := json.Unmarshal(bytes, &jsonMap)

        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            log.WWarn(w, "err key '%s' cannot be parsed into json: %s", key, err.Error())
            return
        }

        val, ok := jsonMap[field]
        if !ok {
            w.WriteHeader(http.StatusBadRequest)
            log.WWarn(w, "err key '%s' does not have field '%s'", key, field)
            return
        }

        // successful field get
        w.Header().Set("Content-Type", "application/json")
        jsonData, _ := json.Marshal(val)
        fmt.Fprintf(w, "%+v", string(jsonData))
        return
    }

    // otherwise write 404
    w.WriteHeader(http.StatusNotFound)
    log.WWarn(w, "key '%s' not found", key)
}

// UpdateKey creates or updates the file with that key with the request body
func UpdateKey(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    key := ps.ByName("key")
    log.Info("put key '%s'", key)
    file, ok := index.I.Lookup(key)

    // get bytes from request body
    bodyBytes, err := ioutil.ReadAll(r.Body)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        log.WWarn(w, "err reading body when key '%s': %s", key, err.Error())
        return
    }

    err = index.I.Put(file, bodyBytes)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        log.WWarn(w, "err updating key '%s': %s", key, err.Error())
        return
    }

    // file is updated
    if ok {
        log.WInfo(w, "update '%s' successful", key)
        return
    }
    log.WInfo(w, "create '%s' successful", key)
}

// RegenerateIndex rebuilds main index with saved directory
func RegenerateIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    index.I.Regenerate(index.I.Dir)
    log.WInfo(w, "regenerated index")
}