package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
)

// DiscussionItem defines the properties of a discussion
type DiscussionItem struct {
	ID          uint64 `json:"id"`
	Author      string `json:"author"`
	ContentText string `json:"text"`
	Timestamp   int64  `json:"timestamp"`
}

// ProposalItem defines all properties of a proposal
type ProposalItem struct {
	ProposalID       uint64           `json:"proposalId"`
	ValidatorAddress string           `json:"validatorAddress"`
	Title            string           `json:"title"`
	ProposalText     string           `json:"text"`
	Vote             string           `json:"vote"`
	Timestamp        int64            `json:"timestamp"`
	Discussions      []DiscussionItem `json:"discussions"`
}

// ProposalDB provides storage and retrieval functionality for proposal data
type ProposalDB struct {
	db *leveldb.DB
}

// NewProposalDB creates and initializes a new proposal database
func NewProposalDB(dataDir string) (*ProposalDB, error) {
	dbPath := filepath.Join(dataDir, "proposals")
	db, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open proposals database: %w", err)
	}

	return &ProposalDB{db: db}, nil
}

// Close closes the database connection
func (pdb *ProposalDB) Close() error {
	return pdb.db.Close()
}

// SaveProposal saves a proposal to the database
func (pdb *ProposalDB) SaveProposal(proposal *ProposalItem) error {
	// If timestamp is not set, set the current time
	if proposal.Timestamp == 0 {
		proposal.Timestamp = time.Now().Unix()
	}

	// Convert proposal ID to key
	key := fmt.Sprintf("proposal:%d", proposal.ProposalID)

	// Serialize the proposal to JSON
	data, err := json.Marshal(proposal)
	if err != nil {
		return fmt.Errorf("failed to marshal proposal: %w", err)
	}

	// Save to database
	err = pdb.db.Put([]byte(key), data, nil)
	if err != nil {
		return fmt.Errorf("failed to save proposal to db: %w", err)
	}

	return nil
}

// GetProposal retrieves a proposal by ID from the database
func (pdb *ProposalDB) GetProposal(proposalID uint64) (*ProposalItem, error) {
	key := fmt.Sprintf("proposal:%d", proposalID)

	data, err := pdb.db.Get([]byte(key), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, nil // Proposal doesn't exist
		}
		return nil, fmt.Errorf("failed to get proposal from db: %w", err)
	}

	var proposal ProposalItem
	if err := json.Unmarshal(data, &proposal); err != nil {
		return nil, fmt.Errorf("failed to unmarshal proposal data: %w", err)
	}

	return &proposal, nil
}

// AddDiscussion adds a discussion to a specific proposal
func (pdb *ProposalDB) AddDiscussion(proposalID uint64, author string, content string) error {
	proposal, err := pdb.GetProposal(proposalID)
	if err != nil {
		return err
	}
	if proposal == nil {
		return fmt.Errorf("proposal with ID %d not found", proposalID)
	}

	// Create new discussion
	newDiscussion := DiscussionItem{
		ID:          uint64(len(proposal.Discussions) + 1),
		Author:      author,
		ContentText: content,
		Timestamp:   time.Now().Unix(),
	}

	// Add to the proposal's discussion list
	proposal.Discussions = append(proposal.Discussions, newDiscussion)

	// Save the updated proposal
	return pdb.SaveProposal(proposal)
}

// GetDiscussions gets all discussions for a specific proposal
func (pdb *ProposalDB) GetDiscussions(proposalID uint64) ([]DiscussionItem, error) {
	proposal, err := pdb.GetProposal(proposalID)
	if err != nil {
		return nil, err
	}
	if proposal == nil {
		return nil, fmt.Errorf("proposal with ID %d not found", proposalID)
	}

	return proposal.Discussions, nil
}

// UpdateVote updates the voting result of a proposal
func (pdb *ProposalDB) UpdateVote(proposalID uint64, vote string) error {
	proposal, err := pdb.GetProposal(proposalID)
	if err != nil {
		return err
	}
	if proposal == nil {
		return fmt.Errorf("proposal with ID %d not found", proposalID)
	}

	proposal.Vote = vote
	return pdb.SaveProposal(proposal)
}

