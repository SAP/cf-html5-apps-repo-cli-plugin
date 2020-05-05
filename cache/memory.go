package cache

var cacheMap = make(map[string]interface{})

// Get get value from cache
func Get(key string) (interface{}, bool) {
	if value, ok := cacheMap[key]; ok {
		return value, ok
	}
	return nil, false
}

// Set upsert value to cache
func Set(key string, value interface{}) {
	cacheMap[key] = value
}
