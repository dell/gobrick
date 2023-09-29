/*
 *
 * Copyright © 2020 Dell Inc. or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package gobrick

import (
	"sync"
	"time"
)

type ttlCacheEntry struct {
	data       interface{}
	expireTime time.Time
}

func newTTLCache(ttl time.Duration) *ttlCache {
	c := &ttlCache{ttl: ttl}
	c.entries = make(map[string]ttlCacheEntry)
	c.ticker = time.NewTicker(time.Second)
	go c.tick()
	return c
}

// ttl cache
type ttlCache struct {
	ttl     time.Duration
	mutex   sync.RWMutex
	entries map[string]ttlCacheEntry
	ticker  *time.Ticker
}

func (c *ttlCache) tick() {
	for range c.ticker.C {
		c.purge()
	}
}

func (c *ttlCache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if value, ok := c.entries[key]; ok {
		return value.data, true
	}
	return nil, false
}

func (c *ttlCache) Set(key string, value interface{}) {
	c.set(key, value, c.ttl)
}

func (c *ttlCache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.set(key, value, ttl)
}

func (c *ttlCache) set(key string, value interface{}, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.entries[key] = ttlCacheEntry{data: value, expireTime: time.Now().Add(ttl)}
}

// remove entries with expired ttl
func (c *ttlCache) purge() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for key, value := range c.entries {
		if value.expireTime.Before(time.Now()) {
			delete(c.entries, key)
		}
	}
}

func (c *ttlCache) Stop() {
	c.ticker.Stop()
}

func newRateLock() *rateLock {
	m := &rateLock{}
	m.ttlCache = newTTLCache(time.Second)
	return m
}

// ttl based named m
type rateLock struct {
	ttlCache *ttlCache
	m        sync.Mutex
}

// RateCheck checks if key exist in internal ttl cache
// If key not exist will create it with required ttl.
// Useful for ttl-based rate limiting
func (r *rateLock) RateCheck(key string, ttl time.Duration) bool {
	r.m.Lock()
	defer r.m.Unlock()
	_, found := r.ttlCache.Get(key)
	if found {
		return false
	}
	r.ttlCache.SetWithTTL(key, "", ttl)
	return true
}

func (r *rateLock) Stop() {
	r.ttlCache.Stop()
}
