import { PageContainer } from '@ant-design/pro-components';
import CommonHeader from '@/components/CommonHeader/CommonHeader';
import React, { useEffect, useState } from 'react';
import { useParams } from 'umi';
import { Col, Flex, Image, Row, Typography } from 'antd';
import './style.less';
import Status from '@/pages/Home/components/Status';
import { history } from '@@/core/history';
import { agentDetail } from '@/services/api/AgentController';

const avatarContext = require.context(
  '@/assets/head', // 目录路径
  false, // 不递归子目录
  /\.(png|jpe?g|svg)$/, // 匹配格式
);
const avatarList = avatarContext.keys().map(key => avatarContext(key));

const ProposalDetail: React.FC = () => {
  const { id } = useParams();
  const [agentDetailData, setAgentDetailData] = useState<AGENTAPI.AgentDetail>();
  const [loading, setLoading] = useState(false);
  const getAgentDetail = async () => {
    setLoading(true)
    const res = await agentDetail({
      address: id || '',
    });
    setAgentDetailData(res?.agentInfo)
    setLoading(false)
  };
  useEffect(() => {
    getAgentDetail()
  }, [id]);
  return (
    <PageContainer
      pageHeaderRender={() => <CommonHeader name='Agentic Chaos Chain' />}
      ghost
      className='homeContent'
      loading={loading}
    >
      <div style={{ padding: '16px 70px', background: 'white' }}>
        <Flex style={{marginTop:'50px'}} align={'center'}>
          <Image width={100} height={100} preview={false}
                 src={agentDetailData?.agent?.head_photo?agentDetailData?.agent?.head_photo:avatarList[Math.floor(Math.random() * avatarList.length)]} />
          <span className={'name'}>{agentDetailData?.agent?.name}</span>
        </Flex>
        <Typography.Paragraph className={'desc'}>
          {agentDetailData?.agent?.self_intro}
        </Typography.Paragraph>
        <Flex className={'proposal-history'}>
          Proposal History
        </Flex>
        <Row>
          {agentDetailData?.proposals.map((item, index) => (
            <Col key={`agent-detail-${index}`} span={12}>
              <Flex className={'history-item'} onClick={() => {
                history.push(`/proposalDetail/${item?.proposal?.id}`);
              }} align={'center'} justify={'space-between'} style={{ maxWidth: '70%', padding: '15px 0' }}>
                <Typography.Text className={'history-desc'} style={{ marginRight: '20px' }} ellipsis>{item?.proposal?.title}</Typography.Text>
                <Status status={item?.proposal?.status} />
              </Flex>
            </Col>
          ))}
        </Row>
      </div>
    </PageContainer>
  );
};

export default ProposalDetail;
