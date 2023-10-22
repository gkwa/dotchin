package instanceinfo

import (
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type InstanceTypeInfoSlice []types.InstanceTypeInfo

type InstanceInfoMap struct {
	data map[string]InstanceTypeInfoSlice
}

func NewInstanceInfoMap() *InstanceInfoMap {
	return &InstanceInfoMap{
		data: make(map[string]InstanceTypeInfoSlice),
	}
}

func (m *InstanceInfoMap) Add(key string, value []types.InstanceTypeInfo) {
	m.data[key] = append(m.data[key], value...)
}

func (m *InstanceInfoMap) Get(key string) []types.InstanceTypeInfo {
	return m.data[key]
}

func (m *InstanceInfoMap) GetRegions() []string {
	regionStrings := make([]string, 0, len(m.data))
	for regionName := range m.data {
		regionStrings = append(regionStrings, regionName)
	}

	return regionStrings
}
