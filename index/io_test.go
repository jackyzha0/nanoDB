package index

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	af "github.com/spf13/afero"
)

func checkDeepEquals(t *testing.T, a interface{}, b interface{}) {
	t.Helper()
	if !cmp.Equal(a, b) {
		t.Errorf("got %+v, want %+v", a, b)
	}
}

func checkJSONEquals(t *testing.T, a interface{}, b interface{}) {
	t.Helper()
	if fmt.Sprintf("%+v", a) != fmt.Sprintf("%+v", b) {
		t.Errorf("got %+v, want %+v", a, b)
	}
}

func assertNilErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("got error %s when shouldn't have", err.Error())
	}
}

func assertErr(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Errorf("didnt get error when wanted error")
	}
}

func assertFileExists(t *testing.T, fs af.Fs, filePath string) {
	t.Helper()
	if _, err := fs.Stat(filePath+".json"); os.IsNotExist(err) {
		t.Errorf("didnt find file at %s when should have", filePath)
	}
}

func assertFileDoesNotExist(t *testing.T, fs af.Fs, filePath string) {
	t.Helper()
	if _, err := fs.Stat(filePath+".json"); err == nil {
		t.Errorf("found file at %s when shouldn't have", filePath)
	}
}

func makeNewFile(fs af.Fs, name string, contents string) {
	af.WriteFile(fs, name, []byte(contents), 0644)
}

func makeNewJSON(fs af.Fs, name string, contents map[string]interface{}) *File {
	jsonData, _ := json.Marshal(contents)
	af.WriteFile(fs, name+".json", jsonData, 0644)
	return &File{FileName: name}
}

func mapToString(contents map[string]interface{}) string {
	jsonData, _ := json.Marshal(contents)
	return string(jsonData)
}

func TestCrawlDirectory(t *testing.T) {

	t.Run("crawl empty directory", func(t *testing.T) {
		fs = af.NewMemMapFs()
		checkDeepEquals(t, crawlDirectory(""), []string{})
	})

	t.Run("crawl directory with two files", func(t *testing.T) {
		fs = af.NewMemMapFs()
		makeNewFile(fs, "test.json", "file1")
		makeNewFile(fs, "test2.json", "file2")
		checkDeepEquals(t, crawlDirectory(""), []string{"test", "test2"})
	})

	t.Run("crawl directory with non json file", func(t *testing.T) {
		fs = af.NewMemMapFs()
		makeNewFile(fs, "test.json", "file1")
		makeNewFile(fs, "asdf.txt", "asdf")
		checkDeepEquals(t, crawlDirectory(""), []string{"test"})
	})
}

func TestToMap(t *testing.T) {

	t.Run("simple flat json to map", func(t *testing.T) {
		fs = af.NewMemMapFs()

		expected := map[string]interface{}{
			"field":  "value",
			"field2": "value2",
		}

		f := makeNewJSON(fs, "test", expected)

		got, err := f.ToMap()
		assertNilErr(t, err)
		checkDeepEquals(t, expected, got)
	})

	t.Run("json with array", func(t *testing.T) {
		fs = af.NewMemMapFs()

		expected := map[string]interface{}{
			"array": []string{
				"a",
				"b",
			},
		}

		f := makeNewJSON(fs, "test", expected)

		got, err := f.ToMap()
		assertNilErr(t, err)
		checkJSONEquals(t, expected, got)
	})

	t.Run("deep nested with map", func(t *testing.T) {
		fs = af.NewMemMapFs()

		expected := map[string]interface{}{
			"array": []interface{}{
				"a",
				map[string]interface{}{
					"test": "deep nest",
				},
			},
		}

		f := makeNewJSON(fs, "test", expected)

		got, err := f.ToMap()
		assertNilErr(t, err)
		checkJSONEquals(t, expected, got)
	})
}

func TestReplaceContent(t *testing.T) {

	t.Run("update existing file", func(t *testing.T) {
		fs = af.NewMemMapFs()

		old := map[string]interface{}{
			"field":  "value",
			"field2": "value2",
		}

		new := map[string]interface{}{
			"field": "value",
		}

		f := makeNewJSON(fs, "test", old)
		assertFileExists(t, fs, "test")

		err := f.ReplaceContent(mapToString(new))
		assertNilErr(t, err)

		got, err := f.ToMap()
		assertNilErr(t, err)
		checkJSONEquals(t, got, new)
	})

	t.Run("create new file", func(t *testing.T) {
		fs = af.NewMemMapFs()

		new := map[string]interface{}{
			"field": "value",
		}

		f := &File{FileName: "test"}
		assertFileDoesNotExist(t, fs, "test")

		err := f.ReplaceContent(mapToString(new))
		assertNilErr(t, err)
		assertFileExists(t, fs, "test")

		got, err := f.ToMap()
		assertNilErr(t, err)
		checkJSONEquals(t, got, new)
	})
}

func TestDelete(t *testing.T) {

	t.Run("delete non-existent file", func(t *testing.T) {
		fs = af.NewMemMapFs()
		f := &File{FileName: "doesnt-exist"}
		assertFileDoesNotExist(t, fs, "doesnt-exist")

		err := f.Delete()
		assertErr(t, err)
		assertFileDoesNotExist(t, fs, "doesnt-exist")
	})

	t.Run("delete existing file", func(t *testing.T) {
		fs = af.NewMemMapFs()

		data := map[string]interface{}{
			"field": "value",
		}

		f := makeNewJSON(fs, "test", data)
		assertFileExists(t, fs, "test")

		err := f.Delete()
		assertNilErr(t, err)
		assertFileDoesNotExist(t, fs, "test")
	})
}
