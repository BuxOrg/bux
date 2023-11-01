package chainstate

import (
	"context"
	"errors"
	"fmt"
)

// VerifyMerkleRoots will try to verify merkle roots with all available providers
func (c *Client) VerifyMerkleRoots(ctx context.Context, merkleRoots []MerkleRootConfirmationRequestItem) error {
	pulseProvider := createPulseProvider(c)
	merkleRootsRes, err := pulseProvider.verifyMerkleRoots(ctx, c, merkleRoots)
	if err != nil {
		debugLog(c, "", fmt.Sprintf("verify merkle root error: %s from Pulse", err))
		return err
	}

	if merkleRootsRes.ConfirmationState == Invalid {
		c.options.logger.Warn(context.Background(), "Warn: Not all merkle roots confirmed")
		return errors.New("not all merkle roots confirmed")
	}
	return nil
}

func createPulseProvider(c *Client) pulseClientProvider {
	return pulseClientProvider{
		url:       c.options.config.pulseClient.url,
		authToken: c.options.config.pulseClient.authToken,
	}
}
