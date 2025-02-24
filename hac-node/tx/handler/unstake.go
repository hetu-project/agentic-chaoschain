package handler

import (
	"context"
	"fmt"
	"strconv"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	cmtlog "github.com/cometbft/cometbft/libs/log"
	"github.com/hetu-project/hetu-chaoschain/state"
	"github.com/hetu-project/hetu-chaoschain/tx"
	"github.com/hetu-project/hetu-chaoschain/types"
)

type UnStakeTxHandler struct {
	logger cmtlog.Logger

	validatorSet map[uint64]bool
}

func NewUnStakeTxHandler(logger cmtlog.Logger) (h *UnStakeTxHandler) {
	logger = logger.With("module", "retractTx")
	h = &UnStakeTxHandler{
		logger:       logger,
		validatorSet: make(map[uint64]bool),
	}
	return
}

func (h *UnStakeTxHandler) Check(ctx context.Context, st *state.State, btx *tx.HACTx) (res *abcitypes.ResponseCheckTx, err error) {
	res = &abcitypes.ResponseCheckTx{Code: 0}
	utx := btx.Tx.(*tx.RetractTx)
	_, err1 := st.UnStake(utx, btx.Validator, true)
	if err1 != nil {
		h.logger.Info("CheckTx retract fail", "err", err1)
		res.Code = 1
		res.Log = err1.Error()
	}
	return
}

func (h *UnStakeTxHandler) NewContext(ctx context.Context) {
	h.validatorSet = make(map[uint64]bool)
}

func (h *UnStakeTxHandler) handle(ctx context.Context, st *state.State, btx *tx.HACTx) (res *abcitypes.ExecTxResult, err error) {
	if _, ok := h.validatorSet[btx.Validator]; ok {
		return nil, state.ErrOneActionInOneBlock
	}
	utx := btx.Tx.(*tx.RetractTx)
	event, err := st.UnStake(utx, btx.Validator, false)
	if err != nil {
		return nil, err
	}

	h.validatorSet[btx.Validator] = true
	res = &abcitypes.ExecTxResult{}
	res.Events = append(res.Events, abcitypes.Event{
		Type: types.EventUnStakeType,
		Attributes: []abcitypes.EventAttribute{
			{Key: "validator", Value: strconv.FormatUint(event.Validator, 10), Index: true},
			{Key: "amount", Value: fmt.Sprintf("%d", event.Amount), Index: false},
			{Key: "addr", Value: fmt.Sprintf("%v", event.Address), Index: false},
		},
	})
	return
}

func (h *UnStakeTxHandler) Prepare(ctx context.Context, st *state.State, btx *tx.HACTx, code tx.VoteCode) (res *abcitypes.ExecTxResult, err error) {
	return h.handle(ctx, st, btx)
}

func (h *UnStakeTxHandler) Process(ctx context.Context, st *state.State, btx *tx.HACTx, code tx.VoteCode) (res *abcitypes.ExecTxResult, err error) {
	return h.handle(ctx, st, btx)
}
