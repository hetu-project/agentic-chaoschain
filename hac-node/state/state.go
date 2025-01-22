package state

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sort"

	"container/heap"

	"github.com/calehh/hac-app/config"
	"github.com/calehh/hac-app/tx"
	txtypes "github.com/calehh/hac-app/tx"
	hac_types "github.com/calehh/hac-app/types"
	abci_types "github.com/cometbft/cometbft/abci/types"
	cmtcrypto "github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	cmtlog "github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/iavl"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/syndtr/goleveldb/leveldb"
	"google.golang.org/protobuf/proto"
)

const (
	StartAccountIdx = 65536

	ModifiedFlagNew = 1 << 0
	ModifiedFlagMod = 1 << 1
	ModifiedFlagPK  = 1 << 2

	MaxValidators = 100
)

var (
	ErrNotFound = errors.New("not found")
)

var (
	KeyState                 = "s"
	KeyAccountIndex          = "i%s"
	KeyAccountBody           = "a%x"
	KeyHash                  = "h%x"
	KeyGrant                 = []byte("k")
	KeyStakesReleaseHeight   = "stake%x"
	KeyRetractsReleaseHeight = "retract%x"
	KeyProposalBody          = "p%v"
	KeyProposalIndex         = "pi"
	KeyDiscussionBody        = "d%v"
	KeyDiscussionIndex       = "di"
	KeyManifest              = "m"
)

var (
	ErrTxValidatorNoexists          = errors.New("validator noexists")
	ErrTxNotMembership              = errors.New("not membership")
	ErrTxNonceInvalid               = errors.New("nonce invalid")
	ErrTxSigInvalid                 = errors.New("signature invalid")
	ErrStateHeightUnmatched         = errors.New("state height unmatched")
	ErrProposerAddressNoexists      = errors.New("proposer address noexists")
	ErrAccountAlreadyExists         = errors.New("account already exists")
	ErrAccountNoexists              = errors.New("account noexists")
	ErrProposaNoexists              = errors.New("proposal noexists")
	ErrUnexpectedWithdrawal         = errors.New("unexpected withdrawal")
	ErrPayloadProposalHeigtMismatch = errors.New("payload proposal height mismatch")
	ErrTxProposalNoexists           = errors.New("proposal noexists")
	ErrTxMoreThanOneProposal        = errors.New("more than one proposal")
	ErrTxVoteCodeInvalid            = errors.New("vote code invalid")
	ErrOneActionInOneBlock          = errors.New("one action in one block")
)

type State struct {
	logger cmtlog.Logger
	db     *iavl.MutableTree
	dbVer  int64

	header     *StateHeader
	validators []abci_types.ValidatorUpdate
	idxs       map[string]uint64
	acnts      map[uint64]*Account

	modifiedAcnts      map[uint64]uint32
	proposalMaxIndex   uint64
	discussionMaxIndex uint64
	modProposal        *hac_types.Proposal
	newDiscussions     map[uint64]hac_types.Discussion
}

func newState(db *iavl.MutableTree, logger cmtlog.Logger) *State {
	s := &State{
		logger:             logger,
		db:                 db,
		dbVer:              0,
		header:             new(StateHeader),
		validators:         []abci_types.ValidatorUpdate{},
		idxs:               make(map[string]uint64),
		acnts:              make(map[uint64]*Account),
		modifiedAcnts:      make(map[uint64]uint32),
		proposalMaxIndex:   0,
		discussionMaxIndex: 0,
		modProposal:        nil,
		newDiscussions:     map[uint64]hac_types.Discussion{},
	}
	s.header.AccountIdx = StartAccountIdx
	return s
}

func (s *State) nextState() *State {
	n := &State{
		logger:             s.logger,
		db:                 s.db,
		dbVer:              s.dbVer,
		idxs:               make(map[string]uint64),
		acnts:              make(map[uint64]*Account),
		modifiedAcnts:      make(map[uint64]uint32),
		proposalMaxIndex:   s.proposalMaxIndex,
		discussionMaxIndex: s.discussionMaxIndex,
		newDiscussions:     make(map[uint64]hac_types.Discussion),
	}
	n.header = proto.Clone(s.header).(*StateHeader)
	if s.header.GetHash() != nil {
		n.header.Height = s.header.Height + 1
	}

	return n
}

func deepCopyMap[K comparable, V any](source map[K]V) map[K]V {
	res := make(map[K]V)
	for k, v := range source {
		switch x := any(v).(type) {
		case *Account:
			account := proto.Clone(x).(*Account)
			res[k] = any(account).(V)
		default:
			res[k] = v
		}
	}
	return res
}

