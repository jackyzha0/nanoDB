package api

import (
    "encoding/json"
    "fmt"
    "net/http"

    "github.com/jackyzha0/nanoDB/index"

    log "github.com/sirupsen/logrus"

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
    // TODO: /{key} PUT -- set value of that key to contents
    // TODO: /{key}/{field} PATCH -- update given key's field with contents
    // TODO: /{key} UNLINK -- kind of like a soft delete, just remove it from map
    // TODO: /{key} DELETE -- hard delete, remove from map and delete actual file

    // start server
    log.Info("starting api server on port 3000")
    log.Fatal(http.ListenAndServe(":3000", router))
}

// Health is a healtcheck endpoint, always returns 200 ok 
func Health(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    fmt.Fprint(w, "ok")
}

// GetIndex returns a JSON of all files in db index
func GetIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    w.Header().Set("Content-Type", "application/json")
    files := index.I.List()

    data := struct {
        Files []string `json:"files"`
    }{
        Files: files,
    }

    jsonData, _ := json.Marshal(data)
    fmt.Fprintf(w, "%+v", string(jsonData))
}

// GetKey returns the file with that key if found, otherwise return 404
func GetKey(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    key := ps.ByName("key")
    log.Infof("get key '%s'", key)

    file, ok := index.I.Lookup(key)

    // if file fetch is successful
    if ok {
        w.Header().Set("Content-Type", "application/json")
        http.ServeFile(w, r, file.ResolvePath())
        return
    }

    // otherwise write 404
    w.WriteHeader(http.StatusNotFound)
    fmt.Fprintf(w, "key '%s' not found", key)
    log.Warnf("key '%s' not found", key)
}

// RegenerateIndex rebuilds main index with saved directory
func RegenerateIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    fmt.Fprintf(w, "hit regenerate index")
    index.I.Regenerate(index.I.Dir)
}
