package handler

import (
	"context"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	cmtlog "github.com/cometbft/cometbft/libs/log"
	"github.com/hetu-project/hetu-chaoschain/state"
	"github.com/hetu-project/hetu-chaoschain/tx"
	"github.com/hetu-project/hetu-chaoschain/types"
)

type DiscussionTxHandler struct {
	logger cmtlog.Logger
}

func NewDiscussionTxHandler(logger cmtlog.Logger) (h *DiscussionTxHandler) {
	logger = logger.With("module", "stakeTx")
	h = &DiscussionTxHandler{
		logger: logger,
	}
	return
}

func (h *DiscussionTxHandler) Check(ctx context.Context, st *state.State, btx *tx.HACTx) (res *abcitypes.ResponseCheckTx, err error) {
	res = &abcitypes.ResponseCheckTx{Code: 0}
	stx := btx.Tx.(*tx.DiscussionTx)
	_, err1 := st.Dicussion(stx, btx.Validator, true)
	if err1 != nil {
		h.logger.Info("CheckTx stake fail", "err", err1)
		res.Code = 1
		res.Log = err1.Error()
	}
	return
}

func (h *DiscussionTxHandler) NewContext(ctx context.Context) {}

func (h *DiscussionTxHandler) handle(ctx context.Context, st *state.State, btx *tx.HACTx) (res *abcitypes.ExecTxResult, err error) {
	wtx := btx.Tx.(*tx.DiscussionTx)
	event, err := st.Dicussion(wtx, btx.Validator, false)
	if err != nil {
		return nil, err
	}
	res = &abcitypes.ExecTxResult{}
	if event != nil {
		res.Events = []abcitypes.Event{types.EncodeEventDiscussion(event)}
	}
	return
}

func (h *DiscussionTxHandler) Prepare(ctx context.Context, st *state.State, btx *tx.HACTx, code tx.VoteCode) (res *abcitypes.ExecTxResult, err error) {
	return h.handle(ctx, st, btx)
}

func (h *DiscussionTxHandler) Process(ctx context.Context, st *state.State, btx *tx.HACTx, code tx.VoteCode) (res *abcitypes.ExecTxResult, err error) {
	return h.handle(ctx, st, btx)
}
