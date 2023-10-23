package dotchin

import (
	"bytes"
	"context"
	"encoding/gob"
	"io"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/taylormonacelli/dotchin/instanceinfo"
	mymazda "github.com/taylormonacelli/forestfish/mymazda"
	"github.com/taylormonacelli/lemondrop"
)

var cachePath = "/tmp/data.gob"

func Main() int {
	slog.Debug("dotchin", "test", true)

	regionDetails, err := lemondrop.GetRegionDetails()
	if err != nil {
		slog.Error("get regions", "error", err)
		return 1
	}

	regionNames := make([]string, 0, len(regionDetails))
	for _, region := range regionDetails {
		regionNames = append(regionNames, region.RegionCode)
	}

	regions := chooseRandomItem(regionNames, 100)
	slog.Debug("searching regions", "regions", regions)

	infoMap := instanceinfo.NewInstanceInfoMap()
	FillInfoMap(regions, infoMap)

	slog.Debug("regions in map", "count", len(infoMap.GetRegions()))
	return 0
}

func FillInfoMap(regions []string, infoMap *instanceinfo.InstanceInfoMap) error {
	if mymazda.FileExists(cachePath) {
		slog.Info("cache", "hit", true)
		buffer := loadFromFile()
		err := readMapFromBuffer(buffer, *infoMap)
		if err != nil {
			return err
		}

		return nil
	}

	slog.Info("cache", "hit", false)
	fetchInfoMapFromNetwork(regions, infoMap)

	err := persistMapToDisk(infoMap)
	if err != nil {
		slog.Error("persistMapToDisk", "error", err)
		return err
	}

	return nil
}

func fetchInfoMapFromNetwork(regions []string, infoMap *instanceinfo.InstanceInfoMap) error {
	concurrencyLimit := 5
	wg := sync.WaitGroup{}

	semaphoreChan := make(chan struct{}, concurrencyLimit)
	defer close(semaphoreChan)

	results := make(chan instanceinfo.InstanceTypeInfoSlice, len(regions))

	for _, region := range regions {
		wg.Add(1)
		semaphoreChan <- struct{}{} // Acquire semaphore
		go func(region string) {
			defer func() {
				<-semaphoreChan // Release semaphore
				wg.Done()
			}()

			var instanceTypes instanceinfo.InstanceTypeInfoSlice
			err := GetInstanceTypesProvidedInRegion(region, &instanceTypes)
			if err != nil {
				slog.Error("GetInstanceTypesAvailableInRegion", "region", region, "error", err)
			}

			slog.Debug("instance types", "region", region, "count", len(instanceTypes))
			results <- instanceTypes
			infoMap.Add(region, instanceTypes)
			slog.Debug("instance metrics", "region", region, "count", len(infoMap.Get(region)))
		}(region)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	// block to complete
	for range results {
	}

	return nil
}

func loadFromFile() bytes.Buffer {
	file, err := os.Open(cachePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var buffer bytes.Buffer

	_, err = io.Copy(&buffer, file)
	if err != nil {
		panic(err)
	}

	return buffer
}

func readMapFromBuffer(buffer bytes.Buffer, infoMap instanceinfo.InstanceInfoMap) error {
	dec := gob.NewDecoder(&buffer)
	gob.Register(instanceinfo.InstanceInfoMap{})

	err := dec.Decode(&infoMap)
	if err != nil {
		slog.Error("decode", "error", err)
		return err
	}

	return nil
}

func persistMapToDisk(infoMap *instanceinfo.InstanceInfoMap) error {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	gob.Register(instanceinfo.InstanceInfoMap{})

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

func GetInstanceTypesProvidedInRegion(region string, allInstanceTypes *instanceinfo.InstanceTypeInfoSlice) error {
	ctx1, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx1, config.WithRegion(region))
	if err != nil {
		slog.Error("error loading AWS SDK configuration", "region", region, "error", err)
		return err
	}

	client := ec2.NewFromConfig(cfg)

	input := &ec2.DescribeInstanceTypesInput{}

	paginator := ec2.NewDescribeInstanceTypesPaginator(client, input)

	pageCount := 1
	for paginator.HasMorePages() {
		slog.Debug("fetching page", "count", pageCount)
		ctx2, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		page, err := paginator.NextPage(ctx2)
		if err != nil {
			slog.Error("error describing instance types", "region", region, "error", err)
			return err
		}

		*allInstanceTypes = append(*allInstanceTypes, page.InstanceTypes...)
		pageCount += 1
	}

	return nil
}
