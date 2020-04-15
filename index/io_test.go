package index

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	os "github.com/spf13/afero"
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

func makeNewFile(fs os.Fs, name string, contents string) {
	os.WriteFile(fs, name, []byte(contents), 0644)
}

func makeNewJSON(fs os.Fs, name string, contents map[string]interface{}) *File {
	jsonData, _ := json.Marshal(contents)
	os.WriteFile(fs, name+".json", jsonData, 0644)
	return &File{FileName: name}
}

func mapToString(contents map[string]interface{}) string {
	jsonData, _ := json.Marshal(contents)
	return string(jsonData)
}

func TestCrawlDirectory(t *testing.T) {

	t.Run("crawl empty directory", func(t *testing.T) {
		fs = os.NewMemMapFs()
		checkDeepEquals(t, crawlDirectory(""), []string{})
	})

	t.Run("crawl directory with two files", func(t *testing.T) {
		fs = os.NewMemMapFs()
		makeNewFile(fs, "test.json", "file1")
		makeNewFile(fs, "test2.json", "file2")
		checkDeepEquals(t, crawlDirectory(""), []string{"test", "test2"})
	})

	t.Run("crawl directory with non json file", func(t *testing.T) {
		fs = os.NewMemMapFs()
		makeNewFile(fs, "test.json", "file1")
		makeNewFile(fs, "asdf.txt", "asdf")
		checkDeepEquals(t, crawlDirectory(""), []string{"test"})
	})
}

func TestToMap(t *testing.T) {

	t.Run("simple flat json to map", func(t *testing.T) {
		fs = os.NewMemMapFs()

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
		fs = os.NewMemMapFs()

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
		fs = os.NewMemMapFs()

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

	t.Run("change all content", func(t *testing.T) {
		fs = os.NewMemMapFs()

		old := map[string]interface{}{
			"field":  "value",
			"field2": "value2",
		}

		new := map[string]interface{}{
			"field": "value",
		}

		f := makeNewJSON(fs, "test", old)

		err := f.ReplaceContent(mapToString(new))
		assertNilErr(t, err)

		got, err := f.ToMap()
		assertNilErr(t, err)
		checkDeepEquals(t, got, new)
	})
}

func TestDelete(t *testing.T) {

	t.Run("delete non-existent file", func(t *testing.T) {
		fs = os.NewMemMapFs()
		f := &File{FileName: "doesnt-exist"}
		err := f.Delete()
		assertErr(t, err)
	})
}
