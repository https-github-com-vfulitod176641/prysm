package testing

import (
	"bytes"
	"context"
	"time"

	"github.com/pkg/errors"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/go-ssz"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/epoch/precompute"
	blockfeed "github.com/prysmaticlabs/prysm/beacon-chain/core/feed/block"
	opfeed "github.com/prysmaticlabs/prysm/beacon-chain/core/feed/operation"
	statefeed "github.com/prysmaticlabs/prysm/beacon-chain/core/feed/state"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/helpers"
	"github.com/prysmaticlabs/prysm/beacon-chain/db"
	stateTrie "github.com/prysmaticlabs/prysm/beacon-chain/state"
	pb "github.com/prysmaticlabs/prysm/proto/beacon/p2p/v1"
	"github.com/prysmaticlabs/prysm/shared/event"
	"github.com/prysmaticlabs/prysm/shared/params"
	"github.com/sirupsen/logrus"
)

// ChainService defines the mock interface for testing
type ChainService struct {
	State                       *stateTrie.BeaconState
	Root                        []byte
	Block                       *ethpb.SignedBeaconBlock
	FinalizedCheckPoint         *ethpb.Checkpoint
	CurrentJustifiedCheckPoint  *ethpb.Checkpoint
	PreviousJustifiedCheckPoint *ethpb.Checkpoint
	BlocksReceived              []*ethpb.SignedBeaconBlock
	Balance                     *precompute.Balance
	Genesis                     time.Time
	Fork                        *pb.Fork
	DB                          db.Database
	stateNotifier               statefeed.Notifier
	blockNotifier               blockfeed.Notifier
	opNotifier                  opfeed.Notifier
	ValidAttestation            bool
}

// StateNotifier mocks the same method in the chain service.
func (ms *ChainService) StateNotifier() statefeed.Notifier {
	if ms.stateNotifier == nil {
		ms.stateNotifier = &MockStateNotifier{}
	}
	return ms.stateNotifier
}

// BlockNotifier mocks the same method in the chain service.
func (ms *ChainService) BlockNotifier() blockfeed.Notifier {
	if ms.blockNotifier == nil {
		ms.blockNotifier = &MockBlockNotifier{}
	}
	return ms.blockNotifier
}

// MockBlockNotifier mocks the block notifier.
type MockBlockNotifier struct {
	feed *event.Feed
}

// BlockFeed returns a block feed.
func (msn *MockBlockNotifier) BlockFeed() *event.Feed {
	if msn.feed == nil {
		msn.feed = new(event.Feed)
	}
	return msn.feed
}

// MockStateNotifier mocks the state notifier.
type MockStateNotifier struct {
	feed *event.Feed
}

// StateFeed returns a state feed.
func (msn *MockStateNotifier) StateFeed() *event.Feed {
	if msn.feed == nil {
		msn.feed = new(event.Feed)
	}
	return msn.feed
}

// OperationNotifier mocks the same method in the chain service.
func (ms *ChainService) OperationNotifier() opfeed.Notifier {
	if ms.opNotifier == nil {
		ms.opNotifier = &MockOperationNotifier{}
	}
	return ms.opNotifier
}

// MockOperationNotifier mocks the operation notifier.
type MockOperationNotifier struct {
	feed *event.Feed
}

// OperationFeed returns an operation feed.
func (mon *MockOperationNotifier) OperationFeed() *event.Feed {
	if mon.feed == nil {
		mon.feed = new(event.Feed)
	}
	return mon.feed
}

// ReceiveBlock mocks ReceiveBlock method in chain service.
func (ms *ChainService) ReceiveBlock(ctx context.Context, block *ethpb.SignedBeaconBlock) error {
	return nil
}

// ReceiveBlockNoVerify mocks ReceiveBlockNoVerify method in chain service.
func (ms *ChainService) ReceiveBlockNoVerify(ctx context.Context, block *ethpb.SignedBeaconBlock) error {
	return nil
}

// ReceiveBlockNoPubsub mocks ReceiveBlockNoPubsub method in chain service.
func (ms *ChainService) ReceiveBlockNoPubsub(ctx context.Context, block *ethpb.SignedBeaconBlock) error {
	return nil
}

