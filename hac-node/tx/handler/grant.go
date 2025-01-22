package handler

import (
	"context"

	"github.com/calehh/hac-app/state"
	"github.com/calehh/hac-app/tx"
	"github.com/calehh/hac-app/types"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	cmtlog "github.com/cometbft/cometbft/libs/log"
)

type GrantTxHandler struct {
	logger cmtlog.Logger
}

func NewGrantTxHandler(logger cmtlog.Logger) (h *GrantTxHandler) {
	logger = logger.With("module", "grantTx")
	h = &GrantTxHandler{
		logger: logger,
	}
	return
}

func (h *GrantTxHandler) Check(ctx context.Context, st *state.State, btx *tx.HACTx) (res *abcitypes.ResponseCheckTx, err error) {
	res = &abcitypes.ResponseCheckTx{Code: 0}
	return
}

func (h *GrantTxHandler) NewContext(ctx context.Context) {}

func (h *GrantTxHandler) handle(ctx context.Context, st *state.State, btx *tx.HACTx, code tx.VoteCode) (res *abcitypes.ExecTxResult, err error) {
	wtx := btx.Tx.(*tx.GrantTx)
	res = &abcitypes.ExecTxResult{}
	for _, grant := range wtx.Grants {
		event, err1 := st.Grant(btx.Validator, grant.Pubkey, grant.Amount, grant.AgentUrl, grant.Name, code)
		if err1 != nil {
			err = err1
			return
		}
		if event != nil {
			res.Events = []abcitypes.Event{types.EncodeEventGrant(event)}
		}
	}
	return
}

func (h *GrantTxHandler) Prepare(ctx context.Context, st *state.State, btx *tx.HACTx, code tx.VoteCode) (res *abcitypes.ExecTxResult, err error) {
	return h.handle(ctx, st, btx, code)
}

func (h *GrantTxHandler) Process(ctx context.Context, st *state.State, btx *tx.HACTx, code tx.VoteCode) (res *abcitypes.ExecTxResult, err error) {
	return h.handle(ctx, st, btx, code)
}
