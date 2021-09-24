package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMessageCache_IsRepeatedMessage(t *testing.T) {
	cache := New()
	key := "messageID"
	t.Run("Should validate if not is repited", func(t *testing.T) {
		defer cache.Clear()
		messageCache := NewMessageCache(cache)

		ok := messageCache.IsRepeatedMessage(key)
		assert.False(t, ok)
		cache.Wait()
		value, ok := cache.Get(assembleMessageKey(key))
		assert.True(t, ok)
		assert.Equal(t, struct{}{}, value)
	})

	t.Run("Should validate if is repited", func(t *testing.T) {
		defer cache.Clear()
		messageCache := NewMessageCache(cache)

		cache.Set(assembleMessageKey(key), struct{}{}, 2*time.Second)
		cache.Wait()

		ok := messageCache.IsRepeatedMessage(key)
		assert.True(t, ok)
	})

}
