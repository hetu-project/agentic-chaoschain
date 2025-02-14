package agent

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/calehh/hac-app/state"
	hac_types "github.com/calehh/hac-app/types"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtlog "github.com/cometbft/cometbft/libs/log"
	comethttp "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/cometbft/cometbft/store"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
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

type ChainIndexer struct {
	logger        cmtlog.Logger
	Url           string
	Height        int64
	db            *gorm.DB
	cli           *comethttp.HTTP
	eventHandlers map[string]eventHandler
	elizaClients  map[string]Client
	BlockStore    *store.BlockStore
}

func NewChainIndexer(logger cmtlog.Logger, dbPath string, chainUrl string, bs *store.BlockStore) (*ChainIndexer, error) {
	logger.Info("NewChainIndexer", "dbPath", dbPath, "url", chainUrl)
	cli, err := comethttp.New(chainUrl, "/websocket")
	if err != nil {
		return nil, err
	}
	db, err := gorm.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&Grant{}, &Discussion{}, &Proposal{}, &Height{}, &GrantVote{}, &ProposalVote{}, &ValidatorAgent{}).Error; err != nil {
		return nil, err
	}
	h := Height{Id: 1}
	if err = db.First(&h).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if DiscussionRate > 0 {
		DiscussionTrigger = rand.New(rand.NewSource(time.Now().UnixNano())).Intn(DiscussionRate)
	} else {
		DiscussionTrigger = 0
	}

	c := ChainIndexer{
		logger:        logger.With("module", "indexer"),
		Url:           chainUrl,
		Height:        int64(h.Height + 1),
		db:            db,
		cli:           cli,
		eventHandlers: map[string]eventHandler{},
		elizaClients:  make(map[string]Client),
		BlockStore:    bs,
	}
	c.eventHandlers = map[string]eventHandler{
		hac_types.EventGrantType:          c.handleEventGrant,
		hac_types.EventDiscussionType:     c.handleEventDiscussion,
		hac_types.EventSettleProposalType: c.handleEventSettleProposal,
		hac_types.EventProposalType:       c.handleEventProposal,
	}
	return &c, nil
}

type eventHandler func(ctx context.Context, event abci.Event, height int64)

func (c *ChainIndexer) handleEvent(ctx context.Context, event abci.Event, height int64) {
	if h, ok := c.eventHandlers[event.Type]; ok {
		h(ctx, event, height)
	}
}

func (c *ChainIndexer) handleEventGrant(ctx context.Context, event abci.Event, height int64) {
	ev := hac_types.ParseEventGrant(event)
	if ev == nil {
		c.logger.Error("decode event fail", "event", event)
		return
	}
	grant := Grant{
		Id:              ev.Validator,
		Address:         ev.Address,
		Height:          uint64(height),
		Stake:           ev.Amount,
		Proposer:        ev.ProposerIndex,
		ProposerAddress: ev.ProposerAddress,
		Grant:           ev.Grant,
	}
	if err := c.db.Save(&grant).Error; err != nil {
		c.logger.Error("save account fail", "err", err)
	}

	val := ValidatorAgent{
		Id:       ev.Validator,
		Address:  ev.Address,
		Stake:    ev.Amount,
		AgentUrl: ev.AgentUrl,
		Name:     ev.Name,
	}

	cli, err := NewElizaClient(ev.AgentUrl, c.logger)
	if err != nil {
		c.logger.Error("new eliza client fail", "err", err)
	} else {
		hp, err := cli.GetHeadPhoto(ctx)
		if err != nil {
			c.logger.Error("get head photo fail", "err", err)
		}
		val.HeadPhoto = hp
	}

	if err := c.db.Save(&val).Error; err != nil {
		c.logger.Error("save validator fail", "err", err)
	}
}

