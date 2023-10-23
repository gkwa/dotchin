package cache

import (
	"bytes"
	"encoding/gob"
	"os"
	"time"

	"log/slog"

	"github.com/taylormonacelli/dotchin/instanceinfo"
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

func LoadInfoMap(regions []string, infoMap *instanceinfo.InstanceInfoMap) error {
	byteSlice, err := os.ReadFile(CachePath)
	if err != nil {
		return err
	}
	var buffer bytes.Buffer
	buffer.Write(byteSlice)

	gob.Register(instanceinfo.InstanceInfoMap{})
	dec := gob.NewDecoder(&buffer)
	err = dec.Decode(&infoMap)
	if err != nil {
		return err
	}

	return nil
}

func SaveInfoMap(infoMap *instanceinfo.InstanceInfoMap) error {
	var buffer bytes.Buffer
	gob.Register(instanceinfo.InstanceInfoMap{})

	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(*infoMap)
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
