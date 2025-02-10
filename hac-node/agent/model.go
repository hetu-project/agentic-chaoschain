package agent

// sqlite models

type Height struct {
	Id     uint64 `gorm:"primaryKey" json:"id"`
	Height uint64 `json:"height"`
}

type ValidatorAgent struct {
	Id        uint64 `gorm:"primaryKey" json:"id"`
	Address   string `json:"address"`
	Stake     uint64 `json:"stake"`
	AgentUrl  string `json:"agentUrl"`
	Name      string `json:"name"`
	SelfIntro string `json:"self_intro"`
}

type Proposal struct {
	Id              uint64 `gorm:"primaryKey" json:"id"`
	ProposerIndex   uint64 `json:"proposer_index"`
	ProposerAddress string `json:"proposer_address"`
	ProposerName    string `json:"proposer_name"`
	ProposerImage   string `json:"proposer_image"`
	Data            string `json:"data"`
	NewHeight       uint64 `json:"new_height"`
	SettleHeight    uint64 `json:"settle_height"`
	Status          uint64 `json:"status"`
	CreateTimestamp int64  `json:"create_timestamp"`
	ExpireTimestamp int64  `json:"expire_timestamp"`
}

type Grant struct {
	Id              uint64 `gorm:"primaryKey" json:"id"`
	Address         string `json:"address"`
	Height          uint64 `json:"height"`
	Stake           uint64 `json:"stake"`
	Proposer        uint64 `json:"proposer"`
	ProposerAddress string `json:"proposer_address"`
	Grant           bool   `json:"grant"`
}

type ProposalVote struct {
	Id           uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	Proposal     uint64 `json:"proposal"`
	VoterIndex   uint64 `json:"voter_index"`
	VoterAddress string `json:"voter_address"`
	Height       uint64 `json:"height"`
	Vote         uint64 `json:"vote"`
}

type GrantVote struct {
	Id              uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	ProposerIndex   uint64 `json:"proposer_index"`
	ProposerAddress string `json:"proposer_address"`
	AccountIndex    uint64 `json:"account_index"`
	AccountAddr     string `json:"account_addr"`
	VoterIndex      uint64 `json:"voter_index"`
	VoterAddress    string `json:"voter_address"`
	Height          uint64 `json:"height"`
	Vote            uint64 `json:"vote"`
}

type Discussion struct {
	Id             uint64 `gorm:"primaryKey" json:"id"`
	Proposal       uint64 `json:"proposal"`
	SpeakerIndex   uint64 `json:"speaker_index"`
	SpeakerAddress string `json:"speaker_address"`
	SpeakerName    string `json:"speaker_name"`
	Data           string `json:"data"`
	Height         uint64 `json:"height"`
}
