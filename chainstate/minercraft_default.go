package chainstate

import "github.com/tonicpow/go-minercraft/v2"

func defaultMinecraftConfig() *minercraftConfig {
	miners, _ := minercraft.DefaultMiners()
	apis, _ := minercraft.DefaultMinersAPIs()

	broadcastMiners := []*minercraft.Miner{}
	queryMiners := []*minercraft.Miner{}
	for _, miner := range miners {
		currentMiner := *miner
		broadcastMiners = append(broadcastMiners, &currentMiner)

		if supportsQuerying(&currentMiner) {
			queryMiners = append(queryMiners, &currentMiner)
		}
	}

	return &minercraftConfig{
		broadcastMiners: broadcastMiners,
		queryMiners:     queryMiners,
		minerAPIs:       apis,
	}
}

func supportsQuerying(mm *minercraft.Miner) bool {
	return mm.Name == minercraft.MinerTaal || mm.Name == minercraft.MinerMempool
}