// ReceiveBlockNoPubsubForkchoice mocks ReceiveBlockNoPubsubForkchoice method in chain service.
func (ms *ChainService) ReceiveBlockNoPubsubForkchoice(ctx context.Context, block *ethpb.SignedBeaconBlock) error {
	if ms.State == nil {
		ms.State = &stateTrie.BeaconState{}
	}
	if !bytes.Equal(ms.Root, block.Block.ParentRoot) {
		return errors.Errorf("wanted %#x but got %#x", ms.Root, block.Block.ParentRoot)
	}
	if err := ms.State.SetSlot(block.Block.Slot); err != nil {
		return err
	}
	ms.BlocksReceived = append(ms.BlocksReceived, block)
	signingRoot, err := ssz.HashTreeRoot(block.Block)
	if err != nil {
		return err
	}
	if ms.DB != nil {
		if err := ms.DB.SaveBlock(ctx, block); err != nil {
			return err
		}
		logrus.Infof("Saved block with root: %#x at slot %d", signingRoot, block.Block.Slot)
	}
	ms.Root = signingRoot[:]
	ms.Block = block
	return nil
}

// HeadSlot mocks HeadSlot method in chain service.
func (ms *ChainService) HeadSlot() uint64 {
	if ms.State == nil {
		return 0
	}
	return ms.State.Slot()
}

// HeadRoot mocks HeadRoot method in chain service.
func (ms *ChainService) HeadRoot(ctx context.Context) ([]byte, error) {
	return ms.Root, nil

}

// HeadBlock mocks HeadBlock method in chain service.
func (ms *ChainService) HeadBlock(context.Context) (*ethpb.SignedBeaconBlock, error) {
	return ms.Block, nil
}

// HeadState mocks HeadState method in chain service.
func (ms *ChainService) HeadState(context.Context) (*stateTrie.BeaconState, error) {
	return ms.State, nil
}

// CurrentFork mocks HeadState method in chain service.
func (ms *ChainService) CurrentFork() *pb.Fork {
	return ms.Fork
}

// FinalizedCheckpt mocks FinalizedCheckpt method in chain service.
func (ms *ChainService) FinalizedCheckpt() *ethpb.Checkpoint {
	return ms.FinalizedCheckPoint
}

// CurrentJustifiedCheckpt mocks CurrentJustifiedCheckpt method in chain service.
func (ms *ChainService) CurrentJustifiedCheckpt() *ethpb.Checkpoint {
	return ms.CurrentJustifiedCheckPoint
}

// PreviousJustifiedCheckpt mocks PreviousJustifiedCheckpt method in chain service.
func (ms *ChainService) PreviousJustifiedCheckpt() *ethpb.Checkpoint {
	return ms.PreviousJustifiedCheckPoint
}

// ReceiveAttestation mocks ReceiveAttestation method in chain service.
func (ms *ChainService) ReceiveAttestation(context.Context, *ethpb.Attestation) error {
	return nil
}

// ReceiveAttestationNoPubsub mocks ReceiveAttestationNoPubsub method in chain service.
func (ms *ChainService) ReceiveAttestationNoPubsub(context.Context, *ethpb.Attestation) error {
	return nil
}

// HeadValidatorsIndices mocks the same method in the chain service.
func (ms *ChainService) HeadValidatorsIndices(epoch uint64) ([]uint64, error) {
	if ms.State == nil {
		return []uint64{}, nil
	}
	return helpers.ActiveValidatorIndices(ms.State, epoch)
}

// HeadSeed mocks the same method in the chain service.
func (ms *ChainService) HeadSeed(epoch uint64) ([32]byte, error) {
	return helpers.Seed(ms.State, epoch, params.BeaconConfig().DomainBeaconAttester)
}

// GenesisTime mocks the same method in the chain service.
func (ms *ChainService) GenesisTime() time.Time {
	return ms.Genesis
}

// CurrentSlot mocks the same method in the chain service.
func (ms *ChainService) CurrentSlot() uint64 {
	return uint64(time.Now().Unix()-ms.Genesis.Unix()) / params.BeaconConfig().SecondsPerSlot
}

// Participation mocks the same method in the chain service.
func (ms *ChainService) Participation(epoch uint64) *precompute.Balance {
	return ms.Balance
}

// IsValidAttestation always returns true.
func (ms *ChainService) IsValidAttestation(ctx context.Context, att *ethpb.Attestation) bool {
	return ms.ValidAttestation
}

// ClearCachedStates does nothing.
func (ms *ChainService) ClearCachedStates() {}
