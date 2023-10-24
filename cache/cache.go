package cache

import (
	"bytes"
	"encoding/gob"
	"os"
	"time"

	"log/slog"

	mymazda "github.com/taylormonacelli/forestfish/mymazda"
)

var CachePath = "/tmp/data.gob"
var CacheLifetime = 24 * time.Hour

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

func LoadMyCachedObject(blob interface{}, infoMap interface{}) error {
	byteSlice, err := os.ReadFile(CachePath)
	if err != nil {
		return err
	}
	var buffer bytes.Buffer
	buffer.Write(byteSlice)

	gob.Register(blob)
	dec := gob.NewDecoder(&buffer)
	err = dec.Decode(&infoMap)
	if err != nil {
		return err
	}

	return nil
}

func SaveMyArbitrationOjbect(myInfoMap interface{}) error {
	var buffer bytes.Buffer
	gob.Register(myInfoMap)

	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(myInfoMap)
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
