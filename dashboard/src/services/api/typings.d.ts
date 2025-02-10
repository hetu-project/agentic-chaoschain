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
    data?: string, //提案内容
    new_height?: number, //提案发起时区块高度
    settle_height?: number, //提案决议时区块高度，为0表示尚未投票决议
    status?: number, //状态码
    create_timestamp?: number, //提案发起unix时间戳
    expire_timestamp?: number, //提案超时unix时间戳
  }
  //
  interface ProposalInfo {
    proposal: Proposal,
    discussionCnt: number, //对于该提案的讨论数
    draftPass: number, // 草案投通过票的个数
    draftReject: number, // 草案投反对票的个数
    decisionPass: number, // 决议投通过票的个数
    decisionReject: number, // 决议投反对票的个数
  }


  interface ProposalsRes {
    proposals?: ProposalInfo[],
    total?: number,
  }

  // manifesto  获取宣言
  interface ManifestoRes {
    manifesto?: string
  }

}