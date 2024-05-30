package cache

import (
	"fmt"
)

func getFullCacheKey(cachePrefix string, key string) string {
	if cachePrefix == "" {
		return key
	}
	return fmt.Sprintf("%s-%s", cachePrefix, key)
}
