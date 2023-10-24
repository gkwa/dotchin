package dotchin

import (
	"log/slog"

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
		cache.LoadMyCachedObject(infoMap, infoMap)
	} else {
		slog.Info("cache", "hit", false)
		instanceinfo.NetworkFetchInfoMap(regionsChosen, infoMap)
	}
	slog.Debug("infoMap", "region count", len(infoMap.GetRegions()))

	for _, region := range infoMap.GetRegions() {
		types := infoMap.Get(region)
		slog.Debug("regions check", "region", region, "types", len(types))
	}

	return 0
}
