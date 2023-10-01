package chainstate

import (
	"context"
	"fmt"
	"sync"

	"github.com/BuxOrg/bux/utils"
)

type merkleRootProvider interface {
	getName() string
	verifyMerkleRoots(ctx context.Context, merkleProofs []string) error
}

// result of single merkle root verification by provider
type verifyResult struct {
	isError  bool
	err      error
	provider string
}

// VerifyMerkleRoots will try to verify merkle roots with all available providers
func (c *Client) VerifyMerkleRoots(ctx context.Context, merkleRoots []string) error {
	var wg sync.WaitGroup
	resultsChannel := make(chan verifyResult)

	for _, merkleRootPvdr := range createMerkleRootsProviders(c) {
		wg.Add(1)
		go func(provider merkleRootProvider) {
			defer wg.Done()
			verifyMerkleRootWithProvider(ctx, provider,
				resultsChannel, merkleRoots)
		}(merkleRootPvdr)
	}

	go func() {
		wg.Wait()
		close(resultsChannel)
	}()

	result := <-resultsChannel

	if result.isError {
		debugLog(c, "", fmt.Sprintf("verify merkle root error: %s from provider %s", result.err, result.provider))
		return result.err
	}
	debugLog(c, "", fmt.Sprintf("successful verification of merkle proofs by %s", result.provider))
	return nil
}

func createMerkleRootsProviders(c *Client) []merkleRootProvider {
	providers := make([]merkleRootProvider, 0, 10)

	if shouldVerifyWithPulse(c) {
		pvdr := pulseClientProvider{url: c.PulseClient().url, authToken: c.PulseClient().authToken}
		providers = append(providers, pvdr)
	}

	return providers
}

func shouldVerifyWithPulse(c *Client) bool {
	return !utils.StringInSlice(ProviderPulse, c.options.config.excludedProviders) &&
		c.PulseClient() != nil
}

func verifyMerkleRootWithProvider(ctx context.Context, provider merkleRootProvider, resultsChannel chan verifyResult, merkleRoots []string) {
	vErr := provider.verifyMerkleRoots(ctx, merkleRoots)

	if vErr != nil {
		resultsChannel <- verifyResult{isError: true, err: vErr, provider: provider.getName()}
	} else {
		resultsChannel <- verifyResult{isError: false, provider: provider.getName()}
	}

}
