package api

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/jackyzha0/nanoDB/index"
	"github.com/julienschmidt/httprouter"
	af "github.com/spf13/afero"
)

func assertHTTPStatus(t *testing.T, rr *httptest.ResponseRecorder, status int) {
	t.Helper()
	got := rr.Code
	if got != status {
		t.Errorf("returned wrong status code: got %+v, wanted %+v", got, status)
	}
}

func assertHTTPContains(t *testing.T, rr *httptest.ResponseRecorder, expected []string) {
	t.Helper()
	for _, v := range expected {
		if !strings.Contains(rr.Body.String(), v) {
			t.Errorf("couldn't find %s in body %+v", v, rr.Body.String())
		}
	}
}

func assertSliceContains(t *testing.T, list []string, s string) {
	found := false
	for _, v := range list {
		if v == s {
			found = true
		}
	}

	if !found {
		t.Errorf("slice %+v didn't contain %s", list, s)
	}
}

func assertEmptySlice(t *testing.T, list []string) {
	if len(list) != 0 {
		t.Errorf("slice %+v wasnt empty", list)
	}
}

func assertHTTPBody(t *testing.T, rr *httptest.ResponseRecorder, expected map[string]interface{}) {
	t.Helper()
	resp := rr.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	var parsedJSON map[string]interface{}
	err := json.Unmarshal(body, &parsedJSON)

	if err != nil {
		t.Errorf("got an error parsing json when shouldn't have")
	}

	if !reflect.DeepEqual(parsedJSON, expected) {
		t.Errorf("json mismatched, got: %+v, want: %+v", parsedJSON, expected)
	}
}

func makeNewJSON(name string, contents map[string]interface{}) *index.File {
	jsonData, _ := json.Marshal(contents)
	af.WriteFile(index.I.FileSystem, name+".json", jsonData, 0644)
	return &index.File{FileName: name}
}

func mapToIOReader(m map[string]interface{}) io.Reader {
	jsonData, _ := json.Marshal(m)
	return bytes.NewReader(jsonData)
}

func TestMain(m *testing.M) {
	index.I = index.NewFileIndex(".")

	exitVal := m.Run()
	os.Exit(exitVal)
}

var exampleJSON = map[string]interface{}{
	"field": "value",
}

func TestGetIndex(t *testing.T) {
	router := httprouter.New()
	router.GET("/", GetIndex)

	t.Run("get empty index", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		req, _ := http.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusOK)
		assertHTTPBody(t, rr, map[string]interface{}{
			"files": nil,
		})
	})

	t.Run("get index with files", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		// add some dummy json files
		_ = makeNewJSON("test1", exampleJSON)
		_ = makeNewJSON("test2", exampleJSON)

		// rebuild index
		index.I.Regenerate()

		req, _ := http.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusOK)
		assertHTTPContains(t, rr, []string{"test1", "test2"})
	})
}

func TestGetKey(t *testing.T) {
	router := httprouter.New()
	router.GET("/:key", GetKey)

	t.Run("get non-existent file", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		req, _ := http.NewRequest("GET", "/nothinghere", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusNotFound)
	})

	t.Run("get file", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		// add some dummy json files
		makeNewJSON("test", exampleJSON)

		// rebuild index
		index.I.Regenerate()

		req, _ := http.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusOK)
		assertHTTPBody(t, rr, map[string]interface{}{
			"field": "value",
		})
	})
}

func TestRegenerateIndex(t *testing.T) {
	router := httprouter.New()
	router.POST("/", RegenerateIndex)

	t.Run("test regenerate modifies index", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		// rebuild index
		index.I.Regenerate()

		// add some dummy json files without rebuilding
		makeNewJSON("test", exampleJSON)

		// make sure no files in index
		assertEmptySlice(t, index.I.List())

		// rebuild index via endpoint
		req, _ := http.NewRequest("POST", "/", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusOK)

		// get list
		assertSliceContains(t, index.I.List(), "test")
	})
}

func TestGetKeyField(t *testing.T) {
	router := httprouter.New()
	router.GET("/:key/:field", GetKeyField)

	t.Run("get field of non-existent key", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		req, _ := http.NewRequest("GET", "/nothinghere/stillnothing", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusNotFound)
	})

	t.Run("get non-existent field of key", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		// add some dummy json files
		makeNewJSON("test", exampleJSON)

		// rebuild index
		index.I.Regenerate()

		req, _ := http.NewRequest("GET", "/test/no-field", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusBadRequest)
	})

	t.Run("get field of key simple value", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		// add some dummy json files
		_ = makeNewJSON("test", exampleJSON)

		// rebuild index
		index.I.Regenerate()

		req, _ := http.NewRequest("GET", "/test/field", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusOK)
		assertHTTPContains(t, rr, []string{"value"})
	})

	t.Run("get field of key nested val", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		// add some dummy json files
		nested := map[string]interface{}{
			"more_fields": "yay",
			"nested_thing": map[string]interface{}{
				"f": "asdf",
			},
		}

		expected := map[string]interface{}{
			"field":       nested,
			"other_field": "yeet",
		}

		_ = makeNewJSON("test", expected)

		// rebuild index
		index.I.Regenerate()

		req, _ := http.NewRequest("GET", "/test/field", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusOK)
		assertHTTPBody(t, rr, nested)
	})
}

func TestDeleteKey(t *testing.T) {
	t.Run("delete non-existent key", func(t *testing.T) {

	})

	t.Run("delete existing key", func(t *testing.T) {

	})
}

func TestUpdateKey(t *testing.T) {
	// requires use of mapToIOReader
	t.Run("update non-existent key", func(t *testing.T) {

	})

	t.Run("update existing key", func(t *testing.T) {

	})

	t.Run("update key with non-json bytes", func(t *testing.T) {

	})
}

func TestPatchKeyField(t *testing.T) {
	// requires use of mapToIOReader
	t.Run("patch field of non-existent key", func(t *testing.T) {

	})

	t.Run("patch non-existent field of existing key", func(t *testing.T) {

	})

	t.Run("patch field of existing key with non-json bytes", func(t *testing.T) {

	})

	t.Run("patch field of existing key with flat field", func(t *testing.T) {

	})

	t.Run("patch field of existing key with nested field", func(t *testing.T) {

	})
}
