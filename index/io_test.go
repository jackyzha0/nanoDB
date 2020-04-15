package index

import (
	"reflect"
	"testing"

	os "github.com/spf13/afero"
)

func TestCrawlDirectory(t *testing.T) {

	checkSliceEquals := func(t *testing.T, a interface{}, b interface{}) {
		t.Helper()
		if !reflect.DeepEqual(a, b) {
			t.Errorf("got %+v, want %+v", a, b)
		}
	}

	t.Run("crawl empty directory", func(t *testing.T) {
		fs = os.NewMemMapFs()
		checkSliceEquals(t, crawlDirectory(""), []string{})
	})
}
