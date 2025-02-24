package app

import (
	"context"
	"encoding/json"
	"strings"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	cmtlog "github.com/cometbft/cometbft/libs/log"
	"github.com/hetu-project/hetu-chaoschain/state"
)

func (app *HACApp) Query(ctx context.Context, req *abcitypes.RequestQuery) (res *abcitypes.ResponseQuery, err error) {
	path := req.Path
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	q, ok := app.queriers[path]
	if !ok {
		res = &abcitypes.ResponseQuery{}
		res.Code = 404
		return
	}
	res, err = q.Query(ctx, req)
	return
}

type AccountQuerier struct {
	db     *state.StateDB
	logger cmtlog.Logger
}

func NewAccountQuerier(db *state.StateDB, logger cmtlog.Logger) (q *AccountQuerier) {
	q = &AccountQuerier{
		db:     db,
		logger: logger,
	}
	return
}

func (q *AccountQuerier) Query(ctx context.Context, req *abcitypes.RequestQuery) (res *abcitypes.ResponseQuery, err error) {
	res = &abcitypes.ResponseQuery{}
	var a *state.Account
	var height uint64
	if len(req.Data) == 20 {
		a, height, _ = q.db.GetAccountByAddress(req.Data)
	} else if len(req.Data) <= 8 {
		var idx uint64
		for _, v := range req.Data {
			idx <<= 8
			idx |= uint64(v)
		}
		a, height, _ = q.db.GetAccountByIndex(idx)
	}
	if a != nil {
		res.Value, _ = a.MarshalJSON()
		res.Height = int64(height)
	} else {
		res.Code = 1
	}
	return
}

type Querier interface {
	Query(ctx context.Context, req *abcitypes.RequestQuery) (res *abcitypes.ResponseQuery, err error)
}

type ValidatorQuerier struct {
	db     *state.StateDB
	logger cmtlog.Logger
}

func NewValidatorQuerier(db *state.StateDB, logger cmtlog.Logger) (q *ValidatorQuerier) {
	q = &ValidatorQuerier{
		db:     db,
		logger: logger,
	}
	return
}

func (q *ValidatorQuerier) Query(ctx context.Context, req *abcitypes.RequestQuery) (res *abcitypes.ResponseQuery, err error) {
	res = &abcitypes.ResponseQuery{}
	validators, height, err := q.db.State().ValidatorAccounts()
	if err != nil {
		res.Code = 1
		return
	}
	res.Height = int64(height)
	res.Value, _ = json.Marshal(validators)
	return
}