func deepCopySlice[E any](source []E) []E {
	res := make([]E, len(source))
	if len(source) == 0 {
		return res
	}
	for idx, ele := range source {
		switch e := any(ele).(type) {
		case abci_types.ValidatorUpdate:
			b, _ := e.Marshal()
			eleClone := abci_types.ValidatorUpdate{}
			eleClone.Unmarshal(b)
			res[idx] = any(eleClone).(E)
		default:
			copy(res, source)
			return res
		}
	}
	return res
}

func (s *State) Clone() *State {
	n := &State{
		logger:             s.logger,
		db:                 s.db,
		dbVer:              s.dbVer,
		header:             &StateHeader{},
		validators:         deepCopySlice(s.validators),
		idxs:               deepCopyMap(s.idxs),
		acnts:              deepCopyMap(s.acnts),
		modifiedAcnts:      deepCopyMap(s.modifiedAcnts),
		proposalMaxIndex:   s.proposalMaxIndex,
		discussionMaxIndex: s.discussionMaxIndex,
		modProposal:        s.modProposal,
		newDiscussions:     deepCopyMap(s.newDiscussions),
	}
	n.header = proto.Clone(s.header).(*StateHeader)
	if s.header.GetHash() != nil {
		n.header.Height = s.header.Height + 1
	}
	return n
}

func (s *State) load() (err error) {
	val, err := s.db.Get([]byte(KeyProposalIndex))
	if err != nil {
		if err != leveldb.ErrNotFound {
			return err
		}
	}
	s.proposalMaxIndex = new(big.Int).SetBytes(val).Uint64()
	val, err = s.db.Get([]byte(KeyDiscussionIndex))
	if err != nil {
		if err != leveldb.ErrNotFound {
			return err
		}
	}
	s.discussionMaxIndex = new(big.Int).SetBytes(val).Uint64()
	val, err = s.db.Get([]byte(KeyState))
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil
		}
		return err
	}
	if val != nil {
		err = proto.Unmarshal(val, s.header)
		if err != nil {
			return
		}
		h := s.db.Hash()
		if h != nil {
			s.calcHash(h, true)
		}
	}
	return
}

func (s *State) calcHash(rootHash []byte, update bool) (h common.Hash) {
	h = crypto.Keccak256Hash(rootHash)
	if update {
		if s.header.RootHash == nil {
			s.header.RootHash = make([]byte, len(rootHash))
		}
		copy(s.header.RootHash, rootHash)
		if s.header.Hash == nil {
			s.header.Hash = make([]byte, len(h))
		}
		copy(s.header.Hash, h[:])
	}
	return
}

func (s *State) Update() (h common.Hash, err error) {
	var hash []byte
	defer func() {
		if hash == nil {
			s.db.Rollback()
		}
	}()
	var val []byte
	val, err = proto.Marshal(s.header)
	if err != nil {
		return
	}
	_, err = s.db.Set([]byte(KeyState), val)
	if err != nil {
		return
	}

	if len(s.newDiscussions) != 0 {
		_, err = s.db.Set([]byte(KeyDiscussionIndex), big.NewInt(int64(s.discussionMaxIndex)).Bytes())
		if err != nil {
			return
		}
		//todo: record discussion in state
	}

	if s.modProposal != nil {
		_, err = s.db.Set([]byte(KeyProposalIndex), big.NewInt(int64(s.proposalMaxIndex)).Bytes())
		if err != nil {
			return
		}
		key := fmt.Sprintf(KeyProposalBody, s.modProposal.Index)
		proposalBz, _ := json.Marshal(s.modProposal)
		_, err = s.db.Set([]byte(key), proposalBz)
		if err != nil {
			return
		}
	}

	n := len(s.modifiedAcnts)
	if n > 0 {
		idxs := make([]uint64, n)
		i := 0
		for idx := range s.modifiedAcnts {
			idxs[i] = idx
			i += 1
		}
		sort.Slice(idxs, func(i, j int) bool {
			return idxs[i] < idxs[j]
		})
		for _, idx := range idxs {
			flag := s.modifiedAcnts[idx]
			acnt := s.acnts[idx]
			key := fmt.Sprintf(KeyAccountBody, acnt.Index)
			val, err = proto.Marshal(acnt)
			if err != nil {
				return
			}
			_, err = s.db.Set([]byte(key), val)
			if err != nil {
				return
			}
			if (flag&ModifiedFlagNew == ModifiedFlagNew) || (flag&ModifiedFlagPK == ModifiedFlagPK) {
				key = fmt.Sprintf(KeyAccountIndex, acnt.Address())
				val, err = rlp.EncodeToBytes(acnt.Index)
				if err != nil {
					return
				}
				_, err = s.db.Set([]byte(key), val)
				if err != nil {
					return
				}
			}
		}
	}
	hash = s.db.WorkingHash()
	h = s.calcHash(hash, false)
	s.modifiedAcnts = make(map[uint64]uint32)
	return
}

