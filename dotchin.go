package dotchin

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	cache "github.com/taylormonacelli/dotchin/cache"
	"github.com/taylormonacelli/dotchin/instanceinfo"
	mymazda "github.com/taylormonacelli/forestfish/mymazda"
	"github.com/taylormonacelli/lemondrop"
)

func Main() int {
	regionDetails, err := lemondrop.GetRegionDetails()
	if err != nil {
		slog.Error("get regions", "error", err)
		return 1
	}

	regions := make([]string, 0, len(regionDetails))
	for _, region := range regionDetails {
		regions = append(regions, region.RegionCode)
	}

	maxRegions := len(regions) //eg. for debug/test limit to 1 region
	regionsChosen := _filterRandomRegions(regions, maxRegions)
	slog.Debug("searching regions", "regions", regions)

	err = cache.ExpireCache(cache.CacheLifetime, cache.CachePath)
	if err != nil {
		slog.Error("ExpireCache", "error", err)
	}

	infoMap := instanceinfo.NewInstanceInfoMap()
	if mymazda.FileExists(cache.CachePath) {
		slog.Info("cache", "hit", true)
		cache.LoadInfoMap(regionsChosen, infoMap)
	} else {
		slog.Info("cache", "hit", false)
		networkFetchInfoMap(regions, infoMap)
	}
	slog.Debug("infoMap", "region count", len(infoMap.GetRegions()))

	for _, region := range infoMap.GetRegions() {
		types := infoMap.Get(region)
		slog.Debug("regions check", "region", region, "types", len(types))
	}

	return 0
}

func networkFetchInfoMap(regions []string, infoMap *instanceinfo.InstanceInfoMap) error {
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
			err := _getInstanceTypesProvidedInRegion(region, &instanceTypes)
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

	err := cache.SaveInfoMap(infoMap)
	if err != nil {
		slog.Error("persistMapToDisk", "error", err)
		return err
	}

	return nil
}

func _getInstanceTypesProvidedInRegion(region string, allInstanceTypes *instanceinfo.InstanceTypeInfoSlice) error {
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