func (c *ChainIndexer) handleEventDiscussion(ctx context.Context, event abci.Event, height int64) {
	ev := hac_types.DecodeEventDiscussion(event)
	if ev == nil {
		c.logger.Error("decode event fail", "event", event)
		return
	}
	speaker, err := c.getValidatorByAddress(ev.SpeakerAddress)
	if err != nil {
		c.logger.Error("get validator fail", "err", err)
		return
	}
	if speaker.Id == 0 {
		c.logger.Error("speaker not found", "address", ev.SpeakerAddress)
		return
	}
	discusstion := Discussion{
		Id:              0,
		Proposal:        ev.Proposal,
		SpeakerIndex:    ev.Speaker,
		SpeakerAddress:  ev.SpeakerAddress,
		SpeakerName:     speaker.Name,
		Data:            string(ev.Data),
		Height:          uint64(height),
		CreateTimestamp: time.Now().Unix(),
	}
	if err := c.db.Save(&discusstion).Error; err != nil {
		c.logger.Error("save discusstion fail", "err", err)
	}
	err = ElizaCli.AddDiscussion(ctx, ev.Proposal, ev.SpeakerAddress, string(ev.Data))
	if err != nil {
		c.logger.Error("add discussion fail", "err", err)
	}
}

func (c *ChainIndexer) handleEventSettleProposal(ctx context.Context, event abci.Event, height int64) {
	ev := hac_types.DecodeEventSettleProposal(event)
	if ev == nil {
		c.logger.Error("decode event fail", "event", event)
		return
	}
	var proposal Proposal
	if err := c.db.First(&proposal, ev.Proposal).Error; err != nil {
		c.logger.Error("get proposal fail", "err", err)
		return
	}
	proposal.Status = uint64(ev.State)
	proposal.SettleHeight = uint64(height)
	if err := c.db.Save(&proposal).Error; err != nil {
		c.logger.Error("save proposal fail", "err", err)
	}
}

func (c *ChainIndexer) handleEventProposal(ctx context.Context, event abci.Event, height int64) {
	ev := hac_types.DecodeEventProposal(event)
	if ev == nil {
		c.logger.Error("decode event fail", "event", event)
		return
	}
	now := time.Now()
	proposal := Proposal{
		Id:              ev.ProposalIndex,
		ProposerIndex:   ev.Proposer,
		ProposerAddress: ev.ProposerAddress,
		Data:            string(ev.Data),
		NewHeight:       uint64(height),
		Status:          ev.Status,
		Title:           ev.Title,
		Link:            ev.Link,
		ImageUrl:        ev.ImageUrl,
		CreateTimestamp: now.Unix(),
		ExpireTimestamp: now.Add(time.Hour * 24 * 365).Unix(),
	}
	validator, err := c.getValidatorByAddress(ev.ProposerAddress)
	if err != nil {
		c.logger.Error("get validator fail", "err", err)
	}
	if validator.Name == "" {
		validator.Name = "Enigma"
	}
	proposal.ProposerName = validator.Name

	if err := c.db.Save(&proposal).Error; err != nil {
		c.logger.Error("save proposal fail", "err", err)
	}
	err = ElizaCli.AddProposal(ctx, ev.ProposalIndex, ev.ProposerAddress, string(ev.Data))
	if err != nil {
		c.logger.Error("add proposal fail", "err", err)
	}
	comment, err := ElizaCli.CommentPropoal(ctx, ev.ProposalIndex, ev.ProposerAddress)
	if err != nil {
		c.logger.Error("comment proposal fail", "err", err)
	} else {
		c.logger.Info("comment proposal", "comment", comment)
	}
}

