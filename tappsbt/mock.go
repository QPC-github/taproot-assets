package tappsbt

import (
	"testing"

	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/txscript"
	"github.com/lightninglabs/taproot-assets/address"
	"github.com/lightninglabs/taproot-assets/asset"
	"github.com/lightninglabs/taproot-assets/commitment"
	"github.com/lightninglabs/taproot-assets/internal/test"
	"github.com/lightningnetwork/lnd/keychain"
)

var (
	testParams = &address.MainNetTap
)

// RandPacket generates a random virtual packet for testing purposes.
func RandPacket(t testing.TB) *VPacket {
	testPubKey := test.RandPubKey(t)
	op := test.RandOp(t)
	keyDesc := keychain.KeyDescriptor{
		PubKey: testPubKey,
		KeyLocator: keychain.KeyLocator{
			Family: 123,
			Index:  456,
		},
	}
	inputScriptKey := asset.NewScriptKeyBip86(keyDesc)
	inputScriptKey.Tweak = []byte("merkle root")

	bip32Derivation, trBip32Derivation := Bip32DerivationFromKeyDesc(
		keyDesc, testParams.HDCoinType,
	)
	bip32Derivations := []*psbt.Bip32Derivation{bip32Derivation}
	trBip32Derivations := []*psbt.TaprootBip32Derivation{trBip32Derivation}
	testAsset := asset.RandAsset(t, asset.Normal)
	testAsset.ScriptKey = inputScriptKey

	testOutputAsset := asset.RandAsset(t, asset.Normal)
	testOutputAsset.ScriptKey = asset.NewScriptKeyBip86(keyDesc)

	// The raw key won't be serialized within the asset, so let's blank it
	// out here to get a fully, byte-by-byte comparable PSBT.
	testAsset.GroupKey.RawKey = keychain.KeyDescriptor{}
	testOutputAsset.GroupKey.RawKey = keychain.KeyDescriptor{}
	testOutputAsset.ScriptKey.TweakedScriptKey = nil
	leaf1 := txscript.TapLeaf{
		LeafVersion: txscript.BaseLeafVersion,
		Script:      []byte("not a valid script"),
	}
	testPreimage1 := commitment.NewPreimageFromLeaf(leaf1)
	testPreimage2 := commitment.NewPreimageFromBranch(
		txscript.NewTapBranch(leaf1, leaf1),
	)

	vPacket := &VPacket{
		Inputs: []*VInput{{
			PrevID: asset.PrevID{
				OutPoint:  op,
				ID:        asset.RandID(t),
				ScriptKey: asset.RandSerializedKey(t),
			},
			Anchor: Anchor{
				Value:             777,
				PkScript:          []byte("anchor pkscript"),
				SigHashType:       txscript.SigHashSingle,
				InternalKey:       testPubKey,
				MerkleRoot:        []byte("merkle root"),
				TapscriptSibling:  []byte("sibling"),
				Bip32Derivation:   bip32Derivations,
				TrBip32Derivation: trBip32Derivations,
			},
		}, {
			// Empty input.
		}},
		Outputs: []*VOutput{{
			Amount:                             123,
			Type:                               TypeSplitRoot,
			Interactive:                        true,
			AnchorOutputIndex:                  0,
			AnchorOutputInternalKey:            testPubKey,
			AnchorOutputBip32Derivation:        bip32Derivations,
			AnchorOutputTaprootBip32Derivation: trBip32Derivations,
			Asset:                              testOutputAsset,
			ScriptKey:                          testOutputAsset.ScriptKey,
			SplitAsset:                         testOutputAsset,
			AnchorOutputTapscriptSibling:       testPreimage1,
		}, {
			Amount: 345,
			Type:   TypeSplitRoot,

			Interactive:                        false,
			AnchorOutputIndex:                  1,
			AnchorOutputInternalKey:            testPubKey,
			AnchorOutputBip32Derivation:        bip32Derivations,
			AnchorOutputTaprootBip32Derivation: trBip32Derivations,
			Asset:                              testOutputAsset,
			ScriptKey:                          testOutputAsset.ScriptKey,
			AnchorOutputTapscriptSibling:       testPreimage2,
		}},
		ChainParams: testParams,
	}
	vPacket.SetInputAsset(0, testAsset, []byte("this is a proof"))

	return vPacket
}
