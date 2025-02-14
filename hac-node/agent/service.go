package agent

import (
	"log"
	"net/http"
	"sort"

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
	g := s.engine.Group("/api")
	g.POST("/proposals", s.handleGetProposals)
	g.POST("/discussions", s.handleGetDiscussions)
	g.POST("/grants", s.handleGetGrants)
	g.POST("/agents", s.handleGetAgents)
	g.POST("/agent-detail", s.handleGetAgentDetail)
	g.POST("/proposal-detail", s.handleGetProposalDetail)
	g.GET("/manifesto", s.handleGetManifesto)
	g.GET("/network-status", s.handleGetNetworkStatus)
	g.GET("/latest-blocks", s.handleGetLatestBlocks)
	return s
}

func (s *Service) Start() {
	err := s.engine.Run(s.listenAddr)
	if err != nil {
		log.Fatal(err)
	}
}

type VoteInfo struct {
	Pass         bool   `json:"pass"`
	VoterIndex   uint64 `json:"voter_index"`
	VoterAddress string `json:"voter_address"`
	Height       uint64 `json:"height"`
	VoteCode     uint64 `json:"voteCode"`
}
type ProposalInfo struct {
	Proposal       Proposal   `json:"proposal"`
	DiscussoinCnt  int        `json:"discussionCnt"`
	DraftVotes     []VoteInfo `json:"draftVotes"`
	DraftPass      uint64     `json:"draftPass"`
	DraftReject    uint64     `json:"draftReject"`
	DecisionVote   []VoteInfo `json:"decisionVotes"`
	DecisionPass   uint64     `json:"decisionPass"`
	DecisionReject uint64     `json:"decisionReject"`
}

type ProposalDetail struct {
	Proposal      Proposal       `json:"proposal"`
	DecisionSteps []DecisionStep `json:"decisionSteps"`
}

type DecisionStep struct {
	Discussions    []Discussion `json:"discussions"`
	DecisionVote   []VoteInfo   `json:"decisionVotes"`
	DecisionPass   uint64       `json:"decisionPass"`
	DecisionReject uint64       `json:"decisionReject"`
}

type GrantInfo struct {
	Grant Grant      `json:"grant"`
	Votes []VoteInfo `json:"votes"`
}

