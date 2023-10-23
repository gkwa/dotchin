package dotchin

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"time"

	"log/slog"

	"github.com/taylormonacelli/dotchin/instanceinfo"
	mymazda "github.com/taylormonacelli/forestfish/mymazda"
)

func expireCache(maxAge time.Duration, filePath string) {
	if !mymazda.FileExists(cachePath) {
		return
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	age := time.Since(fileInfo.ModTime()).Truncate(time.Second)
	expires := time.Until(fileInfo.ModTime().Add(maxAge)).Truncate(time.Second)

	if age > maxAge {
		slog.Debug("cache is old", "age", age, "path", cachePath)
		defer os.Remove(cachePath)
	} else {
		slog.Debug("cache not old", "age", age, "expires in", expires, "path", cachePath)
	}
}

func loadInfoMap(regions []string, infoMap *instanceinfo.InstanceInfoMap) error {
	if mymazda.FileExists(cachePath) {
		slog.Info("cache", "hit", true)

		byteSlice, err := os.ReadFile(cachePath)
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

	slog.Info("cache", "hit", false)
	networkFetchInfoMap(regions, infoMap)

	err := saveInfoMap(infoMap)
	if err != nil {
		slog.Error("persistMapToDisk", "error", err)
		return err
	}

	return nil
}

func saveInfoMap(infoMap *instanceinfo.InstanceInfoMap) error {
	var buffer bytes.Buffer
	gob.Register(instanceinfo.InstanceInfoMap{})

	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(*infoMap)
	if err != nil {
		return err
	}

	file, err := os.Create(cachePath)
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
