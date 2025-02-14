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

var ElizaCli Client

var DiscussionRate = 0

var DiscussionTrigger = 0

type Client interface {
	IfProcessProposal(ctx context.Context, proposer uint64, data []byte) (bool, error)
	IfAcceptProposal(ctx context.Context, proposal uint64, voter string) (bool, error)
	IfGrantNewMember(ctx context.Context, validator uint64, proposer string, amount uint64, statement string) (bool, error)
	CommentPropoal(ctx context.Context, proposal uint64, speaker string) (string, error)
	AddProposal(ctx context.Context, proposal uint64, proposer string, text string) error
	AddDiscussion(ctx context.Context, proposal uint64, speaker string, text string) error
	GetSelfIntro(ctx context.Context) (string, error)
	GetHeadPhoto(ctx context.Context) (string, error)
}

var _ Client = &MockClient{}
var _ Client = &ElizaClient{}

type ElizaClient struct {
	Url     string
	AgentId string
	logger  cmtlog.Logger
}

func (c *ElizaClient) GetHeadPhoto(ctx context.Context) (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s/%s/headphoto", c.Url, c.AgentId))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func (c *ElizaClient) GetSelfIntro(ctx context.Context) (string, error) {
	agentUrl, err := url.JoinPath(c.Url, c.AgentId, "/selfintro")
	if err != nil {
		c.logger.Error("join url fail", "err", err)
		return "", err
	}
	res, err := http.Get(agentUrl)
	if err != nil {
		c.logger.Error("get agent url fail", "err", err)
		return "", err
	}
	buf, err := io.ReadAll(res.Body)
	if err != nil {
		c.logger.Error("read response body fail", "err", err)
		return "", err
	}
	defer res.Body.Close()
	type SelfIntro struct {
		Character string `json:"character"`
	}
	var selfIntro SelfIntro
	err = json.Unmarshal(buf, &selfIntro)
	if err != nil {
		c.logger.Error("unmarshal response body fail", "err", err)
		return "", err
	}
	return selfIntro.Character, nil
}

func NewElizaClient(url string, logger cmtlog.Logger) (*ElizaClient, error) {
	l := logger.With("module", "eliza")
	client := &ElizaClient{
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

func (e *ElizaClient) GetAgentIds(ctx context.Context) ([]string, error) {
	url := fmt.Sprintf("%s/agents", e.Url)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var agents struct {
		Agents []struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		} `json:"agents"`
	}
	err = json.Unmarshal(bodyBytes, &agents)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(agents.Agents))
	for _, ag := range agents.Agents {
		ids = append(ids, ag.Id)
	}
	return ids, nil
}

type VoteGrantReq struct {
	GrantId          uint64 `json:"grantId"`
	ValidatorAddress string `json:"validatorAddress"`
	Text             string `json:"text"`
}

func (e *ElizaClient) IfGrantNewMember(ctx context.Context, validator uint64, proposer string, amount uint64, statement string) (bool, error) {
	e.logger.Info("IfGrantNewMember", "validator", validator, "proposer", proposer, "amount", amount, "statement", statement)
	url := fmt.Sprintf("%s/%s/votegrant", e.Url, e.AgentId)
	req := VoteGrantReq{
		GrantId:          validator,
		ValidatorAddress: proposer,
		Text:             statement,
	}
	data, _ := json.Marshal(req)
	res, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(data)))
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		e.logger.Error("read response body fail", "err", err)
		return false, err
	}
	var vote VoteResponse
	err = json.Unmarshal(bodyBytes, &vote)
	if err != nil {
		e.logger.Error("unmarshal response body fail", "err", err)
		return false, err
	}
	e.logger.Info("vote grant", "validator", validator, "proposer", proposer, "vote", vote.Vote, "reason", vote.Reason)
	if vote.Vote == "yes" {
		return true, nil
	}
	return false, nil
}