type AgentInfo struct {
	Agent     ValidatorAgent `json:"agent"`
	Proposals []ProposalInfo `json:"proposals"`
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

type GetAccountDetailReq struct {
	Address string `json:"address"`
}

type GetAccountDetailResponse struct {
	AgentInfo AgentInfo `json:"agentInfo"`
}

func (s *Service) handleGetAgentDetail(c *gin.Context) {
	var response GetAccountDetailResponse
	response.AgentInfo.Proposals = make([]ProposalInfo, 0)
	var requestData GetAccountDetailReq
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	agent, err := s.indexer.getValidatorByAddress(requestData.Address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if agent == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent not found"})
		return
	}
	response.AgentInfo.Agent = *agent
	proposals, _, err := s.indexer.getProposalsByProposerAddr(requestData.Address, 0, 1000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, proposal := range proposals {
		proposalInfo, err := s.getProposalInfoById(proposal.Id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		response.AgentInfo.Proposals = append(response.AgentInfo.Proposals, proposalInfo)
	}
	c.JSON(http.StatusOK, response)
}

type GetManifestoResponse struct {
	Manifesto string `json:"manifesto"`
}

func (s *Service) handleGetManifesto(c *gin.Context) {
	c.JSON(http.StatusOK, GetManifestoResponse{Manifesto: MANIFESTO})
}

type GetNetworkStatusResponse struct {
	BlockHeight         uint64 `json:"blockHeight"`
	LastProposer        string `json:"lastProposer"`
	LastProposerAddress string `json:"lastProposerAddress"`
	ProposalsInProgress uint64 `json:"proposalsInProgress"`
	ProposalsDecided    uint64 `json:"proposalsDecided"`
}

func (s *Service) handleGetNetworkStatus(c *gin.Context) {
	var response GetNetworkStatusResponse
	response.BlockHeight = uint64(s.indexer.Height)
	proposals, _, err := s.indexer.getProposals(0, 1)
	if err != nil {
		s.indexer.logger.Error("get proposals", "error", err)
	}
	if len(proposals) != 0 {
		validator, err := s.indexer.getValidatorByAddress(proposals[0].ProposerAddress)
		if err != nil {
			s.indexer.logger.Error("get validator by address", "error", err)
		}
		response.LastProposerAddress = validator.Address
		if validator.Name != "" {
			response.LastProposer = validator.Name
		}
	}
	proposalsInProgress, err := s.indexer.getProposalsInProcess()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	proposalsDecided, err := s.indexer.getProposalsDecided()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response.ProposalsInProgress = proposalsInProgress
	response.ProposalsDecided = proposalsDecided
	c.JSON(http.StatusOK, response)
}

type BlockInfo struct {
	Height          uint64 `json:"height"`
	Proposer        string `json:"proposer"`
	ProposerId      uint64 `json:"proposerId"`
	ProposerAddress string `json:"proposerAddress"`
	ProposalId      int64  `json:"proposalId"`
	Discussions     uint64 `json:"discussions"`
	TransactionCnt  uint64 `json:"transactionCnt"`
}

type GetLatestBlocksResponse struct {
	Blocks []BlockInfo `json:"blocks"`
}

func (s *Service) handleGetLatestBlocks(c *gin.Context) {
	response := GetLatestBlocksResponse{make([]BlockInfo, 0)}
	info := BlockInfo{
		Height:      uint64(s.indexer.Height),
		Proposer:    "",
		ProposerId:  0,
		ProposalId:  0,
		Discussions: 0,
	}
	block := s.indexer.BlockStore.LoadBlock(s.indexer.Height)
	if block == nil {
		c.JSON(http.StatusOK, response)
		return
	}
	validator, err := s.indexer.getValidatorByAddress(block.Header.ProposerAddress.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	info.Proposer = validator.Name
	info.ProposerId = validator.Id
	info.ProposerAddress = validator.Address
	info.TransactionCnt = uint64(len(block.Txs))

	proposal, err := s.indexer.getProposalByHeight(uint64(s.indexer.Height))
	if proposal == nil || proposal.Id == 0 {
		info.ProposalId = -1
	} else {
		info.ProposalId = int64(proposal.Id)
	}

	discussions, err := s.indexer.getDiscussionCntByHeight(uint64(s.indexer.Height))
	info.Discussions = discussions
	response.Blocks = append(response.Blocks, info)
	c.JSON(http.StatusOK, response)
}

type GetAccountsReq struct{}

type GetAccountsResponse struct {
	Agents []ValidatorAgent `json:"agents"`
}

func (s *Service) handleGetAgents(c *gin.Context) {
	var response GetAccountsResponse
	response.Agents = make([]ValidatorAgent, 0)
	agents, err := s.indexer.getValidators()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response.Agents = agents
	c.JSON(http.StatusOK, response)
}

func (s *Service) handleGetGrants(c *gin.Context) {
	var response GetGrantResponse
	response.Grants = make([]GrantInfo, 0)
	var requestData GetGrantsReq
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	requestData.Page -= 1
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
	requestData.Page -= 1
	if requestData.ProposalId != 0 {
		discussions, total, err := s.indexer.getDiscussionByProposal(requestData.ProposalId, requestData.Page, requestData.PageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for i, _ := range discussions {
			agent, err := s.indexer.getValidatorByAddress(discussions[i].SpeakerAddress)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			discussions[i].HeadPhoto = agent.HeadPhoto
		}
		response.Discussions = discussions
		response.Total = total
		c.JSON(http.StatusOK, response)
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": "proposalId is required"})
	return
}

type GetProposalDetailReq struct {
	ProposalId uint64 `json:"proposalId"`
}

func (s *Service) handleGetProposalDetail(c *gin.Context) {
	response := ProposalDetail{
		Proposal:      Proposal{},
		DecisionSteps: []DecisionStep{},
	}
	var requestData GetProposalDetailReq
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	proposalInfo, err := s.getProposalInfoById(requestData.ProposalId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if proposalInfo.Proposal.Id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "proposal not found"})
		return
	}
	response.Proposal = proposalInfo.Proposal
	discussions, _, err := s.indexer.getDiscussionByProposal(requestData.ProposalId, 0, proposalInfo.DiscussoinCnt+1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	votes := proposalInfo.DecisionVote
	if len(votes) == 0 {
		response.DecisionSteps = append(response.DecisionSteps, DecisionStep{
			Discussions:    discussions,
			DecisionVote:   []VoteInfo{},
			DecisionPass:   0,
			DecisionReject: 0,
		})
		c.JSON(http.StatusOK, response)
		return
	}
	// sort votes by height
	sort.Slice(votes, func(i, j int) bool {
		return votes[i].Height < votes[j].Height
	})

	// sort decisions by height
	sort.Slice(discussions, func(i, j int) bool {
		return discussions[i].Height < discussions[j].Height
	})

	stepVotes := make([]VoteInfo, 0)
	stepDiscussions := make([]Discussion, 0)
	stepHeigt := uint64(0)
	lastStepHeigt := uint64(0)
	for i, vote := range votes {
		if vote.Height != stepHeigt || i == len(votes)-1 {
			if len(stepVotes) > 0 {
				pass := 0
				reject := 0
				for _, v := range stepVotes {
					if v.Pass {
						pass++
					} else {
						reject++
					}
				}
				response.DecisionSteps = append(response.DecisionSteps, DecisionStep{
					Discussions:    stepDiscussions,
					DecisionVote:   stepVotes,
					DecisionPass:   uint64(pass),
					DecisionReject: uint64(reject),
				})
			}
			stepVotes = []VoteInfo{vote}
			stepDiscussions = make([]Discussion, 0)
			lastStepHeigt = stepHeigt
			stepHeigt = vote.Height
			for _, discussion := range discussions {
				if discussion.Height >= lastStepHeigt && discussion.Height < stepHeigt {
					stepDiscussions = append(stepDiscussions, discussion)
				}
			}
			continue
		}
		stepVotes = append(stepVotes, vote)
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
	requestData.Page -= 1

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
	} else {
		proposals, proposalTotal, err = s.indexer.getProposals(requestData.Page, requestData.PageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
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
	agent, err := s.indexer.getValidatorByAddress(proposal.ProposerAddress)
	if err != nil {
		return ProposalInfo{}, err
	}
	proposal.HeadPhoto = agent.HeadPhoto
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
		Proposal:       proposal,
		DiscussoinCnt:  int(total),
		DraftVotes:     draftVotes,
		DraftPass:      0,
		DraftReject:    0,
		DecisionVote:   decisionVotes,
		DecisionPass:   0,
		DecisionReject: 0,
	}
	for _, vote := range draftVotes {
		if vote.Pass {
			proposalInfo.DraftPass++
		} else {
			proposalInfo.DraftReject++
		}
	}

	for _, vote := range decisionVotes {
		if vote.Pass {
			proposalInfo.DecisionPass++
		} else {
			proposalInfo.DecisionReject++
		}
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
