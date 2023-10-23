package dotchin

import (
	"math/rand"
	"time"
)

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
