package main

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func TestCASPathTransformFunc(t *testing.T) {
	key := "SomeKey"
	pathData := CASPathTransformFunc(key)

	expectedPath := "bf5cd/77691/581e6/a59f8/e3306/703cf/eccb6/e3849"
	expectedFileName := "bf5cd77691581e6a59f8e3306703cfeccb6e3849"

	if pathData.FilePath != expectedPath {
		t.Fatalf("Expected path: %s, got: %s", expectedPath, pathData.FilePath)
	}

	if pathData.FileName != expectedFileName {
		t.Fatalf("Expected file name: %s, got: %s", expectedFileName, pathData.FileName)
	}
}

func TestStore(t *testing.T) {
	store := newStore()
	defer tearDown(t, store)

	for i := 0; i < 50; i++ {

		key := fmt.Sprintf("%s-%d", "SomeKey", i)
		data := []byte(fmt.Sprintf("%s-%d", "SomeData", i))

		if err := store.Write(key, bytes.NewReader(data)); err != nil {
			t.Fatalf("Error writing stream: %v", err)
		}

		if ok := store.Has(key); !ok {
			t.Fatalf("Expected key to be present")
		}

		r, err := store.Read(key)

		if err != nil {
			t.Fatalf("Error reading stream: %v", err)
		}

		b, _ := io.ReadAll(r)

		fmt.Printf("Data: %s\n", string(b))

		if !bytes.Equal(b, data) {
			t.Fatalf("Expected data: %s, got: %s", string(data), string(b))
		}

		if err := store.Delete(key); err != nil {
			t.Fatalf("Error deleting stream: %v", err)
		}

		if store.Has(key) {
			t.Fatalf("Expected key to be deleted")
		}
	}
}

func newStore() *Store {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
		Root:              "testStore",
	}
	store := NewStore(opts)
	store.Clear()
	return store
}

func tearDown(t *testing.T, store *Store) {
	if err := store.Clear(); err != nil {
		t.Fatalf("Error clearing store: %v", err)
	}
}
