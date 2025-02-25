package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Proposal struct {
	ID               uint64 `json:"proposalId"`
	ValidatorAddress string `json:"validatorAddress"`
	Text             string `json:"text"`
}

type Discussion struct {
	ProposalID       uint64 `json:"proposalId"`
	ValidatorAddress string `json:"validatorAddress"`
	Text             string `json:"text"`
}

type VoteRequest struct {
	Proposal string `json:"proposal"`
	Title    string `json:"title"`
}

type VoteResponse struct {
	Vote string `json:"vote"`
}

func (h *HTTPHandler) AddProposal(c *gin.Context) {
	// Parse the incoming JSON request
	var req struct {
		ProposalID       uint64 `json:"proposalId"`
		ValidatorAddress string `json:"validatorAddress"`
		Text             string `json:"text"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create a new ProposalItem from the request data
	proposal := &ProposalItem{
		ProposalID:       req.ProposalID,
		ValidatorAddress: req.ValidatorAddress,
		ProposalText:     req.Text,
		Title:            "", // Empty title since it's not in the request
		Vote:             "", // Initialize with empty vote
		Timestamp:        time.Now().Unix(),
		Discussions:      []DiscussionItem{}, // Initialize with empty discussions
	}

	// Save the proposal to the database using ProposalDB
	if err := h.proposalDB.SaveProposal(proposal); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save proposal: %v", err)})
		return
	}
	h.logger.Println("AddProposal  ===== ok ======")
	h.proposalDB.ExportProposalsToJSON("proposal_info.json")
	c.JSON(http.StatusOK, gin.H{"message": "Proposal added successfully", "proposalId": proposal.ProposalID})
}
func (h *HTTPHandler) AddDiscussion(c *gin.Context) {
	var d Discussion
	if err := c.ShouldBindJSON(&d); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.discussions[d.ProposalID] = append(h.discussions[d.ProposalID], d)
	c.Status(http.StatusOK)
}

func (h *HTTPHandler) ProcessDraftVote(c *gin.Context) {
	var vr VoteRequest
	if err := c.ShouldBindJSON(&vr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement draft voting logic
	c.JSON(http.StatusOK, VoteResponse{Vote: "yes"})
}

func (h *HTTPHandler) GenerateDiscussion(c *gin.Context) {
	var req struct {
		ProposalID uint64 `json:"proposalId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement discussion generation logic
	c.String(http.StatusOK, "New discussion generated for proposal %d", req.ProposalID)
}

func (h *HTTPHandler) FinalVote(c *gin.Context) {
	var req struct {
		ProposalID uint64 `json:"proposalId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement final voting logic
	c.JSON(http.StatusOK, VoteResponse{Vote: "yes"})
}
