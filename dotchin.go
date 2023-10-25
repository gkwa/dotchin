package dotchin

import (
	"log/slog"
	"time"

	"github.com/taylormonacelli/busybus"
	"github.com/taylormonacelli/dotchin/instanceinfo"
	"github.com/taylormonacelli/lemondrop"
	"github.com/taylormonacelli/somespider"
)

var (
	config    *busybus.CacheConfig
	cachePath string
)

func Main(noCache bool) int {
	var err error

	cachePath, err = somespider.GenPath("dotchin/data.gob")
	if err != nil {
		slog.Error("generating cache path failed", "path", cachePath, "error", err)
		return 1
	}

	cacheLifetime := 24 * time.Hour
	if noCache {
		cacheLifetime = 0 * time.Second
	}

	config, err = busybus.NewConfig(cachePath, cacheLifetime)
	if err != nil {
		slog.Error("generating cache config failed", "path", cachePath, "error", err)
		return 1
	}

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

	err = config.RemoveExpiredCache()
	if err != nil {
		slog.Error("ExpireCache", "error", err)
		return 1
	}

	infoMap := instanceinfo.NewInstanceInfoMap()

	cacheErr := busybus.DecodeFromCache(*config, &infoMap)
	if cacheErr == nil {
		slog.Info("cache", "hit", true)
	} else {
		slog.Info("cache", "hit", false)
		instanceinfo.NetworkFetchInfoMap(regionsChosen, infoMap)

		cacheErr = busybus.SaveToCache(*config, &infoMap)
		if cacheErr != nil {
			slog.Error("SaveToCache", "error", cacheErr)
			return 1
		}
	}

	slog.Debug("infoMap", "region count", len(infoMap.GetRegions()))

	for _, region := range infoMap.GetRegions() {
		types := infoMap.Get(region)
		slog.Debug("regions check", "region", region, "types", len(types))
	}

	return 0
}