func (c *ChainIndexer) handleVote(ctx context.Context, height int64) error {
	res, err := c.cli.Commit(ctx, &height)
	if err != nil {
		c.logger.Error("get Commit fail", "err", err)
		if !c.cli.IsRunning() {
			c.cli.Stop()
			c.cli, err = comethttp.New(c.Url, "/websocket")
			if err != nil {
				c.logger.Error("reconnect fail", "err", err)
				return err
			}
		}
	}
	voteHeight := res.Height
	// new proposal
	newProposel := Proposal{}
	if err := c.db.Where("new_height = ?", voteHeight).First(&newProposel).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}
	}
	if newProposel.Id != 0 {
		for _, v := range res.Commit.Signatures {
			acc, err := c.queryAccount(ctx, 0, v.ValidatorAddress.String())
			if err != nil {
				return err
			}
			if acc == nil {
				return fmt.Errorf("commit sig address not exist address:%s", v.ValidatorAddress.String())
			}
			if err := c.db.Where("height = ? And voter_index = ?", voteHeight, acc.Index).First(&ProposalVote{}).Error; err != nil {
				if err != gorm.ErrRecordNotFound {
					return err
				}
				vote := ProposalVote{
					Proposal:     newProposel.Id,
					VoterIndex:   acc.Index,
					VoterAddress: v.ValidatorAddress.String(),
					Height:       uint64(voteHeight),
					Vote:         uint64(v.VoteCode),
				}
				if err := c.db.Create(&vote).Error; err != nil {
					return err
				}
			}
		}
		return nil
	}
	// settle proposal
	settleProposel := Proposal{}
	if err := c.db.Where("settle_height = ?", voteHeight).First(&settleProposel).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}
	}
	if settleProposel.Id != 0 {
		for _, v := range res.Commit.Signatures {
			acc, err := c.queryAccount(ctx, 0, v.ValidatorAddress.String())
			if err != nil {
				return err
			}
			if acc == nil {
				return fmt.Errorf("commit sig address not exist address:%s", v.ValidatorAddress.String())
			}
			if err := c.db.Where("height = ? And voter_index = ?", voteHeight, acc.Index).First(&ProposalVote{}).Error; err != nil {
				if err != gorm.ErrRecordNotFound {
					return err
				}
				vote := ProposalVote{
					Proposal:     settleProposel.Id,
					VoterIndex:   acc.Index,
					VoterAddress: v.ValidatorAddress.String(),
					Height:       uint64(voteHeight),
					Vote:         uint64(v.VoteCode),
				}
				if err := c.db.Create(&vote).Error; err != nil {
					return err
				}
			}
		}
		return nil
	}
	// grant grant
	grant := Grant{}
	if err := c.db.Where("height = ?", voteHeight).First(&grant).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return err
		}
	}
	if grant.Id != 0 {
		for _, v := range res.Commit.Signatures {
			acc, err := c.queryAccount(ctx, 0, v.ValidatorAddress.String())
			if err != nil {
				return err
			}
			if acc == nil {
				return fmt.Errorf("commit sig address not exist address:%s", v.ValidatorAddress.String())
			}
			if err := c.db.Where("height = ? And voter_index = ?", voteHeight, acc.Index).First(&GrantVote{}).Error; err != nil {
				if err != gorm.ErrRecordNotFound {
					return err
				}
				vote := GrantVote{
					ProposerIndex:   grant.Proposer,
					ProposerAddress: grant.ProposerAddress,
					AccountIndex:    grant.Id,
					AccountAddr:     grant.Address,
					VoterIndex:      acc.Index,
					VoterAddress:    acc.Address(),
					Height:          uint64(voteHeight),
					Vote:            uint64(v.VoteCode),
				}
				if err := c.db.Create(&vote).Error; err != nil {
					return err
				}
			}
		}
		return nil
	}
	return nil
}