func (s *State) save() (h common.Hash, err error) {
	hash, ver, err := s.db.SaveVersion()
	if err != nil {
		return h, err
	}

	s.dbVer = ver
	h = s.calcHash(hash, true)

	return
}

func (s *State) getProposalMax() uint64 {
	return s.proposalMaxIndex
}

func (s *State) getProposalByIndex(index uint64) (*hac_types.Proposal, error) {
	key := fmt.Sprintf(KeyProposalBody, index)
	val, err := s.db.Get([]byte(key))
	if err != nil {
		return nil, err
	}
	var proposal hac_types.Proposal
	err = json.Unmarshal(val, &proposal)
	if err != nil {
		return nil, err
	}
	return &proposal, nil
}

func (s *State) getDiscussionMax() uint64 {
	return s.discussionMaxIndex
}

func (s *State) getDiscussionByIndex(index uint64) (*hac_types.Discussion, error) {
	key := fmt.Sprintf(KeyDiscussionBody, index)
	val, err := s.db.Get([]byte(key))
	if err != nil {
		return nil, err
	}
	var dis hac_types.Discussion
	err = json.Unmarshal(val, &dis)
	if err != nil {
		return nil, err
	}
	return &dis, nil
}

func (s *State) getProposal(idx uint64) (proposal *hac_types.Proposal, err error) {
	if idx > s.proposalMaxIndex {
		err = ErrProposaNoexists
		return
	}
	key := fmt.Sprintf(KeyProposalBody, idx)
	val, err := s.db.Get([]byte(key))
	if err != nil {
		return nil, err
	}
	if val == nil {
		err = ErrNotFound
		return
	}
	proposal = new(hac_types.Proposal)
	err = json.Unmarshal(val, proposal)
	return
}

func (s *State) GetAccount(idx uint64) (acnt *Account, err error) {
	if idx >= s.header.AccountIdx {
		err = ErrAccountNoexists
		return
	}
	acnt = s.acnts[idx]
	if acnt != nil {
		return
	}
	key := fmt.Sprintf(KeyAccountBody, idx)
	val, err := s.db.Get([]byte(key))
	if err != nil {
		return nil, err
	}
	if val == nil {
		err = ErrNotFound
		return
	}
	acnt = new(Account)
	err = proto.Unmarshal(val, acnt)
	if err != nil {
		acnt = nil
	}
	s.acnts[idx] = acnt
	return
}

func (s *State) existPubkey(pubkey []byte) (bool, error) {
	addr := ed25519.PubKey(pubkey).Address()[:]
	saddr := cmtcrypto.Address(addr).String()
	// exist in cache
	if _, ok := s.idxs[saddr]; ok {
		return true, nil
	}
	// exist in db
	key := fmt.Sprintf(KeyAccountIndex, saddr)
	val, err := s.db.Get([]byte(key))
	if err != nil {
		if err != leveldb.ErrNotFound {
			return false, err
		}
	}
	if val != nil {
		return true, nil
	}
	//exist in modify
	for _, acc := range s.acnts {
		if bytes.Equal(acc.AddrBytes(), addr) {
			return true, nil
		}
	}
	return false, nil
}

func (s *State) FindAccount(addr []byte) (acnt *Account, err error) {
	saddr := cmtcrypto.Address(addr).String()
	idx, ok := s.idxs[saddr]
	if !ok {
		key := fmt.Sprintf(KeyAccountIndex, saddr)
		val, err := s.db.Get([]byte(key))
		if err != nil {
			if err == leveldb.ErrNotFound {
				return nil, nil
			}
			return nil, err
		}
		if val == nil {
			return nil, nil
		}
		err = rlp.DecodeBytes(val, &idx)
		if err != nil {
			return nil, err
		}
		s.idxs[saddr] = idx
	}
	acnt, err = s.GetAccount(idx)

	return
}

func (s *State) SetManifest(manifest string) error {
	_, err := s.db.Set([]byte(KeyManifest), []byte(manifest))
	return err
}

