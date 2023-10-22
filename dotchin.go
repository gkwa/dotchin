package dotchin

import (
	"encoding/json"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/taylormonacelli/lemondrop"
)

func Main() int {
	slog.Debug("dotchin", "test", true)

	regionDetails, err := lemondrop.GetRegionDetails()
	if err != nil {
		slog.Error("get regions", "error", err)
	}

	regionNames := make([]string, 0)
	for _, region := range regionDetails {
		regionNames = append(regionNames, region.RegionCode)
	}

	// regions := chooseRandomItem(regionNames, 100)
	regions := regionNames
	slog.Debug("searching regions", "regions", regions)

	result := GetInstanceTypesAvailableInRegions(regions)

	indentedJSON, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		slog.Error("marshalling", "error", err)
		return 1
	}

	fmt.Println(string(indentedJSON))
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

func GetInstanceTypesAvailableInRegions(regions []string) map[string][]types.InstanceTypeInfo {
	concurrencyLimit := 100
	semaphoreChan := make(chan struct{}, concurrencyLimit)
	defer close(semaphoreChan)
	wg := sync.WaitGroup{}

	results := make(chan []types.InstanceTypeInfo, len(regions))
	x := make(map[string][]types.InstanceTypeInfo)

	for _, region := range regions {
		wg.Add(1)
		semaphoreChan <- struct{}{} // Acquire semaphore
		go func(region string) {
			defer func() {
				<-semaphoreChan // Release semaphore
				wg.Done()
			}()

			output, err := GetInstanceTypesAvailableInRegion(region)
			if err != nil {
				slog.Error("GetInstanceTypesAvailableInRegion", "error", err)
				return
			}
			x[region] = output

			slog.Debug("instance metrics", "region", region, "count", len(output))

			results <- output
		}(region)
	}

	wg.Wait()
	close(results)

	return x
}

func GetInstanceTypesAvailableInRegion(region string) ([]types.InstanceTypeInfo, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		fmt.Printf("Error loading AWS SDK configuration: %v\n", err)
		return []types.InstanceTypeInfo{}, err
	}

	client := ec2.NewFromConfig(cfg)

	input := &ec2.DescribeInstanceTypesInput{}

	var allInstanceTypes []types.InstanceTypeInfo

	paginator := ec2.NewDescribeInstanceTypesPaginator(client, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			slog.Error("error describing instance types", "error", err)
			return []types.InstanceTypeInfo{}, err
		}

		allInstanceTypes = append(allInstanceTypes, page.InstanceTypes...)
	}

	return allInstanceTypes, nil
}
