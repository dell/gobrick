//go:build test

/*
 *
 * Copyright © 2020-2025 Dell Inc. or its subsidiaries. All Rights Reserved.
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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_rateLock(t *testing.T) {
	r := newRateLock()
	testKey := "foo"
	assert.True(t, r.RateCheck(testKey, 1))
	assert.False(t, r.RateCheck(testKey, 1))
	time.Sleep(time.Second * 2)
	assert.True(t, r.RateCheck(testKey, 1))
	r.Stop()
}

func TestTtlCache(t *testing.T) {
	testKey := "foo"
	testKey2 := "foo2"
	testValue := "bar"

	t.Run("get cached value", func(t *testing.T) {
		cache := newTTLCache(time.Second)
		cache.Set(testKey, testValue)
		value, found := cache.Get(testKey)
		assert.True(t, found)
		assert.Equal(t, testValue, value.(string))
		cache.Stop()
	})
	t.Run("get unknown value", func(t *testing.T) {
		cache := newTTLCache(time.Second)
		testKey := "foo"
		value, found := cache.Get(testKey)
		assert.False(t, found)
		assert.Nil(t, value)
		cache.Stop()
	})
	t.Run("test key expiration", func(t *testing.T) {
		cache := newTTLCache(time.Second)
		cache.Set(testKey, testValue)
		cache.SetWithTTL(testKey2, testValue, time.Second*3)
		value, found := cache.Get(testKey)
		assert.True(t, found)
		assert.Equal(t, testValue, value.(string))
		time.Sleep(time.Second * 2)
		value, found = cache.Get(testKey)
		assert.False(t, found)
		assert.Nil(t, value)
		value, found = cache.Get(testKey2)
		assert.True(t, found)
		assert.Equal(t, testValue, value.(string))
		time.Sleep(time.Second * 2)
		value, found = cache.Get(testKey2)
		assert.False(t, found)
		assert.Nil(t, value)
		cache.Stop()
	})
	t.Run("test concurrent access", func(t *testing.T) {
		cache := newTTLCache(time.Second)
		wg := sync.WaitGroup{}
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				cache.Get(testKey)
				cache.Set(testKey, testValue)
				cache.Get(testKey)
				cache.Set(testKey, testValue)
				wg.Done()
			}()
		}
		wg.Wait()
		value, found := cache.Get(testKey)
		assert.True(t, found)
		assert.Equal(t, testValue, value.(string))
	})
}

func Test_rateLock_Stop_coverage(t *testing.T) {
	// Test to cover the rateLock.Stop() method for 100% coverage
	r := newRateLock()
	r.Stop() // This covers the rateLock.Stop() method
}

func TestTtlCache_Set_coverage(t *testing.T) {
	// Test to cover the ttlCache.Set() method directly for 100% coverage
	cache := newTTLCache(time.Second)
	defer cache.Stop()

	testKey := "direct_set_key"
	testValue := "direct_set_value"

	// Direct call to Set method (not SetWithTTL)
	cache.Set(testKey, testValue)

	value, found := cache.Get(testKey)
	assert.True(t, found)
	assert.Equal(t, testValue, value.(string))
}

func TestTtlCache_Stop_coverage(t *testing.T) {
	// Test to cover the ttlCache.Stop() method for 100% coverage
	cache := newTTLCache(time.Second)
	cache.Stop() // This covers the ttlCache.Stop() method
}

func Test_ttlCache_purge_coverage(t *testing.T) {
	// Test to improve coverage by calling purge directly
	cache := newTTLCache(time.Millisecond * 10) // Very short TTL
	defer cache.Stop()

	// Add a key that will expire quickly
	cache.Set("short_lived", "value")

	// Wait for it to expire
	time.Sleep(time.Millisecond * 20)

	// Force purge to clean expired entries
	cache.purge()

	// Verify the expired key is gone
	value, found := cache.Get("short_lived")
	assert.False(t, found)
	assert.Nil(t, value)
}
