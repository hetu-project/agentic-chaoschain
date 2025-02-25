package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

type HTTPHandler struct {
	logger      *log.Logger
	proposals   map[uint64]Proposal
	discussions map[uint64][]Discussion
	votes       map[string]string
	proposalDB  *ProposalDB
}

func NewHTTPHandler(logger *log.Logger, proposalDB *ProposalDB) *HTTPHandler {
	return &HTTPHandler{
		logger:      logger,
		proposals:   make(map[uint64]Proposal),
		discussions: make(map[uint64][]Discussion),
		votes:       make(map[string]string),
		proposalDB:  proposalDB,
	}
}

func (h *HTTPHandler) RegisterRoutes(router *gin.Engine) {
	router.POST("/add_proposal", h.AddProposal)
	router.POST("/add_discussion", h.AddDiscussion)
	router.POST("/if_process_pr", h.ProcessDraftVote)
	router.POST("/new_discussion", h.GenerateDiscussion)
	router.POST("/voteproposal", h.FinalVote)
}
