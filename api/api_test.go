package api

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/jackyzha0/nanoDB/index"
	"github.com/google/go-cmp/cmp"
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

func assertJSONFileContents(t *testing.T, ind *index.FileIndex, key string, wanted map[string]interface{}) {
	f, ok := ind.Lookup(key)
	if !ok {
		t.Errorf("couldn't find key %s in index", key)
	}

	m, err := f.ToMap()
	if err != nil {
		t.Errorf("got error %+v parsing json when shouldn't have", err.Error())
	}

	if !cmp.Equal(m, wanted) {
		t.Errorf("file content %+v didn't match! wanted %+v", m, wanted)
	}
}

func assertRawFileContents(t *testing.T, ind *index.FileIndex, key string, wanted []byte) {
	f, ok := ind.Lookup(key)
	if !ok {
		t.Errorf("couldn't find key %s in index", key)
	}

	b, _ := f.GetByteArray()
	if !cmp.Equal(b, wanted) {
		t.Errorf("file content %+v didn't match! wanted %+v", string(b), string(wanted))
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

// examples to reuse
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

		_ = makeNewJSON("test1", exampleJSON)
		_ = makeNewJSON("test2", exampleJSON)
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

		makeNewJSON("test", exampleJSON)
		index.I.Regenerate()

		req, _ := http.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusOK)
		assertHTTPBody(t, rr, exampleJSON)
	})
}

func TestRegenerateIndex(t *testing.T) {
	router := httprouter.New()
	router.POST("/", RegenerateIndex)

	t.Run("test regenerate modifies index", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())
		index.I.Regenerate()

		makeNewJSON("test", exampleJSON)
		assertEmptySlice(t, index.I.List())

		// rebuild index via endpoint
		req, _ := http.NewRequest("POST", "/", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusOK)
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

		makeNewJSON("test", exampleJSON)
		index.I.Regenerate()

		req, _ := http.NewRequest("GET", "/test/no-field", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusBadRequest)
	})

	t.Run("get field of key simple value", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		_ = makeNewJSON("test", exampleJSON)
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
		index.I.Regenerate()

		req, _ := http.NewRequest("GET", "/test/field", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusOK)
		assertHTTPBody(t, rr, nested)
	})
}

func TestDeleteKey(t *testing.T) {
	router := httprouter.New()
	router.DELETE("/:key", DeleteKey)

	t.Run("delete non-existent key", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		req, _ := http.NewRequest("DELETE", "/nothinghere", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusNotFound)
	})

	t.Run("delete existing key", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		_ = makeNewJSON("test", exampleJSON)
		index.I.Regenerate()

		req, _ := http.NewRequest("DELETE", "/test", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusOK)
		assertEmptySlice(t, index.I.List())
	})
}

func TestUpdateKey(t *testing.T) {
	router := httprouter.New()
	router.PUT("/:key", UpdateKey)

	t.Run("update non-existent key", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		byteReader := mapToIOReader(exampleJSON)
		req, _ := http.NewRequest("PUT", "/something", byteReader)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusOK)
		assertSliceContains(t, index.I.List(), "something")
		assertJSONFileContents(t, index.I, "something", exampleJSON)
	})

	t.Run("update existing key", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		shortTest := map[string]interface{}{
			"qwer": "asdf",
		}

		_ = makeNewJSON("something", shortTest)
		index.I.Regenerate()
		assertJSONFileContents(t, index.I, "something", shortTest)

		byteReader := mapToIOReader(exampleJSON)
		req, _ := http.NewRequest("PUT", "/something", byteReader)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusOK)
		assertSliceContains(t, index.I.List(), "something")
		assertJSONFileContents(t, index.I, "something", exampleJSON)
	})

	t.Run("update key with non-json bytes", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		jsonBytes := []byte("non-json bytes")
		byteReader := bytes.NewReader(jsonBytes)
		req, _ := http.NewRequest("PUT", "/something", byteReader)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusOK)
		assertSliceContains(t, index.I.List(), "something")
		assertRawFileContents(t, index.I, "something", jsonBytes)
	})
}