func (c *ChainIndexer) Start(ctx context.Context) {
	var err error
	ticker := time.NewTicker(time.Second)
	time.Sleep(10 * time.Second)
	res, err := c.cli.Validators(context.Background(), nil, nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range res.Validators {
		acc, err := c.queryAccount(ctx, 0, v.Address.String())
		if err != nil {
			log.Fatal(err)
		}
		if acc == nil {
			log.Fatal("validator account not exist")
		}

		agent, err := c.getValidatorByAddress(acc.Address())
		if err == nil && agent != nil && agent.Id != 0 {
			continue
		}

		val := ValidatorAgent{
			Id:       acc.Index,
			Address:  acc.Address(),
			Stake:    acc.Stake,
			AgentUrl: acc.AgentUrl,
			Name:     acc.Name,
		}

		cli, err := NewElizaClient(val.AgentUrl, c.logger)
		if err != nil {
			c.logger.Error("new eliza client fail", "err", err)
		} else {
			hp, err := cli.GetHeadPhoto(ctx)
			if err != nil {
				c.logger.Error("get head photo fail", "err", err)
			}
			val.HeadPhoto = hp
		}

		if err := c.db.Save(val).Error; err != nil {
			panic(err)
		}
	}

	go func() {
		for {
			time.Sleep(10 * time.Second)
			c.fillAgentSelfIntro()
		}
	}()

	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if c.cli == nil {
				c.cli, err = comethttp.New(c.Url, "/websocket")
				if err != nil {
					c.logger.Error("connect fail", "err", err)
					continue
				}
			}
			b, err := c.cli.Status(context.TODO())
			if err != nil {
				c.logger.Error("get status fail", "err", err)
				if !c.cli.IsRunning() {
					c.cli.Stop()
					c.cli, err = comethttp.New(c.Url, "/websocket")
					if err != nil {
						c.logger.Error("reconnect fail", "err", err)
						continue
					}
				}
			}
			for b.SyncInfo.LatestBlockHeight > c.Height {
				time.Sleep(time.Millisecond * 100)
				c.logger.Info("indexer syncing", "height", c.Height)
				events, err := c.cli.BlockResults(ctx, &c.Height)
				if err != nil {
					c.logger.Error("get status fail", "err", err)
					if !c.cli.IsRunning() {
						c.cli.Stop()
						c.cli, err = comethttp.New(c.Url, "/websocket")
						if err != nil {
							c.logger.Error("reconnect fail", "err", err)
							continue
						}
					}
				}
				for _, res := range events.TxsResults {
					for _, event := range res.Events {
						c.handleEvent(ctx, event, c.Height)
					}
				}
				err = c.handleVote(ctx, c.Height)
				if err != nil {
					c.logger.Error("handleVote fail", "height", c.Height, "err", err)
					continue
				}
				if err := c.db.Save(Height{
					Id:     1,
					Height: uint64(c.Height),
				}).Error; err != nil {
					c.logger.Error("save height fail", "err", err)
					continue
				}
				// random discuss if latest block height is current height + 1
				if b.SyncInfo.LatestBlockHeight == c.Height+1 {
					c.randomDiscuss()
				}
				c.Height++
			}
		}
	}
}

func (c *ChainIndexer) randomDiscuss() {
	if DiscussionRate == 0 {
		return
	}
	if (c.Height+int64(DiscussionTrigger))%int64(DiscussionRate) != 0 {
		return
	}
	proposals, err := c.getProposalsByStatus(uint64(hac_types.ProposalStatusProcessing), 0, 10)
	if err != nil {
		c.logger.Error("get proposals fail", "err", err)
		return
	}
	if len(proposals) == 0 {
		return
	}
	suitePrs := make([]Proposal, 0)
	for _, p := range proposals {
		_, cnt, err := c.getDiscussionByProposal(p.Id, 0, 1)
		if err == nil && cnt < 15 {
			suitePrs = append(suitePrs, p)
		}
	}
	if len(suitePrs) == 0 {
		return
	}
	randProposal := suitePrs[rand.Intn(len(suitePrs))]
	comment, err := ElizaCli.CommentPropoal(context.Background(), randProposal.Id, randProposal.ProposerAddress)
	if err != nil {
		c.logger.Error("comment proposal fail", "err", err)
		return
	}
	c.logger.Info("comment proposal", "proposal", randProposal.Id, "comment", comment)
}

