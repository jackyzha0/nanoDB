package api

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "reflect"
    "strconv"
    "strings"

    "github.com/jackyzha0/nanoDB/index"
    "github.com/jackyzha0/nanoDB/log"
    "github.com/julienschmidt/httprouter"
)

// GetIndex returns a JSON of all files in db index
func GetIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    log.Info("retrieving index")
    files := index.I.List()

    // create temporary struct with index data
    data := struct {
        Files []string `json:"files"`
    }{
        Files: files,
    }

    // create json representation and return
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

        // unpack bytes into map
        jsonMap, err := file.ToMap()
        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            log.WWarn(w, "err key '%s' cannot be parsed into json: %s", key, err.Error())
            return
        }

        // successful field get
        w.Header().Set("Content-Type", "application/json")
        maxDepth := getMaxDepthParam(r)
        resolvedJsonMap := resolveReferences(jsonMap, maxDepth)

        jsonData, _ := json.Marshal(resolvedJsonMap)
        fmt.Fprintf(w, "%+v", string(jsonData))
        return
    }

    // otherwise write 404
    w.WriteHeader(http.StatusNotFound)
    log.WWarn(w, "key '%s' not found", key)
}

// GetKeyField returns key's field, 404 if not found
func GetKeyField(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    key := ps.ByName("key")
    field := ps.ByName("field")
    log.Info("get field '%s' in key '%s'", field, key)

    file, ok := index.I.Lookup(key)

    // if file fetch is successful
    if ok {
        // unpack bytes into map
        jsonMap, err := file.ToMap()
        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            log.WWarn(w, "err key '%s' cannot be parsed into json: %s", key, err.Error())
            return
        }

        // lookup value
        val, ok := jsonMap[field]
        if !ok {
            w.WriteHeader(http.StatusBadRequest)
            log.WWarn(w, "err key '%s' does not have field '%s'", key, field)
            return
        }

        // successful field get
        w.Header().Set("Content-Type", "application/json")
        maxDepth := getMaxDepthParam(r)
        resolvedValue := resolveReferences(val, maxDepth)

        jsonData, _ := json.Marshal(resolvedValue)
        fmt.Fprintf(w, "%+v", string(jsonData))
        return
    }

    // otherwise write 404
    w.WriteHeader(http.StatusNotFound)
    log.WWarn(w, "key '%s' not found", key)
}

// try to find recursive depth param or else return a default
func getMaxDepthParam(r *http.Request) int {
    maxDepth := 3

    maxDepthStr := r.URL.Query().Get("depth")
    parsedInt, err := strconv.Atoi(maxDepthStr)
    if err == nil {
        maxDepth = parsedInt
    }

    return maxDepth
}

// determines if there are any key references nested within a json
// if found, replace the references with the corresponding value
func resolveReferences(jsonVal interface{}, depthLeft int) interface{} {
    // if max recursive depth is exceeded, return as is
    if depthLeft < 1 {
        return jsonVal
    }

    val := reflect.ValueOf(jsonVal)

    switch val.Kind() {
    case reflect.String:
        valString := val.String()

        // if value is reference to another key
        if strings.Contains(valString, "REF::") {
            resolvedString := resolveString(valString, depthLeft)
            return resolvedString
        }

        // if not a ref, keep as is
        return valString

    case reflect.Slice:
        numberOfValues := val.Len()
        newSlice := make([]interface{}, numberOfValues)

        // for each value in the slice, try to resolve it recursively
        for i := 0; i < numberOfValues; i++ {
            pointer := val.Index(i)
            newSlice[i] = resolveReferences(pointer.Interface(), depthLeft)
        }

        return newSlice

    case reflect.Map:
        newMap := make(map[string]interface{})

        // for each value in the map, try to resolve it recursively
        for _, key := range val.MapKeys() {
            nestedVal := val.MapIndex(key).Interface()
            newMap[key.String()] = resolveReferences(nestedVal, depthLeft)
        }

        return newMap

    default:
        return jsonVal
    }
}

// resolves a single string that has a reference in it
func resolveString(valString string, depthLeft int) interface{} {
    key := strings.Replace(valString, "REF::", "", 1)
    file, ok := index.I.Lookup(key)

    // if key found, get contents
    if ok {
        // change bytes into map
        jsonMap, err := file.ToMap()
        if err != nil {
            errMessage := fmt.Sprintf("REF::ERR key '%s' cannot be parsed into json: %s", key, err.Error())
            return errMessage
        }
        return resolveReferences(jsonMap, depthLeft - 1)
    }

    // if key not found
    return fmt.Sprintf("REF::ERR key '%s' not found", key)
}

// PatchKeyField modifies the field of a key
func PatchKeyField(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    key := ps.ByName("key")
    field := ps.ByName("field")
    log.Info("patch field '%s' in key '%s'", field, key)

    // get bytes from request body
    bodyBytes, err := ioutil.ReadAll(r.Body)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        log.WWarn(w, "err reading body with key '%s': %s", key, err.Error())
        return
    }

    // lookup file by key
    file, ok := index.I.Lookup(key)
    // if file fetch is successful
    if ok {
        // unpack bytes into map
        jsonMap, err := file.ToMap()
        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            log.WWarn(w, "err key '%s' cannot be parsed into json: %s", key, err.Error())
            return
        }

        // set field value to parsed json
        var parsedJSON map[string]interface{}
        err = json.Unmarshal(bodyBytes, &parsedJSON)
        if err != nil {
            // not JSON, set field to string val instead
            jsonMap[field] = string(bodyBytes)
        } else {
            jsonMap[field] = parsedJSON
        }

        // remarshal to json
        jsonData, _ := json.Marshal(jsonMap)

        // write to file
        err = file.ReplaceContent(string(jsonData))
        if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            log.WWarn(w, "err setting content of key '%s': %s", key, err.Error())
            return
        }

        w.WriteHeader(http.StatusOK)
        log.WInfo(w, "patch field '%s' of key '%s' successful", field, key)
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

    // update index
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
    index.I.Regenerate()
    log.WInfo(w, "regenerated index")
}

// DeleteKey deletes the file associated with the given key, returns 404 if not found
func DeleteKey(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    key := ps.ByName("key")
    log.Info("delete key '%s'", key)
    file, ok := index.I.Lookup(key)

    // if file found delete it
    if ok {
        err := index.I.Delete(file)
        if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            log.WWarn(w, "err unable to delete key '%s': '%s'", key, err.Error())
            return
        }

        log.WInfo(w, "delete '%s' successful", key)
        return
    }

    // else state not found
    w.WriteHeader(http.StatusNotFound)
    log.WWarn(w, "key '%s' does not exist", key)
}
