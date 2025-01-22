package types

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	abci "github.com/cometbft/cometbft/abci/types"
)

const (
	EventUnStakeType         = "retract"
	EventGrantType           = "grant"
	EventUpdateValidatorType = "update_validator"
	EventProposalType        = "proposal"
	EventSettleProposalType  = "settle_proposal"
	EventDiscussionType      = "discussion"
)

type EventUnStake struct {
	Validator uint64 `json:"validatorIndex"`
	Address   string `json:"address"`
	Amount    uint64 `json:"amount"`
}

type EventGrant struct {
	Validator       uint64 `json:"validatorIndex"`
	Address         string `json:"address"`
	Amount          uint64 `json:"amount"`
	AgentUrl        string `json:"agentUrl"`
	Nonce           uint64 `json:"nonce"`
	Grant           bool   `json:"grant"`
	ProposerIndex   uint64 `json:"proposerIndex"`
	ProposerAddress string `json:"proposerAddress"`
}

type EventUpdateValiators struct {
	Updates []abci.ValidatorUpdate `json:"updates"`
}

type EventProposal struct {
	ProposalIndex   uint64 `json:"proposalIndex"`
	Proposer        uint64 `json:"proposerIndex"`
	ProposerAddress string `json:"proposerAddress"`
	EndHeight       uint64 `json:"endHeight"`
	Status          uint64 `json:"status"`
	Data            []byte `json:"data"`
}

func EncodeEventProposal(event *EventProposal) abci.Event {
	return abci.Event{
		Type: EventProposalType,
		Attributes: []abci.EventAttribute{
			{Key: "proposal", Value: fmt.Sprintf("%v", event.ProposalIndex), Index: true},
			{Key: "proposer", Value: fmt.Sprintf("%v", event.Proposer), Index: true},
			{Key: "endHeight", Value: fmt.Sprintf("%v", event.EndHeight), Index: false},
			{Key: "status", Value: fmt.Sprintf("%v", event.Status), Index: false},
			{Key: "data", Value: string(event.Data), Index: false},
			{Key: "proposerAddress", Value: event.ProposerAddress, Index: false},
		},
	}
}

func DecodeEventProposal(originEvent abci.Event) *EventProposal {
	event := &EventProposal{}
	for _, v := range originEvent.Attributes {
		switch v.Key {
		case "proposal":
			proposal, err := strconv.ParseUint(v.Value, 10, 64)
			if err != nil {
				return nil
			}
			event.ProposalIndex = proposal
		case "proposer":
			proposer, err := strconv.ParseUint(v.Value, 10, 64)
			if err != nil {
				return nil
			}
			event.Proposer = proposer
		case "endHeight":
			endHeight, err := strconv.ParseUint(v.Value, 10, 64)
			if err != nil {
				return nil
			}
			event.EndHeight = endHeight
		case "status":
			status, err := strconv.ParseUint(v.Value, 10, 64)
			if err != nil {
				return nil
			}
			event.Status = status
		case "data":
			event.Data = []byte(v.Value)
		case "proposerAddress":
			event.ProposerAddress = fmt.Sprintf("%v", v.Value)
		}
	}
	return event
}

type EventSettleProposal struct {
	Proposer uint64 `json:"proposerIndex"`
	Proposal uint64 `json:"proposal"`
	State    int64  `json:"state"`
}

func EncodeEventSettleProposal(event *EventSettleProposal) abci.Event {
	return abci.Event{
		Type: EventSettleProposalType,
		Attributes: []abci.EventAttribute{
			{Key: "proposer", Value: fmt.Sprintf("%v", event.Proposer), Index: true},
			{Key: "proposal", Value: fmt.Sprintf("%v", event.Proposal), Index: true},
			{Key: "state", Value: fmt.Sprintf("%v", event.State), Index: false},
		},
	}
}

func DecodeEventSettleProposal(originEvent abci.Event) *EventSettleProposal {
	event := &EventSettleProposal{}
	for _, v := range originEvent.Attributes {
		switch v.Key {
		case "proposer":
			proposer, err := strconv.ParseUint(v.Value, 10, 64)
			if err != nil {
				return nil
			}
			event.Proposer = proposer
		case "proposal":
			proposal, err := strconv.ParseUint(v.Value, 10, 64)
			if err != nil {
				return nil
			}
			event.Proposal = proposal
		case "state":
			state, err := strconv.ParseInt(v.Value, 10, 64)
			if err != nil {
				return nil
			}
			event.State = state
		}
	}
	return event
}

type EventDiscussion struct {
	Speaker        uint64 `json:"speakerIndex"`
	SpeakerAddress string `json:"address"`
	Proposal       uint64 `json:"proposal"`
	Data           []byte `json:"data"`
}

func EncodeEventDiscussion(event *EventDiscussion) abci.Event {
	return abci.Event{
		Type: EventDiscussionType,
		Attributes: []abci.EventAttribute{
			{Key: "speaker", Value: fmt.Sprintf("%v", event.Speaker), Index: true},
			{Key: "address", Value: event.SpeakerAddress, Index: false},
			{Key: "proposal", Value: fmt.Sprintf("%v", event.Proposal), Index: true},
			{Key: "data", Value: string(event.Data), Index: false},
		},
	}
}

