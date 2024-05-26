package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"
	"os"
	"strings"
)

// Global variable to store the root name of the storage
var StoreRootName = "store"

type FileData struct {
	FilePath string
	FileName string
}

// FullPath returns the full path of the file, with the root store folder
func (f *FileData) FullPath() string {
	return StoreRootName + "/" + f.FilePath + "/" + f.FileName
}

// FileRootPath returns the root path of the file
func (f *FileData) FileRootPath() string {
	index := strings.Index(f.FilePath, "/")

	if index == -1 {
		log.Fatalf("Invalid file path: %s", f.FilePath)
	}

	return f.FilePath[:index]
}

// Default CASPathTransformFunc is the default function to transport the key to a path
func CASPathTransformFunc(key string) FileData {
	hash := sha1.Sum([]byte(key))

	hashStr := hex.EncodeToString(hash[:])

	blockSize := 5
	blockLen := len(hashStr) / blockSize

	paths := make([]string, blockLen)

	for i := 0; i < blockLen; i++ {
		from, to := i*blockSize, (i+1)*blockSize
		paths[i] = hashStr[from:to]
	}

	filePath := strings.Join(paths, "/")
	fileName := hashStr

	return FileData{
		FilePath: filePath,
		FileName: fileName,
	}
}

type PathTransformFunc func(string) FileData

func DefaultPathTransformFunc(key string) FileData {
	return FileData{
		FilePath: key,
		FileName: key,
	}
}

// StoreOpts is the struct that represents the options for the storage
type StoreOpts struct {
	// Root is the root directory where the data is stored
	Root              string
	PathTransformFunc PathTransformFunc
}

// Store is the struct that represents the storage of the data
type Store struct {
	StoreOpts
}

func NewStore(opts StoreOpts) *Store {
	if opts.PathTransformFunc == nil {
		opts.PathTransformFunc = DefaultPathTransformFunc
	}

	if len(opts.Root) == 0 {
		opts.Root = StoreRootName
	} else {
		StoreRootName = opts.Root
	}

	return &Store{
		StoreOpts: opts,
	}
}

// Clear clears the storage
func (s *Store) Clear() error {
	return os.RemoveAll(s.Root)
}

// Has checks if the key exists in the storage
func (s *Store) Has(key string) bool {
	fileData := s.PathTransformFunc(key)

	_, err := os.Stat(fileData.FullPath())

	return !os.IsNotExist(err)
}

// Delete deletes the key from the storage
func (s *Store) Delete(key string) error {
	fileData := s.PathTransformFunc(key)

	if _, err := os.Stat(fileData.FullPath()); os.IsNotExist(err) {
		return err
	}

	defer log.Println("Deleted from the disk", fileData.FullPath())

	return os.RemoveAll(s.Root + "/" + fileData.FileRootPath())
}

// Write writes the data to the storage
func (s *Store) Write(key string, r io.Reader) error {
	return s.writeStream(key, r)
}

// Read reads the data from the storage
func (s *Store) Read(key string) (io.Reader, error) {
	f, err := s.readStream(key)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	buf := new(bytes.Buffer)

	_, err = io.Copy(buf, f)

	log.Println("Read from the disk", key)

	return buf, err
}

// Undlerlying functions to read data from the storage and returns a reader
func (s *Store) readStream(key string) (io.ReadCloser, error) {
	fileData := s.PathTransformFunc(key)

	if _, err := os.Stat(fileData.FullPath()); os.IsNotExist(err) {
		return nil, err
	}

	return os.Open(fileData.FullPath())
}

// Undlerlying functions to write data to the storage
func (s *Store) writeStream(key string, r io.Reader) error {
	fileData := s.PathTransformFunc(key)

	if err := os.MkdirAll(s.Root+"/"+fileData.FilePath, os.ModePerm); err != nil {
		return err
	}

	pathAndFileName := fileData.FullPath()

	f, err := os.Create(pathAndFileName)

	if err != nil {
		return err
	}

	defer f.Close()

	n, err := io.Copy(f, r)

	if err != nil {
		return err
	}

	log.Printf("Wrote %d bytes to %s", n, pathAndFileName)

	return nil
}
