package main

import (
    "fmt"
    "net/http"

    log "github.com/sirupsen/logrus"

    "github.com/julienschmidt/httprouter"
)

func GetIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    fmt.Fprintf(w, "hit get index")
}

func GetKey(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    key := ps.ByName("key")
    fmt.Fprintf(w, fmt.Sprintf("attempt to get key %s", key))
}

func RegenerateIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    fmt.Fprintf(w, "hit regenerate index")
}

func main() {
    router := httprouter.New()
    router.GET("/", GetIndex)
    router.PATCH("/", RegenerateIndex)
    router.GET("/:key", GetKey)

    log.Info("Starting server on port 3000...")
    log.Fatal(http.ListenAndServe(":3000", router))
}
