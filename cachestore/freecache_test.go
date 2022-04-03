package cachestore

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_loadFreeCache(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		c := loadFreeCache(0, 0)
		require.NotNil(t, c)
	})

	t.Run("custom values", func(t *testing.T) {
		c := loadFreeCache(DefaultCacheSize+1024, 15)
		require.NotNil(t, c)
	})
}
