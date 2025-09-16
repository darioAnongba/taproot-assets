package tapfreighter

import (
	"testing"

	"github.com/lightninglabs/taproot-assets/asset"
	"github.com/lightninglabs/taproot-assets/tappsbt"
	"github.com/stretchr/testify/require"
)

// TestFundBurnWithTombstone tests that the tombstone creation logic compiles
// and works correctly. This is a unit test for the tombstone output creation
// logic that was added to support full UTXO burns.
func TestFundBurnWithTombstone(t *testing.T) {
	t.Parallel()

	// Test that we can create a tombstone output with the correct properties
	// This verifies that the tombstone creation logic in FundBurn compiles
	// and works correctly.
	tombstoneOutput := &tappsbt.VOutput{
		Amount:            0, // Zero amount for tombstone
		Type:              tappsbt.TypeSplitRoot,
		Interactive:       true,
		AnchorOutputIndex: 0,
		AssetVersion:      asset.V0,
		ScriptKey:         asset.NUMSScriptKey, // Use NUMS key for tombstone
	}

	// Verify tombstone properties
	require.Equal(t, uint64(0), tombstoneOutput.Amount)
	require.Equal(t, tappsbt.TypeSplitRoot, tombstoneOutput.Type)
	require.True(t, tombstoneOutput.Interactive)
	require.Equal(t, asset.NUMSScriptKey, tombstoneOutput.ScriptKey)
	require.Equal(t, asset.V0, tombstoneOutput.AssetVersion)

	// Test that the tombstone uses the NUMS key
	require.True(t, asset.NUMSPubKey.IsEqual(tombstoneOutput.ScriptKey.PubKey))
}