func DecodeEventDiscussion(originEvent abci.Event) *EventDiscussion {
	event := &EventDiscussion{}
	for _, v := range originEvent.Attributes {
		switch v.Key {
		case "speaker":
			speaker, err := strconv.ParseUint(v.Value, 10, 64)
			if err != nil {
				return nil
			}
			event.Speaker = speaker
		case "address":
			event.SpeakerAddress = fmt.Sprintf("%v", v.Value)
		case "proposal":
			proposal, err := strconv.ParseUint(v.Value, 10, 64)
			if err != nil {
				return nil
			}
			event.Proposal = proposal
		case "data":
			event.Data = []byte(v.Value)
		}
	}
	return event
}

func EncodeEventUpdateValiators(event *EventUpdateValiators) abci.Event {
	pks := make([]string, len(event.Updates))
	powers := make([]string, len(event.Updates))
	for i := range event.Updates {
		ed25519PK := event.Updates[i].PubKey.GetEd25519()
		pks[i] = hex.EncodeToString(ed25519PK)
		powers[i] = fmt.Sprintf("%v", event.Updates[i].Power)
	}
	return abci.Event{
		Type: EventUpdateValidatorType,
		Attributes: []abci.EventAttribute{
			{Key: "pks", Value: strings.Join(pks, ","), Index: false},
			{Key: "powers", Value: strings.Join(powers, ","), Index: false},
		},
	}
}

func ParseEventUpdateValiators(originEvent abci.Event) *EventUpdateValiators {
	event := &EventUpdateValiators{
		Updates: []abci.ValidatorUpdate{},
	}
	pks := make([]string, 0)
	powers := make([]uint64, 0)
	for _, v := range originEvent.Attributes {
		switch v.Key {
		case "pks":
			pks = strings.Split(v.Value, ",")
		case "powers":
			powerStrs := strings.Split(v.Value, ",")
			for _, powerStr := range powerStrs {
				power, err := strconv.ParseUint(powerStr, 10, 64)
				if err != nil {
					return nil
				}
				powers = append(powers, power)
			}
		}
	}
	if len(pks) != len(powers) {
		return nil
	}
	for i := range pks {
		pk, err := hex.DecodeString(pks[i])
		if err != nil {
			return nil
		}
		event.Updates = append(event.Updates, abci.Ed25519ValidatorUpdate(pk, int64(powers[i])))
	}
	return event
}

func EncodeEventGrant(event *EventGrant) abci.Event {

	return abci.Event{
		Type: EventGrantType,
		Attributes: []abci.EventAttribute{
			{Key: "validator", Value: fmt.Sprintf("%v", event.Validator), Index: true},
			{Key: "addr", Value: event.Address, Index: false},
			{Key: "amount", Value: fmt.Sprintf("%v", event.Amount), Index: false},
			{Key: "nonce", Value: fmt.Sprintf("%v", event.Nonce), Index: false},
			{Key: "grant", Value: fmt.Sprintf("%v", event.Grant), Index: false},
			{Key: "proposer", Value: fmt.Sprintf("%v", event.ProposerIndex), Index: false},
			{Key: "proposerAddress", Value: event.ProposerAddress, Index: false},
			{Key: "agentUrl", Value: event.AgentUrl, Index: false},
		},
	}
}

func ParseEventGrant(originEvent abci.Event) *EventGrant {
	event := &EventGrant{}
	for _, v := range originEvent.Attributes {
		switch v.Key {
		case "validator":
			validator, err := strconv.ParseUint(v.Value, 10, 64)
			if err != nil {
				return nil
			}
			event.Validator = validator
		case "addr":
			event.Address = fmt.Sprintf("%v", v.Value)
		case "amount":
			amount, err := strconv.ParseUint(v.Value, 10, 64)
			if err != nil {
				return nil
			}
			event.Amount = amount
		case "nonce":
			nonce, err := strconv.ParseUint(v.Value, 10, 64)
			if err != nil {
				return nil
			}
			event.Nonce = nonce
		case "grant":
			grant, err := strconv.ParseBool(v.Value)
			if err != nil {
				return nil
			}
			event.Grant = grant
		case "proposer":
			proposer, err := strconv.ParseUint(v.Value, 10, 64)
			if err != nil {
				return nil
			}
			event.ProposerIndex = proposer
		case "proposerAddress":
			event.ProposerAddress = v.Value
		case "agentUrl":
			event.AgentUrl = v.Value
		}
	}
	return event
}

func ParseEventUnStake(originEvent abci.Event) *EventUnStake {
	event := &EventUnStake{}
	for _, v := range originEvent.Attributes {
		switch v.Key {
		case "validator":
			validator, err := strconv.ParseUint(v.Value, 10, 64)
			if err != nil {
				return nil
			}
			event.Validator = validator
		case "amount":
			amount, err := strconv.ParseUint(v.Value, 10, 64)
			if err != nil {
				return nil
			}
			event.Amount = amount
		case "addr":
			event.Address = fmt.Sprintf("%v", v.Value)
		}
	}
	return event
}
