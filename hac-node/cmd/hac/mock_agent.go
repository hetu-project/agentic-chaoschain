package main

import (
	"net/http"
	"time"

	"github.com/calehh/hac-app/agent"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

type MockArguments struct {
	Address string
	Vote    bool
}

var mockArguments MockArguments

var mockCmd = &cobra.Command{
	Use:   "mock",
	Short: "",
	Long:  ``,
	Run:   mockRun,
}

func init() {
	mockCmd.Flags().StringVarP(&mockArguments.Address, "address", "a", "0.0.0.0:3631", "proposal data")
	mockCmd.Flags().BoolVarP(&mockArguments.Vote, "vote", "v", false, "vote false")
}

func mockRun(cmd *cobra.Command, args []string) {
	r := gin.Default()

	r.POST("/add_discussion", func(c *gin.Context) {
		var req agent.AddDiscussionReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "discussion added"})
	})

	r.POST("/add_proposal", func(c *gin.Context) {
		var req agent.AddProposalReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "proposal added"})
	})

	r.POST("/new_discussion", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		comment := "mock comment" + time.Now().Format(time.RFC1123Z)
		c.JSON(http.StatusOK, gin.H{"comment": comment})
	})

	r.POST("/voteproposal", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		voteRes := "no"
		if mockArguments.Vote {
			voteRes = "yes"
		}
		c.JSON(http.StatusOK, gin.H{"vote": voteRes, "reason": "mock"})
	})

	r.POST("/if_process_pr", func(c *gin.Context) {
		var req agent.IfProcessProposalReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		voteRes := "no"
		if mockArguments.Vote {
			voteRes = "yes"
		}
		c.JSON(http.StatusOK, gin.H{"vote": voteRes, "reason": "mock"})
	})
	r.Run(mockArguments.Address)
}
