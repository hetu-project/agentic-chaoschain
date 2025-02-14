package app

import (
	"context"
	"errors"
	"time"

	"github.com/calehh/hac-app/state"
	"github.com/calehh/hac-app/tx"
	hac_types "github.com/calehh/hac-app/types"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/ethereum/go-ethereum/common"
)

var (
	ErrOnlySupportOneGrant     = errors.New("only support one grant in one tx")
	ErrMultiProposalInOneBlock = errors.New("multi proposal in one block")
	ErrUnexpectedTxProcess     = errors.New("unexpected tx process")
	ErrUnexpectedGrantTxs      = errors.New("unexpected grants")
)

func (app *HACApp) getState(blkHash *common.Hash) (st *state.State) {
	st = app.db.NewState()
	app.st = st
	return
}

func (app *HACApp) parseTx(txDat []byte, allowNonceGap bool) (btx *tx.HACTx, err error) {
	btx, err = tx.UnmarshalHACTx(txDat)
	if err != nil {
		return
	}
	if btx != nil {
		_, err = app.db.State().Verify(btx, allowNonceGap)
	}
	return
}

func (app *HACApp) CheckTx(ctx context.Context, check *abcitypes.RequestCheckTx) (res *abcitypes.ResponseCheckTx, err error) {
	res = &abcitypes.ResponseCheckTx{Code: 0}
	app.logger.Info("check tx")
	btx, err := app.parseTx(check.Tx, true)
	if err != nil {
		app.logger.Error("parse tx fail", "err", err)
		res.Code = 1
		res.Log = err.Error()
		err = nil
		return
	}
	app.logger.Info("check tx", "type", btx.Type)
	h, ok := app.txHdlrs[btx.Type]
	if !ok {
		app.logger.Error("unsupported tx", "type", btx.Type)
		res.Code = 1
		res.Log = "unsupported tx"
		return
	}
	st := app.db.State()
	res, err = h.Check(ctx, st, btx)
	if err != nil {
		app.logger.Error("check tx fail", "err", err)
		res.Code = 1
		res.Log = err.Error()
		err = nil
	}

	return
}

func (app *HACApp) PrepareProposal(ctx context.Context, proposal *abcitypes.RequestPrepareProposal) (res *abcitypes.ResponsePrepareProposal, err error) {
	app.logger.Info("PrepareProposal")
	st := app.getState(nil)
	for _, h := range app.txHdlrs {
		h.NewContext(ctx)
	}
	proposerAct := false
	prepareTxs := make([][]byte, 0)
	for _, stx := range proposal.Txs {
		btx, err := app.parseTx(stx, false)
		if err != nil {
			app.logger.Error("unsupported tx, parse fail", "err", err)
			continue
		}
		if btx.Type == tx.HACTxTypeGrant || btx.Type == tx.HACTxTypeProposal || btx.Type == tx.HACTxTypeSettleProposal {
			if proposerAct == true {
				continue
			}
			prepareTxs = append(prepareTxs, stx)
			proposerAct = true
		} else {
			prepareTxs = append(prepareTxs, stx)
		}
	}

	code, err := app.getCode(ctx, st, prepareTxs)
	if err != nil {
		app.logger.Error("PrepareProposal getCode failed", "height", uint64(proposal.Height), "err", err)
		return &abcitypes.ResponsePrepareProposal{}, nil
	}
	txs := make([][]byte, 0)
	for _, stx := range prepareTxs {
		stTmp := st.Clone()
		btx, err := app.parseTx(stx, false)
		if err != nil {
			app.logger.Error("unsupported tx, parse fail", "err", err)
			continue
		}
		h, ok := app.txHdlrs[btx.Type]
		if !ok {
			app.logger.Error("unsupported tx", "type", btx.Type)
			continue
		}
		result, err := h.Prepare(ctx, stTmp, btx, code)
		if err != nil {
			app.logger.Error("prepare tx fail ", "type", btx.Type, "err", err)
			continue
		}
		if result == nil {
			app.logger.Error("prepare tx nil result ", "type", btx.Type)
			continue
		}
		if result.Code != 0 {
			app.logger.Error("prepare tx fail", "type", btx.Type, "code", result.Code)
			continue
		}
		st = stTmp
		txs = append(txs, stx)
	}
	return &abcitypes.ResponsePrepareProposal{Txs: txs}, nil
}

