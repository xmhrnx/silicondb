package model

// derived returns the value cached under key, computing it with fn the
// first time. A catalog never changes once built, so anything derived
// from it (counts, groupings, rankings) is worked out once rather than on
// every frame that shows it.
func derived[T any](c *Catalog, key string, fn func() T) T {
	if v, ok := c.cache.Load(key); ok {
		return v.(T)
	}
	v, _ := c.cache.LoadOrStore(key, fn())
	return v.(T)
}
