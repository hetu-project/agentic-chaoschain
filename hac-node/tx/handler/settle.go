package handler

import (
	"context"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	cmtlog "github.com/cometbft/cometbft/libs/log"
	"github.com/hetu-project/hetu-chaoschain/state"
	"github.com/hetu-project/hetu-chaoschain/tx"
	"github.com/hetu-project/hetu-chaoschain/types"
)

type SettleProposalTxHandler struct {
	logger cmtlog.Logger

	validatorSet map[uint64]bool
}

func NewSettleProposalTxHandler(logger cmtlog.Logger) (h *SettleProposalTxHandler) {
	logger = logger.With("module", "stakeTx")
	h = &SettleProposalTxHandler{
		logger:       logger,
		validatorSet: make(map[uint64]bool),
	}
	return
}

func (h *SettleProposalTxHandler) Check(ctx context.Context, st *state.State, btx *tx.HACTx) (res *abcitypes.ResponseCheckTx, err error) {
	res = &abcitypes.ResponseCheckTx{Code: 0}
	stx := btx.Tx.(*tx.SettleProposalTx)
	_, err1 := st.SettleProposal(stx, btx.Validator, true, tx.VoteAcceptProposal)
	if err1 != nil {
		h.logger.Info("CheckTx SettleProposalTx fail", "err", err1)
		res.Code = 1
		res.Log = err1.Error()
	}
	_, err1 = st.SettleProposal(stx, btx.Validator, true, tx.VoteRejectProposal)
	if err1 != nil {
		h.logger.Info("CheckTx SettleProposalTx fail", "err", err1)
		res.Code = 1
		res.Log = err1.Error()
	}
	return
}

func (h *SettleProposalTxHandler) NewContext(ctx context.Context) {
	h.validatorSet = make(map[uint64]bool)
}

func (h *SettleProposalTxHandler) handle(ctx context.Context, st *state.State, btx *tx.HACTx, code tx.VoteCode) (res *abcitypes.ExecTxResult, err error) {
	if _, ok := h.validatorSet[btx.Validator]; ok {
		return nil, state.ErrOneActionInOneBlock
	}
	wtx := btx.Tx.(*tx.SettleProposalTx)
	event, err := st.SettleProposal(wtx, btx.Validator, false, code)
	if err != nil {
		return nil, err
	}
	h.validatorSet[btx.Validator] = true
	res = &abcitypes.ExecTxResult{}
	if event != nil {
		res.Events = []abcitypes.Event{types.EncodeEventSettleProposal(event)}
	}
	return
}

func (h *SettleProposalTxHandler) Prepare(ctx context.Context, st *state.State, btx *tx.HACTx, code tx.VoteCode) (res *abcitypes.ExecTxResult, err error) {
	return h.handle(ctx, st, btx, code)
}

func (h *SettleProposalTxHandler) Process(ctx context.Context, st *state.State, btx *tx.HACTx, code tx.VoteCode) (res *abcitypes.ExecTxResult, err error) {
	return h.handle(ctx, st, btx, code)
}
