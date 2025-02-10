package tx

import (
	"bytes"
	"encoding/json"
)

type HACTx struct {
	Version   uint8     `json:"version"`
	Type      HACTxType `json:"type"`
	Nonce     uint64    `json:"nonce"`
	Validator uint64    `json:"validator"`
	Tx        any       `json:"tx"`
	Sig       [][]byte  `json:"sig"`
}

type GrantTx struct {
	Grants []GrantSt `json:"grants"`
}

type GrantSt struct {
	Statement string `json:"statement"`
	Amount    uint64 `json:"amount"`
	AgentUrl  string `json:"agentUrl"`
	Name      string `json:"name"`
	Pubkey    []byte `json:"pubkey"`
}

func (d *GrantSt) Equal(grant GrantSt) bool {
	if d.Amount != grant.Amount || !bytes.Equal(d.Pubkey, grant.Pubkey) {
		return false
	}
	return true
}

type DiscussionTx struct {
	Proposal uint64 `json:"proposal"`
	Data     []byte `json:"data"`
}

type ProposalTx struct {
	Proposer  uint64 `json:"proposer"`
	EndHeight uint64 `json:"endHeight"`
	ImageUrl  string `json:"imageUrl"`
	Title     string `json:"title"`
	Link      string `json:"link"`
	Data      []byte `json:"data"`
}

type SettleProposalTx struct {
	Proposal uint64 `json:"proposal"`
}

type RetractTx struct {
	Amount uint64 `json:"amount"`
}

type hacTxTmpl[Tx any] struct {
	Version   uint8     `json:"version"`
	Type      HACTxType `json:"type"`
	Nonce     uint64    `json:"nonce"`
	Validator uint64    `json:"validator"`
	Tx        Tx        `json:"tx"`
	Sig       [][]byte  `json:"sig"`
}

func (tx *HACTx) SigData(ext []byte) (dat []byte, err error) {
	ntx := *tx
	ntx.Sig = [][]byte{ext}
	dat, err = json.Marshal(ntx)
	return
}

func parseHACTxType(dat []byte) HACTxType {
	var tx struct {
		Type HACTxType `json:"type"`
	}
	err := json.Unmarshal(dat, &tx)
	if err != nil {
		return HACTxTypeUnknown
	}
	return tx.Type
}

func unmarshalHACTx[Tx any](dat []byte) (btx *HACTx, err error) {
	var txt hacTxTmpl[Tx]
	err = json.Unmarshal(dat, &txt)
	if err != nil {
		return
	}
	btx = new(HACTx)
	btx.Version = txt.Version
	btx.Type = txt.Type
	btx.Nonce = txt.Nonce
	btx.Validator = txt.Validator
	btx.Tx = &txt.Tx
	btx.Sig = txt.Sig
	return
}

func UnmarshalHACTx(dat []byte) (btx *HACTx, err error) {
	tp := parseHACTxType(dat)
	switch tp {
	case HACTxTypeProposal:
		return unmarshalHACTx[ProposalTx](dat)
	case HACTxTypeDiscussion:
		return unmarshalHACTx[DiscussionTx](dat)
	case HACTxTypeGrant:
		return unmarshalHACTx[GrantTx](dat)
	case HACTxTypeRetract:
		return unmarshalHACTx[RetractTx](dat)
	case HACTxTypeSettleProposal:
		return unmarshalHACTx[SettleProposalTx](dat)
	default:
		err = ErrUnsupportedTxType
	}
	return
}

func MarshalHACTx(btx *HACTx) (dat []byte, err error) {
	return json.Marshal(btx)
}
