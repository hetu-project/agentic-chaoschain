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

// DiscussionItem 结构体定义了讨论的属性
type DiscussionItem struct {
	ID          uint64 `json:"id"`
	Author      string `json:"author"`
	ContentText string `json:"text"`
	Timestamp   int64  `json:"timestamp"`
}

// ProposalItem 结构体定义了提案的所有属性
type ProposalItem struct {
	ProposalID       uint64           `json:"proposalId"`
	ValidatorAddress string           `json:"validatorAddress"`
	Title            string           `json:"title"`
	ProposalText     string           `json:"text"`
	Vote             string           `json:"vote"`
	Timestamp        int64            `json:"timestamp"`
	Discussions      []DiscussionItem `json:"discussions"`
}

// ProposalDB 提供了提案数据的存储和检索功能
type ProposalDB struct {
	db *leveldb.DB
}

// NewProposalDB 创建并初始化一个新的提案数据库
func NewProposalDB(dataDir string) (*ProposalDB, error) {
	dbPath := filepath.Join(dataDir, "proposals")
	db, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open proposals database: %w", err)
	}

	return &ProposalDB{db: db}, nil
}

// Close 关闭数据库连接
func (pdb *ProposalDB) Close() error {
	return pdb.db.Close()
}

// SaveProposal 将提案保存到数据库
func (pdb *ProposalDB) SaveProposal(proposal *ProposalItem) error {
	// 如果没有设置时间戳，设置当前时间
	if proposal.Timestamp == 0 {
		proposal.Timestamp = time.Now().Unix()
	}

	// 将提案ID转换为键
	key := fmt.Sprintf("proposal:%d", proposal.ProposalID)

	// 将提案序列化为JSON
	data, err := json.Marshal(proposal)
	if err != nil {
		return fmt.Errorf("failed to marshal proposal: %w", err)
	}

	// 保存到数据库
	err = pdb.db.Put([]byte(key), data, nil)
	if err != nil {
		return fmt.Errorf("failed to save proposal to db: %w", err)
	}

	return nil
}

// GetProposal 通过ID从数据库获取提案
func (pdb *ProposalDB) GetProposal(proposalID uint64) (*ProposalItem, error) {
	key := fmt.Sprintf("proposal:%d", proposalID)

	data, err := pdb.db.Get([]byte(key), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, nil // 提案不存在
		}
		return nil, fmt.Errorf("failed to get proposal from db: %w", err)
	}

	var proposal ProposalItem
	if err := json.Unmarshal(data, &proposal); err != nil {
		return nil, fmt.Errorf("failed to unmarshal proposal data: %w", err)
	}

	return &proposal, nil
}

// AddDiscussion 向特定的提案添加讨论
func (pdb *ProposalDB) AddDiscussion(proposalID uint64, author string, content string) error {
	proposal, err := pdb.GetProposal(proposalID)
	if err != nil {
		return err
	}
	if proposal == nil {
		return fmt.Errorf("proposal with ID %d not found", proposalID)
	}

	// 创建新的讨论
	newDiscussion := DiscussionItem{
		ID:          uint64(len(proposal.Discussions) + 1),
		Author:      author,
		ContentText: content,
		Timestamp:   time.Now().Unix(),
	}

	// 添加到提案的讨论列表
	proposal.Discussions = append(proposal.Discussions, newDiscussion)

	// 保存更新后的提案
	return pdb.SaveProposal(proposal)
}

// GetDiscussions 获取特定提案的所有讨论
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

// UpdateVote 更新提案的投票结果
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

// DeleteProposal 从数据库中删除提案
func (pdb *ProposalDB) DeleteProposal(proposalID uint64) error {
	key := fmt.Sprintf("proposal:%d", proposalID)
	return pdb.db.Delete([]byte(key), nil)
}

// ListProposals 列出所有的提案
func (pdb *ProposalDB) ListProposals() ([]*ProposalItem, error) {
	var proposals []*ProposalItem
	iter := pdb.db.NewIterator(nil, nil)
	defer iter.Release()

	for iter.Next() {
		key := string(iter.Key())
		// 只处理提案键
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

// GetProposalsByValidator 获取特定验证者的所有提案
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

// ExportProposalsToFile 将全部提案数据导出到指定文本文件
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

// ExportProposalsToJSON 将全部提案数据导出为JSON文件
// 如果文件路径包含目录不存在会自动创建，默认输出到当前目录的proposal-info.json
func (pdb *ProposalDB) ExportProposalsToJSON(filePath string) error {
	if filePath == "" {
		filePath = "proposal-info.json"
	}

	// 确保目录存在
	if dir := filepath.Dir(filePath); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	proposals, err := pdb.ListProposals()
	if err != nil {
		return fmt.Errorf("failed to list proposals: %w", err)
	}

	// 创建格式化JSON数据
	jsonData, err := json.MarshalIndent(proposals, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// 原子写入文件
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	return nil
}

// GetNextProposalID 获取下一个可用的提案ID
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
