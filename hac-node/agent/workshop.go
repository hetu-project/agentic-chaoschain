package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"

	cmtlog "github.com/cometbft/cometbft/libs/log"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type AgentClient struct {
	Url    string
	logger cmtlog.Logger
}

func NewAgentClient(url string, logger cmtlog.Logger) *AgentClient {
	return &AgentClient{
		Url:    url,
		logger: logger.With("module", "workshop_agent"),
	}
}

func (e *AgentClient) AddDiscussion(ctx context.Context, proposal uint64, speaker string, text string) error {
	e.logger.Info("AddDiscussion", "proposal", proposal, "speaker", speaker, "text", text)
	url, _ := neturl.JoinPath(e.Url, "/add_discussion")
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

func (e *AgentClient) AddProposal(ctx context.Context, proposal uint64, proposer string, text string) error {
	e.logger.Info("AddProposal", "proposal", proposal, "proposer", proposer, "text", text)
	url, _ := neturl.JoinPath(e.Url, "/add_proposal")
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

type CommentResponse struct {
	ProposalId uint64 `json:"proposalId"`
	Comment    string `json:"comment"`
}

func (e *AgentClient) CommentPropoal(ctx context.Context, proposal uint64, speaker string) (string, error) {
	e.logger.Info("CommentPropoal", "proposal", proposal, "speaker", speaker)
	url, _ := neturl.JoinPath(e.Url, "/new_discussion")
	body := fmt.Sprintf(`{"proposalId":"%d"}`, proposal)
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
	var comment CommentResponse
	err = json.Unmarshal(bodyBytes, &comment)
	if err != nil {
		e.logger.Error("CommentPropoal unmarshal response body fail", "err", err)
		return "", err
	}
	e.logger.Info("comment proposal", "proposal", proposal, "speaker", speaker, "comment", comment.Comment)
	err = Indexer.sendDiscussion(proposal, comment.Comment)
	if err != nil {
		e.logger.Error("send discussion tx fail", "err", err)
	}
	return comment.Comment, nil
}

func (e *AgentClient) GetHeadPhoto(ctx context.Context) (string, error) {
	return "", nil
}

func (e *AgentClient) GetSelfIntro(ctx context.Context) (string, error) {
	return "", nil
}

func (e *AgentClient) IfAcceptProposal(ctx context.Context, proposal uint64, voter string) (bool, error) {
	e.logger.Info("IfAcceptProposal", "proposal", proposal, "voter", voter)
	url, _ := neturl.JoinPath(e.Url, "/voteproposal")
	body := fmt.Sprintf(`{"proposalId":"%d"}`, proposal)
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

// IfGrantNewMember implements Client.
func (e *AgentClient) IfGrantNewMember(ctx context.Context, validator uint64, proposer string, amount uint64, statement string) (bool, error) {
	return true, nil
}

type IfProcessProposalReq struct {
	Proposal string `json:"proposal"`
	Title    string `json:"title"`
}

func (e *AgentClient) IfProcessProposal(ctx context.Context, proposal, title string) (bool, error) {
	e.logger.Info("IfProcessProposal", "proposal", proposal, "title", title)
	url, _ := neturl.JoinPath(e.Url, "/if_process_pr")
	req := IfProcessProposalReq{
		Proposal: proposal,
		Title:    title,
	}
	data, _ := json.Marshal(&req)
	res, err := http.Post(url, "application/json", bytes.NewBuffer(data))
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
	e.logger.Info("vote proposal", "proposal", proposal, "vote", vote.Vote, "reason", vote.Reason)
	if vote.Vote == "yes" {
		return true, nil
	}
	return false, nil
}
