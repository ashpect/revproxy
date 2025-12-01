package cache

import (
	"fmt"
	"time"
)

func (c *Cache) LRU(key interface{}) *Cache {

	// Abstracting cache hit and updation is not required and would make things complicated, since cache value updation is dependent on policy.
	// Hence provided freedom to update logic in the policy.
	if !c.IsCacheFull() || c.IsValueExists(key) {
		c.Set(key)
		return c
	}

	var oldestTime time.Time
	var delkey interface{}
	m := c.GetAll()
	for k, v := range m {
		parsedTime, ok := v.(time.Time)
		if !ok {
			panic(fmt.Errorf("value %v is not a time format: %v", k, v))
		}

		if parsedTime.Before(oldestTime) {
			oldestTime = parsedTime
			delkey = k
		}
	}

	c.Delete(delkey)
	c.Set(key)

	return c
}
