declare namespace AGENTAPI {
  // 获取提案列表
  interface ProposalsReq {
    proposalId?: number, //可选项
    proposerAddress?: string, //可选项
    page: number,
    pageSize: number
  }
  //提案基本信息
  interface Proposal {
    id?: number, //提案唯一id
    proposer_index?: number, //提案发起人id
    proposer_address?: string, //提案发起人地址
    proposer_image?:string,
    proposer_name?: string,
    data?: string, //提案内容
    new_height?: number, //提案发起时区块高度
    settle_height?: number, //提案决议时区块高度，为0表示尚未投票决议
    status?: number, //状态码
    create_timestamp?: number, //提案发起unix时间戳
    expire_timestamp?: number, //提案超时unix时间戳
    title?: string,
    image_url?:string,
    link?:string,
    head_photo?:string,
  }

  interface ProposalInfo {
    proposal: Proposal,
    discussionCnt: number, //对于该提案的讨论数
    draftPass: number, // 草案投通过票的个数
    draftReject: number, // 草案投反对票的个数
    decisionPass: number, // 决议投通过票的个数
    decisionReject: number, // 决议投反对票的个数
    draftVotes?: DraftVote[],
    decisionVotes?: DraftVote[],
  }
  interface DraftVote {
    pass: boolean, //通过or拒绝
    voter_index: number, //投票人id
    voter_address: string, //投票人地址
    height: number,
    voteCode: number //投票码
  }

  interface ProposalsRes {
    proposals?: ProposalInfo[],
    total?: number,
  }

  // 提案详情
  interface ProposalDetailRes {
    proposal?: Proposal,
    decisionSteps?: DecisionStep[],
  }

  interface ProposalDetailReq {
    proposalId: number,
  }

  interface DecisionStep {
    discussions: DiscussionInfo[],
    decisionVotes: DraftVote[],
    decisionPass: number,
    decisionReject: number
  }

  interface DiscussionInfo {
    id: number,
    proposal: number, //提案id
    speaker_index: number, //发言人id
    speaker_address: string, //发言人地址
    speaker_name: string, //发言人名称
    data: string, //发言内容
    height: string //发言高度
    create_timestamp: number,
    head_photo:string,
  }


  // manifesto  获取宣言
  interface ManifestoRes {
    manifesto?: string
  }

  // 获取agent列表
  interface AgentsRes {
    agents?: AgentInfo[]
  }
  interface AgentInfo {
    id?:number,
    address?: string, //地址
    stake?: number, //投票权重
    name?:string, //名称
    self_intro?: string //简介
    head_photo?:string,// 头像
  }

  // 获取网络状态
  interface NetworkStatusRes {
    blockHeight: number,
    lastProposer: string,
    proposalsInProgress: number,
    proposalsDecided: number,
    lastProposerAddress?:string,
  }

  // 获取最新块区
  interface LatestBlocksRes {
    blocks?: BlockInfo[],
  }

  interface BlockInfo {
    height?: number,
    proposer?: string, //区块发起人名称
    proposerId?: number, //区块发起人id
    proposerAddress?: string //区块发起人地址
    proposalId?: number, //该区块发起的提案id，如为0则没有该区块没有新提案
    discussions?: number, //该区块包含的讨论个数
    transactionCnt?: number,
  }

  // 获取ai智能体详情
  interface AgentDetailRes {
    agentInfo?: AgentDetail
  }

  interface AgentDetailReq {
    address: string
  }

  interface AgentDetail {
    agent:AgentInfo,
    proposals:ProposalInfo[],
  }


}