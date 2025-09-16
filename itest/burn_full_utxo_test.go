package itest

import (
	"context"

	taprootassets "github.com/lightninglabs/taproot-assets"
	"github.com/lightninglabs/taproot-assets/address"
	"github.com/lightninglabs/taproot-assets/asset"
	"github.com/lightninglabs/taproot-assets/tappsbt"
	"github.com/lightninglabs/taproot-assets/taprpc"
	wrpc "github.com/lightninglabs/taproot-assets/taprpc/assetwalletrpc"
	"github.com/lightninglabs/taproot-assets/taprpc/mintrpc"
	"github.com/stretchr/testify/require"
)

// testFullBurnUTXO verifies that a whole-UTXO burn works by creating a
// tombstone split root output to the NUMS key.
func testFullBurnUTXO(t *harnessTest) {
	minerClient := t.lndHarness.Miner().Client
	ctxb := context.Background()
	ctxt, cancel := context.WithTimeout(ctxb, defaultWaitTimeout)
	defer cancel()

	// Mint a simple asset and a collectible to have a realistic setup.
	rpcAssets := MintAssetsConfirmBatch(
		t.t, minerClient, t.tapd, []*mintrpc.MintAssetRequest{
			simpleAssets[0], simpleAssets[1],
		},
	)
	simpleAsset := rpcAssets[0]
	simpleAssetGen := simpleAsset.AssetGenesis

	var simpleAssetID [32]byte
	copy(simpleAssetID[:], simpleAssetGen.AssetId)

	// Fan out to isolate a single large output in its own anchor so a full
	// burn can target it.
	scriptKey1, anchorKeyDesc1 := DeriveKeys(t.t, t.tapd)
	scriptKey2, _ := DeriveKeys(t.t, t.tapd)
	chainParams := &address.RegressionNetTap

	// Create two outputs in the same anchor, and rely on a later single
	// output anchor for the large value.
	vPkt := tappsbt.ForInteractiveSend(
		simpleAssetID, 1100, scriptKey1, 0, 0, 0, anchorKeyDesc1, asset.V0,
		chainParams,
	)
	tappsbt.AddOutput(vPkt, 1200, scriptKey2, 0, anchorKeyDesc1, asset.V0)

	// Anchor the fan-out.
	fundResp := fundPacket(t, t.tapd, vPkt)
	signResp, err := t.tapd.SignVirtualPsbt(ctxt, &wrpc.SignVirtualPsbtRequest{
		FundedPsbt: fundResp.FundedPsbt,
	})
	require.NoError(t.t, err)
	sendResp, err := t.tapd.AnchorVirtualPsbts(
		ctxt, &wrpc.AnchorVirtualPsbtsRequest{VirtualPsbts: [][]byte{signResp.SignedPsbt}},
	)
	require.NoError(t.t, err)
	ConfirmAndAssertOutboundTransfer(
		t.t, minerClient, t.tapd, sendResp, simpleAssetGen.AssetId,
		[]uint64{1100, 1200}, 0, 1,
	)

	// Verify pre-burn balance.
	AssertBalanceByID(t.t, t.tapd, simpleAssetGen.AssetId, simpleAsset.Amount)

	// Perform a full burn of the isolated largest output (1200).
	fullBurnAmt := uint64(1200)
	burnResp, err := t.tapd.BurnAsset(ctxt, &taprpc.BurnAssetRequest{
		Asset: &taprpc.BurnAssetRequest_AssetId{AssetId: simpleAssetID[:]},
		AmountToBurn:     fullBurnAmt,
		ConfirmationText: taprootassets.AssetBurnConfirmationText,
	})
	require.NoError(t.t, err)

	// Expect burn output and tombstone split root.
	AssertAssetOutboundTransferWithOutputs(
		t.t, minerClient, t.tapd, burnResp.BurnTransfer,
		[][]byte{simpleAssetGen.AssetId}, []uint64{fullBurnAmt, 0}, 1, 2, 2, true,
	)

	// Balance reflects the full burn amount.
	AssertBalanceByID(
		t.t, t.tapd, simpleAssetGen.AssetId, simpleAsset.Amount-fullBurnAmt,
	)
}