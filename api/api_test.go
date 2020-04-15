package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jackyzha0/nanoDB/index"
	"github.com/julienschmidt/httprouter"
	af "github.com/spf13/afero"
)

func assertHTTPStatus(t *testing.T, rr *httptest.ResponseRecorder, status int) {
	t.Helper()
	got := rr.Code
	if got != http.StatusOK {
		t.Errorf("returned wrong status code: got %+v, wanted %+v", got, status)
	}
}

func assertHTTPBody(t *testing.T, rr *httptest.ResponseRecorder, expected string) {
	t.Helper()
	if rr.Body.String() != expected {
		t.Errorf("returned unexpected body: got %+v want %+v", rr.Body.String(), expected)
	}
}

func checkJSONEquals(t *testing.T, a interface{}, b interface{}) {
	t.Helper()
	if fmt.Sprintf("%+v", a) != fmt.Sprintf("%+v", b) {
		t.Errorf("got %+v, want %+v", a, b)
	}
}

func makeNewJSON(name string, contents map[string]interface{}) *index.File {
	jsonData, _ := json.Marshal(contents)
	af.WriteFile(index.I.FileSystem, name+".json", jsonData, 0644)
	return &index.File{FileName: name}
}

func TestMain(m *testing.M) {
	index.I = index.NewFileIndex("")
	exitVal := m.Run()
	os.Exit(exitVal)
}

func TestGetIndex(t *testing.T) {
	t.Run("get empty index", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		router := httprouter.New()
		router.GET("/", GetIndex)

		req, _ := http.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusOK)
		assertHTTPBody(t, rr, `{"files":null}`)
	})

	t.Run("get index with files", func(t *testing.T) {
		index.I.SetFileSystem(af.NewMemMapFs())

		// add some dummy json files
		expected := map[string]interface{}{
			"field": "value",
		}

		_ = makeNewJSON("test1", expected)
		_ = makeNewJSON("test2", expected)

		// rebuild index
		index.I.Regenerate()

		router := httprouter.New()
		router.GET("/", GetIndex)

		req, _ := http.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)
		assertHTTPStatus(t, rr, http.StatusOK)
		assertHTTPBody(t, rr, `{"files":["test1", "test2"]}`)
	})
}
