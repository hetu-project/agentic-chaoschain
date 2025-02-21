package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	cmtlog "github.com/cometbft/cometbft/libs/log"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"io"
	"net/http"
	"net/url"
)

var ClientInstance Client

var DiscussionRate = 0

var DiscussionTrigger = 0

type Client interface {
	IfProcessProposal(ctx context.Context, data []byte) (bool, error)
	IfAcceptProposal(ctx context.Context, proposal uint64, voter string) (bool, error)
	IfGrantNewMember(ctx context.Context, validator uint64, proposer string, amount uint64, statement string) (bool, error)
	CommentPropoal(ctx context.Context, proposal uint64, speaker string) (string, error)
	AddProposal(ctx context.Context, proposal uint64, proposer string, text string) error
	AddDiscussion(ctx context.Context, proposal uint64, speaker string, text string) error
	GetSelfIntro(ctx context.Context) (string, error)
	GetHeadPhoto(ctx context.Context) (string, error)
}

var _ Client = &MockClient{}


type VoteGrantReq struct {
	GrantId          uint64 `json:"grantId"`
	ValidatorAddress string `json:"validatorAddress"`
	Text             string `json:"text"`
}

type AddDiscussionReq struct {
	ProposalId       uint64 `json:"proposalId"`
	ValidatorAddress string `json:"validatorAddress"`
	Text             string `json:"text"`
}

type AddProposalReq struct {
	ProposalId       uint64 `json:"proposalId"`
	ValidatorAddress string `json:"validatorAddress"`
	Text             string `json:"text"`
}

type VoteResponse struct {
	Vote   string `json:"vote"`
	Reason string `json:"reason"`
}

type MockClient struct {
	Url     string
	logger  cmtlog.Logger
	AgentId string
}

func (m *MockClient) GetAgentIds(ctx context.Context) ([]string, error) {
	mockResponse := `{
		"agents": [
			{
				"id": "mock-agent-1",
				"name": "Agent One"
			},
			{
				"id": "mock-agent-2",
				"name": "Agent Two"
			}
		]
	}`
	var agents struct {
		Agents []struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		} `json:"agents"`
	}

	err := json.Unmarshal([]byte(mockResponse), &agents)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(agents.Agents))
	for _, ag := range agents.Agents {
		ids = append(ids, ag.Id)
	}
	return ids, nil
}

type MockClient struct {
	Url     string
	logger  cmtlog.Logger
	AgentId string
}

func (m *MockClient) GetAgentIds(ctx context.Context) ([]string, error) {
	mockResponse := `{
		"agents": [
			{
				"id": "mock-agent-1",
				"name": "Agent One"
			},
			{
				"id": "mock-agent-2",
				"name": "Agent Two"
			}
		]
	}`
	var agents struct {
		Agents []struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		} `json:"agents"`
	}

	err := json.Unmarshal([]byte(mockResponse), &agents)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(agents.Agents))
	for _, ag := range agents.Agents {
		ids = append(ids, ag.Id)
	}
	return ids, nil
}

func (m *MockClient) GetHeadPhoto(ctx context.Context) (string, error) {
	return "", nil
}

func (m *MockClient) GetSelfIntro(ctx context.Context) (string, error) {
	return "mock", nil
}

func (m *MockClient) AddDiscussion(ctx context.Context, proposal uint64, speaker string, text string) error {
	return nil
}

func (m *MockClient) AddProposal(ctx context.Context, proposal uint64, proposer string, text string) error {
	return nil
}

func (m *MockClient) CommentPropoal(ctx context.Context, proposal uint64, speaker string) (string, error) {
	return "", nil
}

func NewMockClient(url string, logger cmtlog.Logger) (*MockClient, error) {
	l := logger.With("module", "mock")
	client := &MockClient{
		Url:    url,
		logger: l,
	}
	ids, err := client.GetAgentIds(context.Background())
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, errors.New("no agent id")
	}
	client.AgentId = ids[0]
	return client, nil
}

func (m *MockClient) IfAcceptProposal(ctx context.Context, proposal uint64, voter string) (bool, error) {
	return true, nil
}

func (m *MockClient) IfGrantNewMember(ctx context.Context, validator uint64, proposer string, amount uint64, statement string) (bool, error) {
	return true, nil
}

func (m *MockClient) IfProcessProposal(ctx context.Context, data []byte) (bool, error) {
	return true, nil
}