func (app *HACApp) finalize(ctx context.Context, st *state.State, txs [][]byte, proposer []byte, height uint64, code tx.VoteCode) (res []*abcitypes.ExecTxResult, events []abcitypes.Event, err error) {
	for _, h := range app.txHdlrs {
		h.NewContext(ctx)
	}
	res = make([]*abcitypes.ExecTxResult, len(txs))
	for i, stx := range txs {
		btx, err := app.parseTx(stx, false)
		if err != nil {
			app.logger.Error("unexpected tx, parse fail", "err", err)
			return nil, nil, err
		}
		h, ok := app.txHdlrs[btx.Type]
		if !ok {
			app.logger.Error("unexpected tx, no handler", "type", btx.Type)
			err = ErrUnexpectedTxProcess
			return nil, nil, err
		}
		result, err := h.Process(ctx, st, btx, code)
		if err != nil {
			app.logger.Error("unexpected process tx fail", "type", btx.Type, "err", err)
			err = ErrUnexpectedTxProcess
			return nil, nil, err
		}
		if result == nil {
			app.logger.Error("unexpected process tx nil result", "type", btx.Type)
			err = ErrUnexpectedTxProcess
			return nil, nil, err
		}
		if result.Code != 0 {
			app.logger.Error("unexpected process tx fail", "type", btx.Type, "err", err, "code", result.Code)
			err = ErrUnexpectedTxProcess
			return nil, nil, err
		}
		res[i] = result
	}
	return
}

func (app *HACApp) process(ctx context.Context, st *state.State, txs [][]byte, proposer []byte, height uint64, code tx.VoteCode) (res []*abcitypes.ExecTxResult, events []abcitypes.Event, err error) {
	for _, h := range app.txHdlrs {
		h.NewContext(ctx)
	}
	res = make([]*abcitypes.ExecTxResult, len(txs))
	for i, stx := range txs {
		btx, err := app.parseTx(stx, false)
		if err != nil {
			app.logger.Error("unexpected tx, parse fail", "err", err)
			return nil, nil, err
		}

		h, ok := app.txHdlrs[btx.Type]
		if !ok {
			app.logger.Error("unexpected tx, no handler", "type", btx.Type)
			err = ErrUnexpectedTxProcess
			return nil, nil, err
		}
		result, err := h.Process(ctx, st, btx, code)
		if err != nil {
			app.logger.Error("unexpected process tx fail", "type", btx.Type, "err", err)
			err = ErrUnexpectedTxProcess
			return nil, nil, err
		}
		if result == nil {
			app.logger.Error("unexpected process tx nil result", "type", btx.Type)
			err = ErrUnexpectedTxProcess
			return nil, nil, err
		}
		if result.Code != 0 {
			app.logger.Error("unexpected process tx fail", "type", btx.Type, "err", err, "code", result.Code)
			err = ErrUnexpectedTxProcess
			return nil, nil, err
		}
		res[i] = result
	}
	return
}

func (app *HACApp) ProcessProposal(ctx context.Context, proposal *abcitypes.RequestProcessProposal) (res *abcitypes.ResponseProcessProposal, err error) {
	app.logger.Info("ProcessProposal")
	res = &abcitypes.ResponseProcessProposal{Status: abcitypes.ResponseProcessProposal_REJECT}
	if len(proposal.Txs) == 0 {
		res.Status = abcitypes.ResponseProcessProposal_ACCEPT
		return res, nil
	}
	st := app.getState(nil)

	code, err := app.getCode(ctx, st, proposal.Txs)
	if err != nil {
		app.logger.Error("ProcessProposal getCode failed", "height", uint64(proposal.Height), "err", err)
		return res, nil
	}
	res.VoteCode = int64(code)

	_, _, err = app.process(ctx, st, proposal.Txs, proposal.ProposerAddress, uint64(proposal.Height), code)
	if err != nil {
		app.logger.Error("process fail", "err", err)
		return res, nil
	}
	res.Status = abcitypes.ResponseProcessProposal_ACCEPT
	app.logger.Info("proposal accepted", "height", proposal.Height, "voteCode", res.VoteCode)
	return res, nil
}

