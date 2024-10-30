package utils

import (
	"fmt"
	"strings"
	"sync"
)

var requestCache sync.Map

// Helper function to create a cache key from headers
func CreateCacheKey(headers map[string]string) string {
	var keys []string
	for key, value := range headers {
		keys = append(keys, fmt.Sprintf("%s=%s", key, value))
	}
	return strings.Join(keys, "&")
}

func GetValue(headers map[string]string) (string, bool) {
	cacheKey := CreateCacheKey(headers)
	cachedResponse, found := requestCache.Load(cacheKey)
	if found {
		return cachedResponse.(string), found
	}
	return "", found
}

func StoreValue(headers map[string]string, response string) {
	key := CreateCacheKey(headers)
	requestCache.Store(key, response)
}
