package instanceinfo

import (
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type InstanceTypeInfoSlice []types.InstanceTypeInfo

type InstanceInfoMap struct {
	Data map[string]InstanceTypeInfoSlice
}

func NewInstanceInfoMap() *InstanceInfoMap {
	return &InstanceInfoMap{
		Data: make(map[string]InstanceTypeInfoSlice),
	}
}

func (m *InstanceInfoMap) Add(key string, value []types.InstanceTypeInfo) {
	m.Data[key] = append(m.Data[key], value...)
}

func (m *InstanceInfoMap) Get(key string) []types.InstanceTypeInfo {
	return m.Data[key]
}

func (m *InstanceInfoMap) GetRegions() []string {
	regionStrings := make([]string, 0, len(m.Data))
	for regionName := range m.Data {
		regionStrings = append(regionStrings, regionName)
	}

	return regionStrings
}