func (app *HACApp) FinalizeBlock(ctx context.Context, req *abcitypes.RequestFinalizeBlock) (*abcitypes.ResponseFinalizeBlock, error) {
	app.logger.Info("FinalizeBlock", "height", req.Height, "voteCode", req.VoteCode)
	app.lastBlk.Set(req)
	st := app.getState(nil)
	res, events, err := app.finalize(ctx, st, req.Txs, req.ProposerAddress, uint64(req.Height), tx.VoteCode(req.VoteCode))
	if err != nil {
		return nil, err
	}
	curVals, err := st.Validators()
	if err != nil {
		app.logger.Error("get validators fail", "err", err)
		return nil, err
	}
	h, err := st.Update()
	if err != nil {
		app.logger.Error("state update hash fail", "err", err)
		return nil, err
	}
	updateVals, err := st.ValidatorsUpdate(curVals)
	if err != nil {
		app.logger.Error("state update validators hash fail", "err", err)
		return nil, err
	}
	if len(updateVals) != 0 {
		events = append(events, hac_types.EncodeEventUpdateValiators(&hac_types.EventUpdateValiators{Updates: updateVals}))
	}
	return &abcitypes.ResponseFinalizeBlock{
		TxResults:        res,
		AppHash:          h.Bytes(),
		ValidatorUpdates: updateVals,
		Events:           events,
	}, nil
}

func (app *HACApp) Commit(ctx context.Context, commit *abcitypes.RequestCommit) (*abcitypes.ResponseCommit, error) {
	_, err := app.db.SetState(app.st)
	if err != nil {
		return nil, err
	}
	app.st = nil
	app.logger.Info("Commit")
	return &abcitypes.ResponseCommit{}, nil
}

func (app *HACApp) getCode(ctx context.Context, st *state.State, txs [][]byte) (code tx.VoteCode, err error) {
	proposerAct := false
	for _, stx := range txs {
		btx, err := app.parseTx(stx, false)
		if err != nil {
			app.logger.Error("unsupported tx, parse fail", "err", err)
			continue
		}
		switch btx.Type {
		case tx.HACTxTypeGrant:
			stx := btx.Tx.(*tx.GrantTx)
			if proposerAct == true {
				return 0, ErrMultiProposalInOneBlock
			}
			proposerAct = true
			if len(stx.Grants) != 1 {
				return 0, ErrOnlySupportOneGrant
			}
			proposerAct, err := st.GetAccount(btx.Validator)
			if proposerAct == nil {
				return 0, errors.New("proposer not found")
			}
			pass, err := app.agentCli.IfGrantNewMember(ctx, st.Header().AccountIdx, proposerAct.Address(), stx.Grants[0].Amount, stx.Grants[0].Statement)
			if err != nil {
				return 0, err
			}
			if pass {
				code = tx.VoteGrantNewMember
			} else {
				code = tx.VoteRejectNewMember
			}
			continue
		case tx.HACTxTypeProposal:
			if proposerAct == true {
				return 0, ErrMultiProposalInOneBlock
			}
			proposerAct = true
			stx := btx.Tx.(*tx.ProposalTx)
			pass, err := app.agentCli.IfProcessProposal(ctx, stx.Proposer, stx.Data)
			if err != nil {
				return 0, err
			}
			if pass {
				code = tx.VoteProcessProposal
			} else {
				code = tx.VoteIgnoreProposal
			}
			continue
		case tx.HACTxTypeSettleProposal:
			if proposerAct == true {
				return 0, ErrMultiProposalInOneBlock
			}
			proposerAct = true
			stx := btx.Tx.(*tx.SettleProposalTx)
			voterAct, err := st.GetAccount(btx.Validator)
			if voterAct == nil {
				return 0, errors.New("voter not found")
			}
			if time.Now().Unix() > int64(stx.ExpireTimestamp) {
				code = tx.VoteRejectProposal
				continue
			}
			pass, err := app.agentCli.IfAcceptProposal(ctx, stx.Proposal, voterAct.Address())
			if err != nil {
				return 0, err
			}
			if pass {
				code = tx.VoteAcceptProposal
			} else {
				code = tx.VoteRejectProposal
			}
			continue
		}
	}
	return
}
