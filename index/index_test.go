package index

import (
    "encoding/json"
    "testing"
)

func getFile(t *testing.T, key string) *File {
    t.Helper()

    I.RegenerateNew("db")
    file, ok := I.Lookup(key)
    if !ok {
        t.Errorf("file '%s' was not found", key)
    }

    return file
}

func checkKeyNotInIndex(t *testing.T, key string) {
    t.Helper()

    if _, ok := I.index[key]; ok {
        t.Errorf("should not have found key: '%s'", key)
    }
}

func TestFile_ResolvePath(t *testing.T) {
    t.Run("file path correct with directories", func(t *testing.T) {
        setup()
        makeNewFile("db/resolve_test.json", "test")

        file := getFile(t, "resolve_test")

        got := file.ResolvePath()
        want := "db/resolve_test.json"

        checkDeepEquals(t, got, want)
    })
}

func TestFileIndex_Lookup(t *testing.T) {
    t.Run("lookup existing file", func(t *testing.T) {
        setup()
        makeNewFile("db/lookup1.json", "test")

        file := getFile(t, "lookup1")

        bytes, _ := file.getByteArray()
        checkDeepEquals(t, string(bytes), "test")
    })

    t.Run("lookup non-existent file", func(t *testing.T) {
        setup()

        file, ok := I.Lookup("doesnt_exist")
        if ok {
            t.Errorf("should not have found file: '%s'", file.FileName)
        }
    })
}

func TestFileIndex_Delete(t *testing.T) {
    t.Run("delete file that exists", func(t *testing.T) {
        setup()
        makeNewFile("db/delete_test1.json", "test")

        key := "delete_test1"
        file := getFile(t, key)
        err := I.Delete(file)
        assertNilErr(t, err)

        checkKeyNotInIndex(t, key)
    })

    t.Run("delete file that does not exist", func(t *testing.T) {
        setup()

        key := "doesnt_exist"
        file := &File{FileName: "doesnt-exist"}
        assertFileDoesNotExist(t, "doesnt-exist")

        err := I.Delete(file)
        assertErr(t, err)

        checkKeyNotInIndex(t, key)
    })
}

func TestFileIndex_List(t *testing.T) {
    t.Run("list empty dir", func(t *testing.T) {
        setup()

        list := I.List()
        checkDeepEquals(t, len(list), 0)
    })

    t.Run("list dir with two files", func(t *testing.T) {
        setup()

        makeNewFile("db/list1.json", "test")
        makeNewFile("db/list2.json", "test")
        I.RegenerateNew("db")

        checkDeepEquals(t, I.List(), []string{"list1", "list2"})
    })
}

func TestFileIndex_Regenerate(t *testing.T) {
    t.Run("test if new files are added to index", func(t *testing.T) {
        setup()

        makeNewFile("regenerate1.json", "test")
        makeNewFile("regenerate2.json", "test")

        // index should be empty before regenerating
        checkDeepEquals(t, len(I.List()), 0)

        I.Regenerate()

        checkDeepEquals(t, I.List(), []string{"regenerate1", "regenerate2"})
    })

    t.Run("test RegenerateNew moves to given dir", func(t *testing.T) {
        setup()

        // in . not db
        makeNewFile("regenerate_new.json", "test")

        // in db
        makeNewFile("db/regenerate_new_db.json", "test")

        checkDeepEquals(t, len(I.List()), 0)

        I.RegenerateNew("db")

        checkDeepEquals(t, I.List(), []string{"regenerate_new_db"})
        checkDeepEquals(t, I.dir, "db")
    })
}

func TestFileIndex_Put(t *testing.T) {
    content := map[string]interface{}{
        "array": []interface{}{
            "a",
            map[string]interface{}{
                "test": "deep nest",
            },
        },
    }

    t.Run("create empty file with given content", func(t *testing.T) {
        setup()

        key := "put_empty"
        file := &File{FileName: key}
        assertFileDoesNotExist(t, key)

        bytes, _ := json.Marshal(content)
        err := I.Put(file, bytes)
        assertNilErr(t, err)
        assertFileExists(t, key)

        contentBytes, err := I.index[key].getByteArray()
        assertNilErr(t, err)
        checkJSONEquals(t, string(contentBytes), mapToString(content))
    })

    t.Run("replace existing file with given content", func(t *testing.T) {
        setup()

        newContent := map[string]interface{}{
            "field": "value",
        }

        key := "put_existing"
        file := makeNewJSON(key, content)
        assertFileExists(t, key)

        bytes, _ := json.Marshal(newContent)
        err := I.Put(file, bytes)
        assertNilErr(t, err)
        assertFileExists(t, key)

        contentBytes, err := I.index[key].getByteArray()
        assertNilErr(t, err)
        checkJSONEquals(t, string(contentBytes), mapToString(newContent))
    })
}