func (c *ChainIndexer) fillAgentSelfIntro() {
	// find agent where self_intro is ""
	var agents []ValidatorAgent
	err := c.db.Where("self_intro = ?", "").Find(&agents).Error
	if err != nil {
		c.logger.Error("find agent fail", "err", err)
		return
	}
	for _, a := range agents {
		if a.AgentUrl != "" {
			if _, ok := c.elizaClients[a.Address]; !ok {
				client, err := NewElizaClient(a.AgentUrl, c.logger)
				if err != nil {
					c.logger.Error("new eliza client fail", "err", err)
					continue
				}
				c.elizaClients[a.Address] = client
			}
			selfIntro, err := c.elizaClients[a.Address].GetSelfIntro(context.Background())
			if err != nil {
				c.logger.Error("get self intro fail", "err", err)
				continue
			}
			a.SelfIntro = selfIntro
			if err := c.db.Save(&a).Error; err != nil {
				c.logger.Error("save agent fail", "err", err)
				continue
			}
		}
	}
}

func (c *ChainIndexer) queryAccount(ctx context.Context, index uint64, address string) (*state.Account, error) {
	var err error
	var dat []byte
	if len(address) > 0 {
		dat, err = hex.DecodeString(address)
		if err != nil {
			return nil, err
		}
	} else {
		s := fmt.Sprintf("0%x", index)
		if len(s)&1 == 1 {
			s = s[1:]
		}
		dat, _ = hex.DecodeString(s)
	}
	res, err := c.cli.ABCIQuery(ctx, "/accounts/", dat)
	if err != nil {
		c.logger.Error("ABCIQuery fail", "err", err)
		if !c.cli.IsRunning() {
			c.cli.Stop()
			c.cli, err = comethttp.New(c.Url, "/websocket")
			if err != nil {
				c.logger.Error("reconnect fail", "err", err)
				return nil, err
			}
		}
	}
	if res.Response.Code != 0 {
		fmt.Printf("%#v\n", res)
		return nil, errors.New("response code 0")
	}
	var act state.Account
	err = act.UnmarshalJSON(res.Response.Value)
	if err != nil {
		return nil, err
	}
	return &act, err
}

func (c *ChainIndexer) getProposalsByStatus(status uint64, page int, pageSize int) ([]Proposal, error) {
	var proposals []Proposal
	err := c.db.Where("status = ?", status).Order("id desc").Offset(page * pageSize).Limit(pageSize).Find(&proposals).Error
	if err != nil {
		return nil, err
	}
	return proposals, nil
}

