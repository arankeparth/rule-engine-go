package utils

import (
	"fmt"
	"log"
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
	newKey := strings.Join(keys, "&")
	log.Printf("key: %s", newKey)
	return newKey
}

func GetValue(headers map[string]string) ([]string, bool) {
	cacheKey := CreateCacheKey(headers)
	cachedResponse, found := requestCache.Load(cacheKey)
	if found {
		log.Printf("############## %v", found)
		return cachedResponse.([]string), found
	}
	return nil, found
}

func StoreValue(headers map[string]string, response []string) {
	key := CreateCacheKey(headers)
	requestCache.Store(key, response)
}
