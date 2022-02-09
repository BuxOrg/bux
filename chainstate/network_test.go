package chainstate

import (
	"testing"

	"github.com/mrz1836/go-mattercloud"
	"github.com/mrz1836/go-whatsonchain"
	"github.com/stretchr/testify/assert"
)

// TestNetwork_String will test the method String()
func TestNetwork_String(t *testing.T) {
	t.Parallel()

	t.Run("test all networks", func(t *testing.T) {
		assert.Equal(t, mainNet, MainNet.String())
		assert.Equal(t, stn, StressTestNet.String())
		assert.Equal(t, testNet, TestNet.String())
	})

	t.Run("unknown network", func(t *testing.T) {
		un := Network("")
		assert.Equal(t, "", un.String())
	})
}

// TestNetwork_WhatsOnChain will test the method WhatsOnChain()
func TestNetwork_WhatsOnChain(t *testing.T) {
	t.Parallel()

	t.Run("test all networks", func(t *testing.T) {
		assert.Equal(t, whatsonchain.NetworkMain, MainNet.WhatsOnChain())
		assert.Equal(t, whatsonchain.NetworkStn, StressTestNet.WhatsOnChain())
		assert.Equal(t, whatsonchain.NetworkTest, TestNet.WhatsOnChain())
	})

	t.Run("unknown network", func(t *testing.T) {
		un := Network("")
		assert.Equal(t, whatsonchain.NetworkMain, un.WhatsOnChain())
	})
}

// TestNetwork_MatterCloud will test the method MatterCloud()
func TestNetwork_MatterCloud(t *testing.T) {
	t.Parallel()

	t.Run("test all networks", func(t *testing.T) {
		assert.Equal(t, mattercloud.NetworkMain, MainNet.MatterCloud())
		assert.Equal(t, mattercloud.NetworkStn, StressTestNet.MatterCloud())
		assert.Equal(t, mattercloud.NetworkTest, TestNet.MatterCloud())
	})

	t.Run("unknown network", func(t *testing.T) {
		un := Network("")
		assert.Equal(t, mattercloud.NetworkMain, un.MatterCloud())
	})
}

// TestNetwork_Alternate will test the method Alternate()
func TestNetwork_Alternate(t *testing.T) {
	t.Parallel()

	t.Run("test all networks", func(t *testing.T) {
		assert.Equal(t, mainNetAlt, MainNet.Alternate())
		assert.Equal(t, stn, StressTestNet.Alternate())
		assert.Equal(t, testNetAlt, TestNet.Alternate())
	})

	t.Run("unknown network", func(t *testing.T) {
		un := Network("")
		assert.Equal(t, "", un.Alternate())
	})
}
