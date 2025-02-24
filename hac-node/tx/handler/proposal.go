package handler

import (
	"context"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	cmtlog "github.com/cometbft/cometbft/libs/log"
	"github.com/hetu-project/hetu-chaoschain/state"
	"github.com/hetu-project/hetu-chaoschain/tx"
	"github.com/hetu-project/hetu-chaoschain/types"
)

type ProposalTxHandler struct {
	logger cmtlog.Logger

	validatorSet map[uint64]bool
}

func NewProposalTxHandler(logger cmtlog.Logger) (h *ProposalTxHandler) {
	logger = logger.With("module", "stakeTx")
	h = &ProposalTxHandler{
		logger:       logger,
		validatorSet: make(map[uint64]bool),
	}
	return
}

func (h *ProposalTxHandler) Check(ctx context.Context, st *state.State, btx *tx.HACTx) (res *abcitypes.ResponseCheckTx, err error) {
	res = &abcitypes.ResponseCheckTx{Code: 0}
	// stx := btx.Tx.(*tx.ProposalTx)
	// _, err1 := st.Proposal(stx, btx.Validator, true, tx.VoteIgnoreProposal)
	// if err1 != nil {
	// 	h.logger.Info("CheckTx ProposalTx fail", "err", err1)
	// 	res.Code = 1
	// 	res.Log = err1.Error()
	// }
	// _, err1 = st.Proposal(stx, btx.Validator, true, tx.VoteProcessProposal)
	// if err1 != nil {
	// 	h.logger.Info("CheckTx ProposalTx fail", "err", err1)
	// 	res.Code = 1
	// 	res.Log = err1.Error()
	// }
	return
}

func (h *ProposalTxHandler) NewContext(ctx context.Context) {
	h.validatorSet = make(map[uint64]bool)
}

func (h *ProposalTxHandler) handle(ctx context.Context, st *state.State, btx *tx.HACTx, code tx.VoteCode) (res *abcitypes.ExecTxResult, err error) {
	if _, ok := h.validatorSet[btx.Validator]; ok {
		return nil, state.ErrOneActionInOneBlock
	}
	wtx := btx.Tx.(*tx.ProposalTx)
	event, err := st.Proposal(wtx, btx.Validator, false, code)
	if err != nil {
		return nil, err
	}
	h.validatorSet[btx.Validator] = true
	res = &abcitypes.ExecTxResult{}
	if event != nil {
		res.Events = []abcitypes.Event{types.EncodeEventProposal(event)}
	}
	return
}

func (h *ProposalTxHandler) Prepare(ctx context.Context, st *state.State, btx *tx.HACTx, code tx.VoteCode) (res *abcitypes.ExecTxResult, err error) {
	return h.handle(ctx, st, btx, code)
}

func (h *ProposalTxHandler) Process(ctx context.Context, st *state.State, btx *tx.HACTx, code tx.VoteCode) (res *abcitypes.ExecTxResult, err error) {
	return h.handle(ctx, st, btx, code)
}
