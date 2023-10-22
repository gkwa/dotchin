package dotchin

import (
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/taylormonacelli/dotchin/instanceinfo"
	"github.com/taylormonacelli/lemondrop"
)

func Main() int {
	slog.Debug("dotchin", "test", true)

	regionDetails, err := lemondrop.GetRegionDetails()
	if err != nil {
		slog.Error("get regions", "error", err)
		return 1
	}

	regionNames := make([]string, 0)
	for _, region := range regionDetails {
		regionNames = append(regionNames, region.RegionCode)
	}

	regions := chooseRandomItem(regionNames, 100)
	slog.Debug("searching regions", "regions", regions)

	x := instanceinfo.NewInstanceInfoMap()
	GetInstanceTypesAvailableInRegions(regions, *x)
	return 0
}

func chooseRandomItem(items []string, count int) []string {
	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))

	if count > len(items) {
		count = len(items)
	}

	rng.Shuffle(len(items), func(i, j int) {
		items[i], items[j] = items[j], items[i]
	})

	randomSlice := items[:count]

	return randomSlice
}

func GetInstanceTypesAvailableInRegions(regions []string, myMap instanceinfo.InstanceInfoMap) {
	concurrencyLimit := 100
	wg := sync.WaitGroup{}

	semaphoreChan := make(chan struct{}, concurrencyLimit)
	defer close(semaphoreChan)

	results := make(chan instanceinfo.InstanceTypeInfoSlice)

	for _, region := range regions {
		wg.Add(1)
		semaphoreChan <- struct{}{} // Acquire semaphore
		go func(region string) {
			defer func() {
				<-semaphoreChan // Release semaphore
				wg.Done()
			}()

			var instanceInfos instanceinfo.InstanceTypeInfoSlice
			err := GetInstanceTypesProvidedInRegion(region, &instanceInfos)
			if err != nil {
				slog.Error("GetInstanceTypesAvailableInRegion", "error", err)
				return
			}

			results <- instanceInfos
			myMap.Add(region, instanceInfos)
			slog.Debug("instance metrics", "region", region, "count", len(myMap.Get(region)))
		}(region)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	// block to complete
	for range results {
	}
}

func GetInstanceTypesProvidedInRegion(region string, allInstanceTypes *instanceinfo.InstanceTypeInfoSlice) error {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		slog.Error("error loading AWS SDK configuration", "region", region, "error", err)
		return err
	}

	client := ec2.NewFromConfig(cfg)

	input := &ec2.DescribeInstanceTypesInput{}

	paginator := ec2.NewDescribeInstanceTypesPaginator(client, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			slog.Error("error describing instance types", "error", err)
			return err
		}

		*allInstanceTypes = append(*allInstanceTypes, page.InstanceTypes...)
	}

	return nil
}