func (s *State) GetManifest() (manifest string, err error) {
	val, err := s.db.Get([]byte(KeyManifest))
	if err != nil {
		if err != leveldb.ErrNotFound {
			return "", err
		}
	}
	manifest = string(val)
	return
}

func (s *State) ValidatorAccounts() (acounts []*Account, height uint64, err error) {
	vals := s.validators
	for _, val := range vals {
		pk := ed25519.PubKey(val.PubKey.GetEd25519()[:])
		addr := pk.Address()[:]
		act, _ := s.FindAccount(addr)
		if act != nil {
			acounts = append(acounts, act)
		}
	}
	height = s.header.Height
	return
}

func (s *State) Header() *StateHeader {
	return s.header
}

func (s *State) Hash() (h common.Hash) {
	if s.header.Hash != nil {
		copy(h[:], s.header.Hash)
	}
	return
}

func (s *State) SetChainId(chainId string) {
	s.header.ChainId = chainId
}

func (s *State) AddAccount(acnt *Account) (err error) {
	a, err := s.FindAccount(acnt.AddrBytes())
	if err != nil {
		return err
	}
	if a != nil {
		err = ErrAccountAlreadyExists
		return
	}
	acnt.Index = s.header.AccountIdx
	s.header.AccountIdx += 1
	s.acnts[acnt.Index] = acnt.Clone()
	s.modifiedAcnts[acnt.Index] = ModifiedFlagNew
	return
}

func (s *State) Verify(tx *tx.HACTx, allowNonceGap bool) (succ bool, err error) {
	a, err := s.GetAccount(tx.Validator)
	if err != nil {
		return succ, err
	}
	if a == nil {
		err = ErrTxValidatorNoexists
		return
	}
	if !(a.Nonce == tx.Nonce || (allowNonceGap && a.Nonce < tx.Nonce)) {
		err = ErrTxNonceInvalid
		return
	}
	dat, err := tx.SigData([]byte(s.header.ChainId))
	if err != nil {
		return succ, err
	}
	succ = a.Verify(dat, tx.Sig)
	if !succ {
		err = ErrTxSigInvalid
	}
	return
}

func (s *State) Proposal(tx *tx.ProposalTx, validator uint64, checkOnly bool, code tx.VoteCode) (event *hac_types.EventProposal, err error) {
	if code != txtypes.VoteIgnoreProposal && code != txtypes.VoteProcessProposal {
		return nil, ErrTxVoteCodeInvalid
	}
	s.logger.Debug("apply proposal", "validator", validator, "height", s.header.Height)
	a, err := s.GetAccount(validator)
	if err != nil {
		return nil, err
	}
	if a == nil {
		err = ErrTxValidatorNoexists
		return
	}
	if a.Stake == 0 {
		err = ErrTxNotMembership
		return
	}
	if s.modProposal != nil && s.modProposal.Index != 0 {
		err = ErrTxMoreThanOneProposal
		return
	}
	if !checkOnly {
		s.proposalMaxIndex += 1
		proposal := hac_types.Proposal{
			Index:           s.proposalMaxIndex,
			Proposer:        a.Index,
			ProposerAddress: a.Address(),
			Data:            tx.Data,
			Height:          s.header.Height,
			EndHeight:       tx.EndHeight,
		}
		if code == txtypes.VoteIgnoreProposal {
			proposal.Status = hac_types.ProposalStatusIgnore
		} else {
			proposal.Status = hac_types.ProposalStatusProcessing
		}
		s.modProposal = &proposal

		a.Nonce += 1
		v := s.modifiedAcnts[a.Index]
		v |= ModifiedFlagMod
		s.modifiedAcnts[a.Index] = v
		s.acnts[a.Index] = a.Clone()

		event = &hac_types.EventProposal{
			ProposalIndex:   proposal.Index,
			ProposerAddress: a.Address(),
			Proposer:        a.Index,
			EndHeight:       tx.EndHeight,
			Status:          uint64(proposal.Status),
			Data:            proposal.Data,
		}
	}
	return
}

