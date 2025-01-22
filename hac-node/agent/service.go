package agent

import (
	"net/http"

	"github.com/calehh/hac-app/tx"
	"github.com/gin-gonic/gin"
)

type Service struct {
	engine     *gin.Engine
	indexer    *ChainIndexer
	listenAddr string
}

func NewService(ListenAddr string, indexer *ChainIndexer) *Service {
	r := gin.Default()
	s := &Service{
		engine:     r,
		indexer:    indexer,
		listenAddr: ListenAddr,
	}
	s.engine.POST("/getProposals", s.handleGetProposals)
	s.engine.POST("/getDiscussions", s.handleGetDiscussions)
	s.engine.POST("/getGrants", s.handleGetGrants)
	return s
}

func (s *Service) Start() {
	s.engine.Run(s.listenAddr)
}

type VoteInfo struct {
	Pass         bool   `json:"pass"`
	VoterIndex   uint64 `json:"voter_index"`
	VoterAddress string `json:"voter_address"`
	Height       uint64 `json:"height"`
	VoteCode     uint64 `json:"voteCode"`
}
type ProposalInfo struct {
	Proposal      Proposal   `json:"proposal"`
	DiscussoinCnt int        `json:"discussionCnt"`
	DraftVotes    []VoteInfo `json:"draftVotes"`
	DecisionVote  []VoteInfo `json:"decisionVotes"`
}

type GrantInfo struct {
	Grant Grant      `json:"grant"`
	Votes []VoteInfo `json:"votes"`
}

type GetGrantsReq struct {
	GrantId  uint64 `json:"grantId"`
	Address  string `json:"address"`
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
}

type GetGrantResponse struct {
	Grants []GrantInfo `json:"grants"`
	Total  uint64      `json:"total"`
}

