package api

import (
    "fmt"
    "net/http"
    "github.com/jackyzha0/nanoDB/index"

    log "github.com/sirupsen/logrus"

    "github.com/julienschmidt/httprouter"
)

func Serve() {
    router := httprouter.New()

    // define endpoints
    router.GET("/", GetIndex)
    router.POST("/regenerate", RegenerateIndex)
    router.GET("/get/:key", GetKey)

    // start server
    log.Info("starting api server on port 3000")
    log.Fatal(http.ListenAndServe(":3000", router))
}

func GetIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    files := index.I.List()
    fmt.Fprintf(w, "%+v", files)
}

func GetKey(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    key := ps.ByName("key")
    log.Infof("attempt to get key %s", key)

    file, ok := index.I.Lookup(key)
    fmt.Fprintf(w, "%+v, %t", file.FileName, ok)
}

func RegenerateIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    fmt.Fprintf(w, "hit regenerate index")
    index.I.Regenerate(index.I.Dir)
}
