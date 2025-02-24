package handler

import (
	"context"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/hetu-project/hetu-chaoschain/state"
	"github.com/hetu-project/hetu-chaoschain/tx"
)

type TxHandler interface {
	Check(ctx context.Context, st *state.State, btx *tx.HACTx) (res *abcitypes.ResponseCheckTx, err error)
	NewContext(ctx context.Context)
	Prepare(ctx context.Context, st *state.State, btx *tx.HACTx, code tx.VoteCode) (res *abcitypes.ExecTxResult, err error)
	Process(ctx context.Context, st *state.State, btx *tx.HACTx, code tx.VoteCode) (res *abcitypes.ExecTxResult, err error)
}
