package app

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/calehh/hac-app/agent"
	"github.com/calehh/hac-app/config"
	"github.com/calehh/hac-app/state"
	"github.com/calehh/hac-app/tx"
	"github.com/calehh/hac-app/tx/handler"
	"github.com/calehh/hac-app/types"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	cmtlog "github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/store"
	"github.com/ethereum/go-ethereum/common"
)

type finalizeBlock struct {
	Height uint64
	Hash   common.Hash
}

func (b *finalizeBlock) Set(blk *abcitypes.RequestFinalizeBlock) {
	b.Height = uint64(blk.Height)
	b.Hash = common.BytesToHash(blk.Hash)
}

var _ abcitypes.Application = &HACApp{}

type HACApp struct {
	cfg    *config.HACAppConfig
	logger cmtlog.Logger

	db       *state.StateDB
	lastBlk  finalizeBlock
	txHdlrs  map[tx.HACTxType]handler.TxHandler
	queriers map[string]Querier

	st *state.State
}

func NewHACApp(cfg *config.HACAppConfig, agentClient agent.Client, logger cmtlog.Logger) (app *HACApp, err error) {
	logger = logger.With("module", "app")

	dir := cfg.Home + "/data"
	db, err := state.NewStateDB(dir, logger)
	if err != nil {
		return nil, err
	}

	app = &HACApp{
		cfg:      cfg,
		logger:   logger,
		db:       db,
		txHdlrs:  make(map[tx.HACTxType]handler.TxHandler),
		queriers: make(map[string]Querier),
	}
	app.registerTxHandler()
	app.registerQuerier()
	return
}

func (app *HACApp) Start(bs *store.BlockStore) {
	height := app.db.Header().Height
	if height > 0 {
		blk := bs.LoadBlock(int64(height))
		if blk == nil {
			panic("unexpected BlockStore")
		}
		app.lastBlk.Height = height
		app.lastBlk.Hash = common.BytesToHash(blk.Hash())
	}
}

func (app *HACApp) Stop() {
	err := app.db.Close()
	if err != nil {
		app.logger.Error("close db fail", "err", err)
	}
	app.logger.Info("HAC app stopped")
}

func (app *HACApp) registerTxHandler() {
	app.txHdlrs = map[tx.HACTxType]handler.TxHandler{
		tx.HACTxTypeRetract:        handler.NewUnStakeTxHandler(app.logger),
		tx.HACTxTypeSettleProposal: handler.NewSettleProposalTxHandler(app.logger),
		tx.HACTxTypeProposal:       handler.NewProposalTxHandler(app.logger),
		tx.HACTxTypeDiscussion:     handler.NewDiscussionTxHandler(app.logger),
		tx.HACTxTypeGrant:          handler.NewGrantTxHandler(app.logger),
	}
}

func (app *HACApp) registerQuerier() {
	aq := NewAccountQuerier(app.db, app.logger)
	vq := NewValidatorQuerier(app.db, app.logger)
	app.queriers["/accounts/"] = aq
	app.queriers["/validators/"] = vq
}

func (app *HACApp) InitChain(_ context.Context, chain *abcitypes.RequestInitChain) (res *abcitypes.ResponseInitChain, err error) {
	st := app.db.NewState()
	st.SetChainId(chain.ChainId)
	var appState types.GenesisAppState
	err = json.Unmarshal(chain.AppStateBytes, &appState)
	if err != nil {
		app.logger.Error("InitChain unmarshal app state fail", "err", err)
		return nil, err
	}
	agentInfoMap := make(map[string]types.AgentInfo)
	for _, v := range appState.Agents {
		agentInfoMap[v.Address] = v
	}
	for _, v := range chain.Validators {
		var acnt state.Account
		acnt.SetPubKey(v.PubKey.GetEd25519())
		acnt.Stake = uint64(v.Power) * config.GWeiPerPower(0)
		if info, ok := agentInfoMap[acnt.Address()]; ok {
			acnt.AgentUrl = info.AgentUrl
			acnt.Name = info.Name
		}
		app.logger.Info("InitChain", "validator", acnt.Address(), "agentUrl", acnt.AgentUrl, "name", acnt.Name)
		err = st.AddAccount(&acnt)
		if err != nil {
			if errors.Is(err, state.ErrAccountAlreadyExists) {
				continue
			}
			app.logger.Error("InitChain add account fail", "err", err)
			return nil, err
		}
	}
	var h common.Hash
	_, err = st.Update()
	if err != nil {
		app.logger.Error("InitChain update state fail", "err", err)
		return nil, err
	}
	err = st.SetManifest(appState.Manifest)
	if err != nil {
		app.logger.Error("InitChain set manifest fail", "err", err)
		return nil, err
	}
	h, err = app.db.SetState(st)
	if err != nil {
		app.logger.Error("InitChain apply state fail", "err", err)
		return nil, err
	}
	return &abcitypes.ResponseInitChain{
		AppHash: h.Bytes(),
	}, nil
}

func (app *HACApp) Info(ctx context.Context, info *abcitypes.RequestInfo) (*abcitypes.ResponseInfo, error) {
	header := app.db.Header()
	return &abcitypes.ResponseInfo{
		LastBlockHeight:  int64(header.Height),
		LastBlockAppHash: header.Hash,
	}, nil
}

func (app *HACApp) ExtendVote(_ context.Context, extend *abcitypes.RequestExtendVote) (*abcitypes.ResponseExtendVote, error) {
	return &abcitypes.ResponseExtendVote{}, nil
}

func (app *HACApp) VerifyVoteExtension(_ context.Context, verify *abcitypes.RequestVerifyVoteExtension) (*abcitypes.ResponseVerifyVoteExtension, error) {
	return &abcitypes.ResponseVerifyVoteExtension{}, nil
}

func (app *HACApp) ApplySnapshotChunk(context.Context, *abcitypes.RequestApplySnapshotChunk) (*abcitypes.ResponseApplySnapshotChunk, error) {
	return nil, nil
}

func (app *HACApp) ListSnapshots(context.Context, *abcitypes.RequestListSnapshots) (*abcitypes.ResponseListSnapshots, error) {
	return nil, nil
}

func (app *HACApp) LoadSnapshotChunk(context.Context, *abcitypes.RequestLoadSnapshotChunk) (*abcitypes.ResponseLoadSnapshotChunk, error) {
	return nil, nil
}

func (app *HACApp) OfferSnapshot(context.Context, *abcitypes.RequestOfferSnapshot) (*abcitypes.ResponseOfferSnapshot, error) {
	return nil, nil
}