func (s *State) SettleProposal(tx *tx.SettleProposalTx, validator uint64, checkOnly bool, code tx.VoteCode) (event *hac_types.EventSettleProposal, err error) {
	if code != txtypes.VoteAcceptProposal && code != txtypes.VoteRejectProposal {
		return nil, ErrTxVoteCodeInvalid
	}
	s.logger.Debug("apply settle proposal", "validator", validator, "height", s.header.Height)
	if s.modProposal != nil && s.modProposal.Index != 0 {
		err = ErrTxMoreThanOneProposal
		return
	}
	a, err := s.GetAccount(validator)
	if err != nil {
		return nil, err
	}
	if a == nil {
		err = ErrTxValidatorNoexists
		return
	}
	proposal, err := s.getProposal(tx.Proposal)
	if err != nil {
		return nil, err
	}
	if proposal.Proposer != validator {
		return nil, fmt.Errorf("proposal not settle by proposer")
	}
	if proposal.Status != hac_types.ProposalStatusProcessing {
		return nil, fmt.Errorf("proposal not processing status is %v", proposal.Status)
	}
	if !checkOnly {
		if code == txtypes.VoteAcceptProposal {
			proposal.Status = hac_types.ProposalStatusAccepted
		} else {
			proposal.Status = hac_types.ProposalStatusRejected
		}
		s.modProposal = proposal

		a.Nonce += 1
		v := s.modifiedAcnts[a.Index]
		v |= ModifiedFlagMod
		s.modifiedAcnts[a.Index] = v
		s.acnts[a.Index] = a.Clone()

		event = &hac_types.EventSettleProposal{
			Proposer: proposal.Proposer,
			Proposal: tx.Proposal,
			State:    int64(proposal.Status),
		}
	}
	return
}

func (s *State) Dicussion(tx *tx.DiscussionTx, validator uint64, checkOnly bool) (event *hac_types.EventDiscussion, err error) {
	s.logger.Debug("apply discussion", "validator", validator, "height", s.header.Height)
	a, err := s.GetAccount(validator)
	if err != nil {
		return nil, err
	}
	if a == nil {
		err = ErrTxValidatorNoexists
		return
	}
	if a.Stake == 0 {
		err = ErrTxNotMembership
		return
	}
	if tx.Proposal > s.getProposalMax() {
		err = ErrTxProposalNoexists
		return
	}

	if !checkOnly {
		s.discussionMaxIndex += 1
		dis := hac_types.Discussion{
			Index:          s.discussionMaxIndex,
			Proposal:       tx.Proposal,
			Speaker:        a.Index,
			SpeakerAddress: a.Address(),
			Data:           tx.Data,
			Height:         s.header.Height,
		}
		s.newDiscussions[s.discussionMaxIndex] = dis

		a.Nonce += 1
		v := s.modifiedAcnts[a.Index]
		v |= ModifiedFlagMod
		s.modifiedAcnts[a.Index] = v
		s.acnts[a.Index] = a.Clone()

		event = &hac_types.EventDiscussion{
			Speaker:        a.Index,
			SpeakerAddress: a.Address(),
			Proposal:       tx.Proposal,
			Data:           dis.Data,
		}
	}
	return
}

func (s *State) Grant(proposer uint64, pk []byte, amount uint64, agentUrl, name string, code tx.VoteCode) (event *hac_types.EventGrant, err error) {
	if code != txtypes.VoteGrantNewMember && code != txtypes.VoteRejectNewMember {
		return nil, ErrTxVoteCodeInvalid
	}
	proposerAcc, err := s.GetAccount(proposer)
	if err != nil {
		return nil, err
	}
	if proposerAcc == nil {
		err = ErrTxValidatorNoexists
		return
	}
	if proposerAcc.Stake == 0 {
		err = ErrTxNotMembership
		return
	}
	addr := ed25519.PubKey(pk).Address()
	a, err := s.FindAccount(addr)
	if err != nil {
		return nil, err
	}
	if a != nil {
		return nil, errors.New("account already exists")
	}
	if code != txtypes.VoteGrantNewMember {
		a = &Account{
			Index:    s.header.AccountIdx,
			PubKey:   pk,
			Stake:    amount,
			AgentUrl: agentUrl,
			Name:     name,
			Nonce:    0,
		}
		event = &hac_types.EventGrant{
			Validator:       s.header.AccountIdx,
			Address:         addr.String(),
			Amount:          amount,
			Nonce:           0,
			Grant:           false,
			AgentUrl:        agentUrl,
			ProposerIndex:   proposer,
			ProposerAddress: proposerAcc.Address(),
		}
	} else {
		a = &Account{
			Index:    s.header.AccountIdx,
			PubKey:   pk,
			Stake:    amount,
			AgentUrl: agentUrl,
			Name:     name,
			Nonce:    0,
		}
		event = &hac_types.EventGrant{
			Validator:       a.Index,
			Address:         a.Address(),
			Amount:          amount,
			Nonce:           a.Nonce,
			Grant:           true,
			AgentUrl:        agentUrl,
			ProposerIndex:   proposer,
			ProposerAddress: proposerAcc.Address(),
		}
	}
	s.header.AccountIdx += 1
	s.modifiedAcnts[a.Index] = ModifiedFlagNew
	s.acnts[a.Index] = a.Clone()
	return
}

