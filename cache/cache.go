package cache

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log/slog"
	"os"
	"time"

	mymazda "github.com/taylormonacelli/forestfish/mymazda"
)

var (
	CachePath     = "/tmp/data.gob"
	CacheLifetime = 24 * time.Hour
)

func DecodeFromCache(target interface{}) error {
	if !mymazda.FileExists(CachePath) {
		return fmt.Errorf("cache file does not exist")
	}

	byteSlice, err := os.ReadFile(CachePath)
	if err != nil {
		return err
	}

	var buffer bytes.Buffer
	buffer.Write(byteSlice)

	dec := gob.NewDecoder(&buffer)
	err = dec.Decode(target)

	if err != nil {
		return err
	}

	return nil
}

func SaveToCache(data interface{}) error {
	var buffer bytes.Buffer
	gob.Register(data)

	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(data)
	if err != nil {
		return err
	}

	file, err := os.Create(CachePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(buffer.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func ExpireCache(maxAge time.Duration, filePath string) error {
	if !mymazda.FileExists(CachePath) {
		return nil
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	age := time.Since(fileInfo.ModTime()).Truncate(time.Second)
	expires := time.Until(fileInfo.ModTime().Add(maxAge)).Truncate(time.Second)

	if age > maxAge {
		slog.Debug("cache is old", "age", age, "path", CachePath)
		defer os.Remove(CachePath)
	} else {
		slog.Debug("cache not old", "age", age, "expires in", expires, "path", CachePath)
	}

	return nil
}