func (e *ElizaClient) CommentPropoal(ctx context.Context, proposal uint64, speaker string) (string, error) {
	e.logger.Info("CommentPropoal", "proposal", proposal, "speaker", speaker)
	url := fmt.Sprintf("%s/%s/newdiscussion", e.Url, e.AgentId)
	body := fmt.Sprintf(`{"proposalId":"%d","validatorAddress":"%s","text":"comment"}`, proposal, speaker)
	res, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(body)))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		e.logger.Error("read response body fail", "err", err)
		return "", err
	}
	e.logger.Info("comment proposal", "proposal", proposal, "speaker", speaker, "comment", string(bodyBytes))
	return string(bodyBytes), nil
}

type AddDiscussionReq struct {
	ProposalId       uint64 `json:"proposalId"`
	ValidatorAddress string `json:"validatorAddress"`
	Text             string `json:"text"`
}

func (e *ElizaClient) AddDiscussion(ctx context.Context, proposal uint64, speaker string, text string) error {
	e.logger.Info("AddDiscussion", "proposal", proposal, "speaker", speaker, "text", text)
	url := fmt.Sprintf("%s/%s/discussion", e.Url, e.AgentId)
	req := AddDiscussionReq{
		ProposalId:       proposal,
		ValidatorAddress: speaker,
		Text:             text,
	}
	data, _ := json.Marshal(req)
	res, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(data)))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	e.logger.Info("add discussion", "proposal", proposal, "speaker", speaker, "text", text)
	return nil
}

type AddProposalReq struct {
	ProposalId       uint64 `json:"proposalId"`
	ValidatorAddress string `json:"validatorAddress"`
	Text             string `json:"text"`
}

func (e *ElizaClient) AddProposal(ctx context.Context, proposal uint64, proposer string, text string) error {
	e.logger.Info("AddProposal", "proposal", proposal, "proposer", proposer, "text", text)
	url := fmt.Sprintf("%s/%s/proposal", e.Url, e.AgentId)
	req := AddProposalReq{
		ProposalId:       proposal,
		ValidatorAddress: proposer,
		Text:             text,
	}
	data, _ := json.Marshal(req)
	res, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(data)))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	data, err = io.ReadAll(res.Body)
	resp := ""
	if err == nil {
		resp = string(data)
	}
	e.logger.Info("add proposal", "proposal", proposal, "proposer", proposer, "text", text, "resp", resp)
	return nil
}

type VoteResponse struct {
	Vote   string `json:"vote"`
	Reason string `json:"reason"`
}

func (e *ElizaClient) IfAcceptProposal(ctx context.Context, proposal uint64, voter string) (bool, error) {
	e.logger.Info("IfAcceptProposal", "proposal", proposal, "voter", voter)
	url := fmt.Sprintf("%s/%s/voteproposal", e.Url, e.AgentId)
	body := fmt.Sprintf(`{"proposalId":"%d","validatorAddress":"%s","text":"analyze proposal"}`, proposal, voter)
	res, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(body)))
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		e.logger.Error("read response body fail", "err", err)
		return false, err
	}
	var vote VoteResponse
	err = json.Unmarshal(bodyBytes, &vote)
	if err != nil {
		e.logger.Error("unmarshal response body fail", "err", err)
		return false, err
	}
	e.logger.Info("vote proposal", "proposal", proposal, "voter", voter, "vote", vote.Vote, "reason", vote.Reason)
	if vote.Vote == "yes" {
		return true, nil
	}
	return false, nil
}

func (e *ElizaClient) IfProcessProposal(ctx context.Context, proposer uint64, data []byte) (bool, error) {
	return true, nil
}

type MockClient struct {
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

func NewMockClient() *MockClient {
	return &MockClient{}
}

func (m *MockClient) IfAcceptProposal(ctx context.Context, proposal uint64, voter string) (bool, error) {
	return true, nil
}

func (m *MockClient) IfGrantNewMember(ctx context.Context, validator uint64, proposer string, amount uint64, statement string) (bool, error) {
	return true, nil
}

func (m *MockClient) IfProcessProposal(ctx context.Context, proposer uint64, data []byte) (bool, error) {
	return true, nil
}