func (s *State) UnStake(tx *tx.RetractTx, validator uint64, checkOnly bool) (event *hac_types.EventUnStake, err error) {
	s.logger.Debug("apply retract", "validator", validator, "amount", tx.Amount, "height", s.header.Height)
	a, err := s.GetAccount(validator)
	if err != nil {
		return nil, err
	}
	if a == nil {
		err = ErrTxValidatorNoexists
		return
	}
	if a.Stake == 0 {
		err = ErrTxNotMembership
		return
	}
	if a.Stake != tx.Amount {
		err = fmt.Errorf("must retract all")
		return
	}
	if !checkOnly {
		event = &hac_types.EventUnStake{
			Validator: validator,
			Address:   a.Address(),
			Amount:    tx.Amount,
		}
		a.Stake -= tx.Amount
		a.Nonce += 1
		v := s.modifiedAcnts[a.Index]
		v |= ModifiedFlagMod
		s.modifiedAcnts[a.Index] = v
		s.acnts[a.Index] = a.Clone()
	}
	return
}

func (s *State) Validators() (updateVals map[string]abci_types.ValidatorUpdate, err error) {
	updateVals = make(map[string]abci_types.ValidatorUpdate, 0)
	start := []byte(fmt.Sprintf(KeyAccountBody, ""))
	end := PrefixEndBytes(start)
	aIterator, err := s.db.Iterator(start, end, false)
	if err != nil {
		return nil, err
	}

	valsQueue := &PowerQueue{}
	heap.Init(valsQueue)
	for ; aIterator.Valid(); aIterator.Next() {
		var act Account
		valBytes := aIterator.Value()
		err = proto.Unmarshal(valBytes, &act)
		if err != nil {
			return nil, err
		}
		power := config.PowerPerStake(act.Stake, s.header.Height)
		if power > 0 {
			heap.Push(valsQueue, validatorWithPower{
				Index:  act.Index,
				Pubkey: act.PubKey,
				Power:  power,
			})
		}
	}

	vals := make([]abci_types.ValidatorUpdate, 0)
	for valsQueue.Len() > 0 && len(vals) < MaxValidators {
		val := heap.Pop(valsQueue).(validatorWithPower)
		vals = append(vals, abci_types.Ed25519ValidatorUpdate(val.Pubkey, val.Power))
	}
	s.validators = vals

	for _, val := range vals {
		updateVals[val.PubKey.String()] = val
	}

	return updateVals, nil
}

func (s *State) ValidatorsUpdate(curVals map[string]abci_types.ValidatorUpdate) (updateVals []abci_types.ValidatorUpdate, err error) {
	nextVals, err := s.Validators()
	if err != nil {
		return nil, err
	}

	for key, val := range nextVals {
		if v, ok := curVals[key]; ok {
			if v.Power != val.Power {
				updateVals = append(updateVals, val)
			}
		} else {
			updateVals = append(updateVals, val)
		}
	}

	for key, curVal := range curVals {
		if _, ok := nextVals[key]; !ok {
			curVal.Power = 0
			updateVals = append(updateVals, curVal)
		}
	}
	return
}

type validatorWithPower struct {
	Index  uint64
	Pubkey []byte
	Power  int64
}

type PowerQueue []validatorWithPower

func (pq PowerQueue) Len() int { return len(pq) }

func (pq PowerQueue) Less(i, j int) bool {
	if pq[i].Power == pq[j].Power {
		return pq[i].Index < pq[j].Index
	}
	return pq[i].Power > pq[j].Power
}

func (pq PowerQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PowerQueue) Push(x any) {
	item := x.(validatorWithPower)
	*pq = append(*pq, item)
}

func (pq *PowerQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

func PrefixEndBytes(prefix []byte) []byte {
	if len(prefix) == 0 {
		return nil
	}

	end := make([]byte, len(prefix))
	copy(end, prefix)

	for {
		if end[len(end)-1] != byte(255) {
			end[len(end)-1]++
			break
		}

		end = end[:len(end)-1]

		if len(end) == 0 {
			end = nil
			break
		}
	}

	return end
}
