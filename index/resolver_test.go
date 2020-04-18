package index

import (
    af "github.com/spf13/afero"
    "github.com/stretchr/testify/assert"
    "reflect"
    "strings"
    "testing"
)

func TestResolveReferences(t *testing.T) {

    firstContentWithRef := map[string]interface{}{
        "test":      "testVal",
        "secondVal": "REF::second",
    }

    secondContentWithRef := map[string]interface{}{
        "just": "strings",
        "ref":  "REF::third",
    }

    baseContent := map[string]interface{}{
        "key": "value",
    }

    t.Run("string with no ref should be returned as is", func(t *testing.T) {
        I.SetFileSystem(af.NewMemMapFs())

        got := ResolveReferences("test", 1)
        want := "test"

        assert.Equal(t, got, want)
    })

    t.Run("datatypes other than string, slice, and map are returned as is", func(t *testing.T) {
        I.SetFileSystem(af.NewMemMapFs())

        got := ResolveReferences(2, 1)
        want := 2

        assert.Equal(t, got, want)
    })

    t.Run("string with ref should replace the ref correctly", func(t *testing.T) {
        I.SetFileSystem(af.NewMemMapFs())

        makeNewJSON("testjson", baseContent)
        I.Regenerate()

        got := ResolveReferences("REF::testjson", 1)

        assert.Equal(t, got, baseContent)
    })

    t.Run("string with non-existent ref should return error message", func(t *testing.T) {
        got := ResolveReferences("REF::nonexistent", 1)
        gotVal := reflect.ValueOf(got)

        if gotVal.Kind() != reflect.String {
            t.Errorf("the resolved value should have been a string but got type '%s'", gotVal.Kind())
        }

        assert.True(t, strings.Contains(gotVal.String(), "REF::ERR"))
    })

    t.Run("refs within a slice should all be replaced", func(t *testing.T) {
        I.SetFileSystem(af.NewMemMapFs())

        makeNewJSON("testjson1", baseContent)
        makeNewJSON("testjson2", baseContent)
        I.Regenerate()

        refSlice := []string{"test", "REF::testjson1", "notref", "REF::testjson2"}
        got := ResolveReferences(refSlice, 1)

        expectedSlice := []interface{}{"test", baseContent, "notref", baseContent}
        assert.Equal(t, got, expectedSlice)
    })

    t.Run("refs within map values should all be replaced", func(t *testing.T) {
        I.SetFileSystem(af.NewMemMapFs())

        makeNewJSON("testjson1", baseContent)
        makeNewJSON("testjson2", baseContent)
        I.Regenerate()

        refMap := map[string]interface{}{
            "firstRef":  "REF::testjson1",
            "nonRef":    "nothing here",
            "secondRef": "REF::testjson2",
        }
        got := ResolveReferences(refMap, 1)

        expectedMap := map[string]interface{}{
            "firstRef":  baseContent,
            "nonRef":    "nothing here",
            "secondRef": baseContent,
        }
        assert.Equal(t, got, expectedMap)
    })

    t.Run("double nested refs should be resolved when depth permits", func(t *testing.T) {
        I.SetFileSystem(af.NewMemMapFs())

        makeNewJSON("first", firstContentWithRef)
        makeNewJSON("second", secondContentWithRef)
        makeNewJSON("third", baseContent)
        I.Regenerate()

        got := ResolveReferences(firstContentWithRef, 2)

        expectedMap := map[string]interface{}{
            "test": "testVal",
            "secondVal": map[string]interface{}{
                "just": "strings",
                "ref":  baseContent,
            },
        }
        assert.Equal(t, got, expectedMap)
    })

    t.Run("double nested refs only resolve one because of depth param", func(t *testing.T) {
        I.SetFileSystem(af.NewMemMapFs())

        makeNewJSON("first", firstContentWithRef)
        makeNewJSON("second", secondContentWithRef)
        makeNewJSON("third", baseContent)
        I.Regenerate()

        got := ResolveReferences(firstContentWithRef, 1)

        expectedMap := map[string]interface{}{
            "test": "testVal",
            "secondVal": map[string]interface{}{
                "just": "strings",
                "ref":  "REF::third",
            },
        }
        assert.Equal(t, got, expectedMap)
    })

}