func (s *Service) handleGetGrants(c *gin.Context) {
	var response GetGrantResponse
	response.Grants = make([]GrantInfo, 0)
	var requestData GetGrantsReq
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if requestData.GrantId != 0 {
		grant, err := s.indexer.getGrantById(requestData.GrantId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		votes, err := s.indexer.getGrantVotesByGrant(requestData.GrantId, 0, 1000)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		voteInfos := GrantVotesToVoteInfo(votes)
		grantInfo := GrantInfo{
			Grant: grant,
			Votes: voteInfos,
		}
		response.Grants = append(response.Grants, grantInfo)
		c.JSON(http.StatusOK, response)
		return
	}

	grants, grantTotal, err := s.indexer.getGrants(requestData.Page, requestData.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response.Total = grantTotal
	for _, grant := range grants {
		votes, err := s.indexer.getGrantVotesByGrant(grant.Id, 0, 1000)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		voteInfos := GrantVotesToVoteInfo(votes)
		grantInfo := GrantInfo{
			Grant: grant,
			Votes: voteInfos,
		}
		response.Grants = append(response.Grants, grantInfo)
	}
	c.JSON(http.StatusOK, response)
}

type GetDiscussionReq struct {
	ProposalId uint64 `json:"proposalId"`
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
}

type GetDiscussionResponse struct {
	Discussions []Discussion `json:"discussions"`
	Total       uint64       `json:"total"`
}

func (s *Service) handleGetDiscussions(c *gin.Context) {
	var response GetDiscussionResponse
	response.Discussions = make([]Discussion, 0)
	var requestData GetDiscussionReq
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if requestData.ProposalId != 0 {
		discussions, total, err := s.indexer.getDiscussionByProposal(requestData.ProposalId, requestData.Page, requestData.PageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		response.Discussions = discussions
		response.Total = total
		c.JSON(http.StatusOK, response)
		return
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "proposalId is required"})
	}
	c.JSON(http.StatusOK, response)
}

type GetProposalsReq struct {
	ProposalId      uint64 `json:"proposalId"`
	ProposerAddress string `json:"proposer"`
	Page            int    `json:"page"`
	PageSize        int    `json:"pageSize"`
}
type GetProposalResponse struct {
	Proposals []ProposalInfo `json:"proposals"`
	Total     uint64         `json:"total"`
}

func (s *Service) handleGetProposals(c *gin.Context) {
	var response GetProposalResponse
	response.Proposals = make([]ProposalInfo, 0)
	var err error
	var requestData GetProposalsReq
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if requestData.ProposalId != 0 {
		proposalInfo, err := s.getProposalInfoById(requestData.ProposalId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		response.Proposals = append(response.Proposals, proposalInfo)
		c.JSON(http.StatusOK, response)
		return
	}
	proposalTotal := uint64(0)
	proposals := make([]Proposal, 0)
	if requestData.ProposerAddress != "" {
		proposals, proposalTotal, err = s.indexer.getProposalsByProposerAddr(requestData.ProposerAddress, requestData.Page, requestData.PageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	proposals, proposalTotal, err = s.indexer.getProposals(requestData.Page, requestData.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response.Total = proposalTotal
	for _, proposal := range proposals {
		proposalInfo, err := s.getProposalInfoById(proposal.Id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		response.Proposals = append(response.Proposals, proposalInfo)
	}
	c.JSON(http.StatusOK, response)
}

func (s *Service) getProposalInfoById(proposalId uint64) (ProposalInfo, error) {
	proposal, err := s.indexer.getProposalById(proposalId)
	if err != nil {
		return ProposalInfo{}, err
	}
	_, total, err := s.indexer.getDiscussionByProposal(proposalId, 0, 1)
	if err != nil {
		return ProposalInfo{}, err
	}
	votes, err := s.indexer.getProposalVotesByProposal(proposalId, 0, 1000)
	if err != nil {
		return ProposalInfo{}, err
	}
	draftVotes, decisionVotes := ProposalVotesToVoteInfo(votes)
	proposalInfo := ProposalInfo{
		Proposal:      proposal,
		DiscussoinCnt: int(total),
		DraftVotes:    draftVotes,
		DecisionVote:  decisionVotes,
	}
	return proposalInfo, nil
}

func GrantVotesToVoteInfo(votes []GrantVote) []VoteInfo {
	grantInfo := GrantInfo{
		Grant: Grant{},
		Votes: []VoteInfo{},
	}

	for _, vote := range votes {
		switch vote.Vote {
		case uint64(tx.VoteGrantNewMember):
			grantInfo.Votes = append(grantInfo.Votes, VoteInfo{
				Pass:         true,
				VoterIndex:   vote.VoterIndex,
				VoterAddress: vote.VoterAddress,
				Height:       vote.Height,
				VoteCode:     vote.Vote,
			})
		case uint64(tx.VoteRejectNewMember):
			grantInfo.Votes = append(grantInfo.Votes, VoteInfo{
				Pass:         false,
				VoterIndex:   vote.VoterIndex,
				VoterAddress: vote.VoterAddress,
				Height:       vote.Height,
				VoteCode:     vote.Vote,
			})
		}
	}
	return grantInfo.Votes
}

func ProposalVotesToVoteInfo(votes []ProposalVote) ([]VoteInfo, []VoteInfo) {
	proposalInfo := ProposalInfo{
		DraftVotes:   []VoteInfo{},
		DecisionVote: []VoteInfo{},
	}

	for _, vote := range votes {
		switch vote.Vote {
		case uint64(tx.VoteIgnoreProposal):
			proposalInfo.DraftVotes = append(proposalInfo.DraftVotes, VoteInfo{
				Pass:         false,
				VoterIndex:   vote.VoterIndex,
				VoterAddress: vote.VoterAddress,
				Height:       vote.Height,
				VoteCode:     vote.Vote,
			})
		case uint64(tx.VoteProcessProposal):
			proposalInfo.DraftVotes = append(proposalInfo.DraftVotes, VoteInfo{
				Pass:         true,
				VoterIndex:   vote.VoterIndex,
				VoterAddress: vote.VoterAddress,
				Height:       vote.Height,
				VoteCode:     vote.Vote,
			})
		case uint64(tx.VoteRejectProposal):
			proposalInfo.DecisionVote = append(proposalInfo.DecisionVote, VoteInfo{
				Pass:         false,
				VoterIndex:   vote.VoterIndex,
				VoterAddress: vote.VoterAddress,
				Height:       vote.Height,
				VoteCode:     vote.Vote,
			})
		case uint64(tx.VoteAcceptProposal):
			proposalInfo.DecisionVote = append(proposalInfo.DecisionVote, VoteInfo{
				Pass:         true,
				VoterIndex:   vote.VoterIndex,
				VoterAddress: vote.VoterAddress,
				Height:       vote.Height,
				VoteCode:     vote.Vote,
			})
		}
	}
	return proposalInfo.DraftVotes, proposalInfo.DecisionVote
}
