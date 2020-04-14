package api

import (
    "fmt"
    "net/http"
    "github.com/jackyzha0/nanoDB/index"

    log "github.com/sirupsen/logrus"

    "github.com/julienschmidt/httprouter"
)

func init() {
    router := httprouter.New()

    // define endpoints
    router.GET("/", GetIndex)
    router.POST("/regenerate", RegenerateIndex)
    router.GET("/get/:key", GetKey)

    // start server
    log.Info("Starting server on port 3000...")
    log.Fatal(http.ListenAndServe(":3000", router))
}

func GetIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    fmt.Fprintf(w, "hit get index")
}

func GetKey(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    key := ps.ByName("key")
    fmt.Fprintf(w, fmt.Sprintf("attempt to get key %s", key))
}

func RegenerateIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    fmt.Fprintf(w, "hit regenerate index")
    index.CrawlDirectory(".")
}
