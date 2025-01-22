package state

import (
	"encoding/json"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"google.golang.org/protobuf/proto"
)

type accountSt struct {
	Index    uint64         `json:"index"`
	PubKey   ed25519.PubKey `json:"pubKey"`
	Stake    uint64         `json:"stake"`
	AgentUrl string         `json:"agentUrl"`
	Nonce    uint64         `json:"nonce"`
}

func (a *Account) MarshalJSON() (dat []byte, err error) {
	o := accountSt{
		Index:  a.Index,
		PubKey: a.PubKey,
		Stake:  a.Stake,
		Nonce:  a.Nonce,
	}
	return json.Marshal(o)
}

func (a *Account) UnmarshalJSON(dat []byte) (err error) {
	var o accountSt
	err = json.Unmarshal(dat, &o)
	if err != nil {
		return
	}
	a.Index = o.Index
	a.PubKey = o.PubKey
	a.Stake = o.Stake
	a.AgentUrl = o.AgentUrl
	a.Nonce = o.Nonce
	return
}

func (a *Account) Clone() *Account {
	n := proto.Clone(a)
	return n.(*Account)
}

func (a *Account) SetPubKey(pkey []byte) {
	if a.PubKey == nil {
		a.PubKey = make([]byte, len(pkey))
	}
	copy(a.PubKey, pkey)
}

func (a *Account) AddrBytes() []byte {
	pk := ed25519.PubKey(a.PubKey[:])
	return pk.Address()[:]
}

func (a *Account) Address() string {
	pk := ed25519.PubKey(a.PubKey[:])
	return pk.Address().String()
}

func (a *Account) Verify(msg []byte, sigs [][]byte) (succ bool) {
	if len(sigs) != 1 {
		return false
	}
	pk := ed25519.PubKey(a.PubKey[:])
	return pk.VerifySignature(msg, sigs[0])
}
