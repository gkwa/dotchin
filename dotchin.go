package dotchin

import (
	"log/slog"

	"github.com/taylormonacelli/busybus"
	"github.com/taylormonacelli/dotchin/instanceinfo"
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

	maxRegions := len(regions) // eg. for debug/test limit to 1 region
	regionsChosen := _filterRandomRegions(regions, maxRegions)
	slog.Debug("searching regions", "regions", regions)

	err = busybus.ExpireCache(busybus.CacheLifetime, busybus.CachePath)
	if err != nil {
		slog.Error("ExpireCache", "error", err)
	}

	infoMap := instanceinfo.NewInstanceInfoMap()
	cacheErr := busybus.DecodeFromCache(&infoMap)

	if cacheErr == nil {
		slog.Info("cache", "hit", true)
	} else {
		slog.Info("cache", "hit", false)
		instanceinfo.NetworkFetchInfoMap(regionsChosen, infoMap)

		cacheErr = busybus.SaveToCache(&infoMap)
		if cacheErr != nil {
			slog.Error("SaveToCache", "error", cacheErr)
		}
	}
	slog.Debug("infoMap", "region count", len(infoMap.GetRegions()))

	for _, region := range infoMap.GetRegions() {
		types := infoMap.Get(region)
		slog.Debug("regions check", "region", region, "types", len(types))
	}

	return 0
}
