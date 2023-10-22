package dotchin

import (
	"log/slog"

	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func Main() int {
	slog.Debug("dotchin", "test", true)
	test()
	return 0
}

func test() {
	// regions,err := lemondrop.GetRegionDetails()
	// if err != nil {
	// 	slog.Error("fetching aws regions","error",err)
	// 	return
	// }

	region := "us-west-2"
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		fmt.Printf("Error loading AWS SDK configuration: %v\n", err)
		return
	}

	client := ec2.NewFromConfig(cfg)

	// Create an input for DescribeInstanceTypes operation.
	input := &ec2.DescribeInstanceTypesInput{}

	resp, err := client.DescribeInstanceTypes(context.TODO(), input)
	if err != nil {
		fmt.Printf("Error describing instance types: %v\n", err)
		return
	}

	// Iterate over the instance types and print their details.
	for _, instanceType := range resp.InstanceTypes {
		fmt.Printf("Instance Type: %s\n", instanceType.InstanceType)
		fmt.Println()
	}
}