func TestPatchKeyField(t *testing.T) {
	router := httprouter.New()
	router.PATCH("/:key/:field", PatchKeyField)

	t.Run("patch field of non-existent key", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		byteReader := mapToIOReader(exampleJSON)
		req, _ := http.NewRequest("PATCH", "/nofile/nofield", byteReader)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusNotFound)
	})

	t.Run("patch non-existent field of existing key", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		_ = makeNewJSON("test", exampleJSON)
		index.I.Regenerate()

		byteReader := mapToIOReader(exampleJSON)
		req, _ := http.NewRequest("PATCH", "/test/nofield", byteReader)
		rr := httptest.NewRecorder()

		expected := map[string]interface{}{
			"field": "value",
			"nofield": exampleJSON,
		}

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusOK)
		assertJSONFileContents(t, index.I, "test", expected)
	})

	t.Run("patch field of existing key with non-json bytes", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		_ = makeNewJSON("test", exampleJSON)
		index.I.Regenerate()

		jsonBytes := []byte("non-json bytes")
		byteReader := bytes.NewReader(jsonBytes)
		req, _ := http.NewRequest("PATCH", "/test/field", byteReader)
		rr := httptest.NewRecorder()

		expected := map[string]interface{}{
			"field": "non-json bytes",
		}

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusOK)
		assertJSONFileContents(t, index.I, "test", expected)
	})
}

func TestResolveReferences(t *testing.T) {

	firstContentWithRef := map[string]interface{} {
		"test": "testVal",
		"secondVal": "REF::second",
	}

	secondContentWithRef := map[string]interface{} {
		"just": "strings",
		"ref": "REF::third",
	}

	baseContent := map[string]interface{} {
		"key": "value",
	}

	t.Run("string with no ref should be returned as is", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		got := resolveReferences("test", 1)
		want := "test"

		assert.Equal(t, got, want)
	})

	t.Run("datatypes other than string, slice, and map are returned as is", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		got := resolveReferences(2, 1)
		want := 2

		assert.Equal(t, got, want)
	})

	t.Run("string with ref should replace the ref correctly", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		makeNewJSON("testjson", baseContent)
		index.I.Regenerate()

		got := resolveReferences("REF::testjson", 1)

		assert.Equal(t, got, baseContent)
	})

	t.Run("string with non-existent ref should return error message", func(t *testing.T) {
		got := resolveReferences("REF::nonexistent", 1)
		gotVal := reflect.ValueOf(got)

		if gotVal.Kind() != reflect.String {
			t.Errorf("the resolved value should have been a string but got type '%s'", gotVal.Kind())
		}

		assert.True(t, strings.Contains(gotVal.String(), "REF::ERR"))
	})

	t.Run("refs within a slice should all be replaced", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		makeNewJSON("testjson1", baseContent)
		makeNewJSON("testjson2", baseContent)
		index.I.Regenerate()

		refSlice := []string{"test", "REF::testjson1", "notref", "REF::testjson2"}
		got := resolveReferences(refSlice, 1)

		expectedSlice := []interface{}{"test", baseContent, "notref", baseContent}
		assert.Equal(t, got, expectedSlice)
	})

	t.Run("refs within map values should all be replaced", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		makeNewJSON("testjson1", baseContent)
		makeNewJSON("testjson2", baseContent)
		index.I.Regenerate()

		refMap := map[string]interface{} {
			"firstRef": "REF::testjson1",
			"nonRef": "nothing here",
			"secondRef": "REF::testjson2",
		}
		got := resolveReferences(refMap, 1)

		expectedMap := map[string]interface{} {
			"firstRef":  baseContent,
			"nonRef":    "nothing here",
			"secondRef": baseContent,
		}
		assert.Equal(t, got, expectedMap)
	})

	t.Run("double nested refs should be resolved when depth permits", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		makeNewJSON("first", firstContentWithRef)
		makeNewJSON("second", secondContentWithRef)
		makeNewJSON("third", baseContent)
		index.I.Regenerate()

		got := resolveReferences(firstContentWithRef, 2)

		expectedMap := map[string]interface{} {
			"test": "testVal",
			"secondVal": map[string]interface{} {
				"just": "strings",
				"ref":  baseContent,
			},
		}
		assert.Equal(t, got, expectedMap)
	})

	t.Run("double nested refs only resolve one because of depth param", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		makeNewJSON("first", firstContentWithRef)
		makeNewJSON("second", secondContentWithRef)
		makeNewJSON("third", baseContent)
		index.I.Regenerate()

		got := resolveReferences(firstContentWithRef, 1)

		expectedMap := map[string]interface{} {
			"test": "testVal",
			"secondVal": map[string]interface{} {
				"just": "strings",
				"ref":  "REF::third",
			},
		}
		assert.Equal(t, got, expectedMap)
	})
}