func (c *ChainIndexer) getProposalsInProcess() (uint64, error) {
	var total uint64
	err := c.db.Model(&Proposal{}).Where("status = ?", hac_types.ProposalStatusProcessing).Count(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (c *ChainIndexer) getProposalsDecided() (uint64, error) {
	var total uint64
	err := c.db.Model(&Proposal{}).Where("status > ?", hac_types.ProposalStatusProcessing).Count(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (c *ChainIndexer) getProposals(page int, pageSize int) ([]Proposal, uint64, error) {
	var proposals []Proposal
	err := c.db.Order("id desc").Offset(page * pageSize).Limit(pageSize).Find(&proposals).Error
	if err != nil {
		return nil, 0, err
	}
	// get total proposals
	var total uint64
	err = c.db.Model(&Proposal{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	return proposals, total, nil
}

func (c *ChainIndexer) getProposalById(proposalId uint64) (Proposal, error) {
	var proposal Proposal
	err := c.db.Where("id = ?", proposalId).First(&proposal).Error
	if err != nil {
		return Proposal{}, err
	}
	return proposal, nil
}

func (c *ChainIndexer) getProposalsByProposerAddr(proposerAddr string, page int, pageSize int) ([]Proposal, uint64, error) {
	var proposals []Proposal
	err := c.db.Where("proposer_address = ?", proposerAddr).Order("id desc").Offset(page * pageSize).Limit(pageSize).Find(&proposals).Error
	if err != nil {
		return nil, 0, err
	}
	var total uint64
	err = c.db.Model(&Proposal{}).Where("proposer_address = ?", proposerAddr).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	return proposals, total, nil
}

func (c *ChainIndexer) getDiscussionByProposal(proposal uint64, page int, pageSize int) ([]Discussion, uint64, error) {
	var discussions []Discussion
	err := c.db.Where("proposal = ?", proposal).Order("id desc").Offset(page * pageSize).Limit(pageSize).Find(&discussions).Error
	if err != nil {
		return nil, 0, err
	}
	var total uint64
	err = c.db.Model(&Discussion{}).Where("proposal = ?", proposal).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	return discussions, total, nil
}

func (c *ChainIndexer) getDiscussionCntByHeight(height uint64) (uint64, error) {
	var total uint64
	err := c.db.Model(&Discussion{}).Where("height = ?", height).Count(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (c *ChainIndexer) getGrantById(grantId uint64) (Grant, error) {
	var grant Grant
	err := c.db.Where("id = ?", grantId).First(&grant).Error
	if err != nil {
		return Grant{}, err
	}
	return grant, nil
}

func (c *ChainIndexer) getValidators() ([]ValidatorAgent, error) {
	var validators []ValidatorAgent
	err := c.db.Find(&validators).Error
	if err != nil {
		return nil, err
	}
	return validators, nil
}

func (c *ChainIndexer) getValidatorByAddress(address string) (*ValidatorAgent, error) {
	var val ValidatorAgent
	err := c.db.Where("address = ?", address).First(&val).Error
	if err != nil {
		return nil, err
	}
	return &val, nil
}

func (c *ChainIndexer) getGrants(page int, pageSize int) ([]Grant, uint64, error) {
	var grants []Grant
	err := c.db.Order("id desc").Offset(page * pageSize).Limit(pageSize).Find(&grants).Error
	if err != nil {
		return nil, 0, err
	}
	var total uint64
	err = c.db.Model(&Grant{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	return grants, total, nil
}

func (c *ChainIndexer) getProposalByHeight(height uint64) (*Proposal, error) {
	var proposal Proposal
	err := c.db.Where("new_height = ?", height).First(&proposal).Error
	if err != nil {
		return nil, err
	}
	return &proposal, nil
}

func (c *ChainIndexer) getProposalVotesByProposal(proposal uint64, page int, pageSize int) ([]ProposalVote, error) {
	var votes []ProposalVote
	err := c.db.Where("proposal = ?", proposal).Order("id desc").Offset(page * pageSize).Limit(pageSize).Find(&votes).Error
	if err != nil {
		return nil, err
	}
	return votes, nil
}

func (c *ChainIndexer) getGrantVotesByGrant(grant uint64, page int, pageSize int) ([]GrantVote, error) {
	var votes []GrantVote
	err := c.db.Where("account_index = ?", grant).Order("id desc").Offset(page * pageSize).Limit(pageSize).Find(&votes).Error
	if err != nil {
		return nil, err
	}
	return votes, nil
}

func (c *ChainIndexer) getProposalVotesByVoter(voter string, page int, pageSize int) ([]ProposalVote, error) {
	var votes []ProposalVote
	err := c.db.Where("voter_address = ?", voter).Order("id desc").Offset(page * pageSize).Limit(pageSize).Find(&votes).Error
	if err != nil {
		return nil, err
	}
	return votes, nil
}

func (c *ChainIndexer) getGrantVotesByVoter(voter string, page int, pageSize int) ([]GrantVote, error) {
	var votes []GrantVote
	err := c.db.Where("voter_address = ?", voter).Order("id desc").Offset(page * pageSize).Limit(pageSize).Find(&votes).Error
	if err != nil {
		return nil, err
	}
	return votes, nil
}