// DeleteProposal deletes a proposal from the database
func (pdb *ProposalDB) DeleteProposal(proposalID uint64) error {
	key := fmt.Sprintf("proposal:%d", proposalID)
	return pdb.db.Delete([]byte(key), nil)
}

// ListProposals lists all proposals
func (pdb *ProposalDB) ListProposals() ([]*ProposalItem, error) {
	var proposals []*ProposalItem
	iter := pdb.db.NewIterator(nil, nil)
	defer iter.Release()

	for iter.Next() {
		key := string(iter.Key())
		// Only the proposal keys are processed
		if len(key) >= 9 && key[:9] == "proposal:" {
			var proposal ProposalItem
			if err := json.Unmarshal(iter.Value(), &proposal); err != nil {
				log.Printf("Failed to unmarshal proposal: %v", err)
				continue
			}
			proposals = append(proposals, &proposal)
		}
	}

	if err := iter.Error(); err != nil {
		return nil, fmt.Errorf("error iterating over proposals: %w", err)
	}

	return proposals, nil
}

// GetProposalsByValidator gets all proposals from a specific validator
func (pdb *ProposalDB) GetProposalsByValidator(validatorAddress string) ([]*ProposalItem, error) {
	var proposals []*ProposalItem
	iter := pdb.db.NewIterator(nil, nil)
	defer iter.Release()

	for iter.Next() {
		var proposal ProposalItem
		if err := json.Unmarshal(iter.Value(), &proposal); err != nil {
			log.Printf("Failed to unmarshal proposal: %v", err)
			continue
		}

		if proposal.ValidatorAddress == validatorAddress {
			proposals = append(proposals, &proposal)
		}
	}

	if err := iter.Error(); err != nil {
		return nil, fmt.Errorf("error iterating over proposals: %w", err)
	}

	return proposals, nil
}

// ExportProposalsToFile exports all proposal data to a specified text file
func (pdb *ProposalDB) ExportProposalsToFile(filePath string) error {
	proposals, err := pdb.ListProposals()
	if err != nil {
		return fmt.Errorf("failed to list proposals: %w", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	for _, p := range proposals {
		_, err := fmt.Fprintf(file, "Proposal ID: %d\n", p.ProposalID)
		if err != nil {
			return fmt.Errorf("write error: %w", err)
		}
		fmt.Fprintf(file, "Validator: %s\n", p.ValidatorAddress)
		fmt.Fprintf(file, "Title: %s\n", p.Title)
		fmt.Fprintf(file, "Vote Result: %s\n", p.Vote)
		fmt.Fprintf(file, "Created At: %s\n", time.Unix(p.Timestamp, 0).UTC())
		fmt.Fprintf(file, "Discussions (%d):\n", len(p.Discussions))

		for _, d := range p.Discussions {
			fmt.Fprintf(file, " - [%s @ %s] %s\n",
				d.Author,
				time.Unix(d.Timestamp, 0).UTC().Format(time.RFC3339),
				d.ContentText)
		}
		fmt.Fprintln(file, "\n"+strings.Repeat("-", 80)+"\n")
	}

	return nil
}

// ExportProposalsToJSON exports all proposal data as JSON file
// Automatically creates directories in path if missing, defaults to proposal-info.json in current directory
func (pdb *ProposalDB) ExportProposalsToJSON(filePath string) error {
	if filePath == "" {
		filePath = "proposal-info.json"
	}

	// Ensure directory exists
	if dir := filepath.Dir(filePath); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	proposals, err := pdb.ListProposals()
	if err != nil {
		return fmt.Errorf("failed to list proposals: %w", err)
	}

	// Create formatted JSON data
	jsonData, err := json.MarshalIndent(proposals, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write file atomically
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	return nil
}

// GetNextProposalID gets the next available proposal ID
func (pdb *ProposalDB) GetNextProposalID() (uint64, error) {
	var maxID uint64 = 0
	iter := pdb.db.NewIterator(nil, nil)
	defer iter.Release()

	for iter.Next() {
		var proposal ProposalItem
		if err := json.Unmarshal(iter.Value(), &proposal); err != nil {
			continue
		}

		if proposal.ProposalID > maxID {
			maxID = proposal.ProposalID
		}
	}

	if err := iter.Error(); err != nil {
		return 0, fmt.Errorf("error iterating over proposals: %w", err)
	}

	return maxID + 1, nil
}
