package instanceinfo

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func NetworkFetchInfoMap(regions []string, infoMap *InstanceInfoMap) error {
	concurrencyLimit := 12
	wg := sync.WaitGroup{}

	semaphoreChan := make(chan struct{}, concurrencyLimit)
	defer close(semaphoreChan)

	results := make(chan InstanceTypeInfoSlice, len(regions))

	for _, region := range regions {
		wg.Add(1)
		semaphoreChan <- struct{}{} // Acquire semaphore
		go func(region string) {
			defer func() {
				<-semaphoreChan // Release semaphore
				wg.Done()
			}()

			var instanceTypes InstanceTypeInfoSlice
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

	for range results {
	}

	return nil
}

func _getInstanceTypesProvidedInRegion(region string, allInstanceTypes *InstanceTypeInfoSlice) error {
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
