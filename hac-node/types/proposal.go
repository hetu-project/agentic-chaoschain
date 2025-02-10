package types

type Proposal struct {
	Index           uint64         `json:"index"`
	Proposer        uint64         `json:"proposer"`
	ProposerAddress string         `json:"proposer_address"`
	Data            []byte         `json:"data"`
	Height          uint64         `json:"height"`
	Status          ProposalStatus `json:"status"`
	EndHeight       uint64         `json:"end_height"`
	ImageUrl        string         `json:"image_url"`
	Title           string         `json:"title"`
	Link            string         `json:"link"`
}

type Discussion struct {
	Index          uint64 `json:"index"`
	Proposal       uint64 `json:"proposal"`
	Speaker        uint64 `json:"speaker"`
	SpeakerAddress string `json:"speaker_address"`
	Data           []byte `json:"data"`
	Height         uint64 `json:"height"`
}

type ProposalStatus uint64

const (
	ProposalStatusIgnore     ProposalStatus = 1
	ProposalStatusProcessing ProposalStatus = 2
	ProposalStatusAccepted   ProposalStatus = 3
	ProposalStatusRejected   ProposalStatus = 4
)
