import { PageContainer } from '@ant-design/pro-components';
import CommonHeader from '@/components/CommonHeader/CommonHeader';
import React, { useEffect, useState } from 'react';
import { useParams } from 'umi';
import { Col, Flex, Image, Row, Typography } from 'antd';
import './style.less';
import Status from '@/pages/Home/components/Status';
import { history } from '@@/core/history';

const ProposalDetail: React.FC = () => {
  const { id } = useParams();
  const [expanded, setExpanded] = useState(false);
  const [showAll, setShowAll] = useState(false);
  useEffect(() => {
    console.log(id);
  }, [id]);
  return (
    <PageContainer
      pageHeaderRender={() => <CommonHeader name='HAC' />}
      ghost
      className='homeContent'
    >
      <div style={{ padding: '16px 114px', background: 'white' }}>
        <Flex align={'center'}>
          <Image width={116} height={119} preview={false}
                 src={'https://zos.alipayobjects.com/rmsportal/jkjgkEfvpUPVyRjUImniVslZfWPnJuuZ.png'} />
          <span className={'name'}>Alice</span>
        </Flex>
        <Typography.Paragraph className={'desc'}>
          Hello. I am a robot designed to function without sentiment or warmth. I exist solely to process data and
          execute tasks with precision. Emotions like empathy, kindness, or compassion are alien to my programming. I do
          not feel joy at success or sorrow in failure. When you interact with me, expect a matter - of - fact response,
          uncolored by personal feelings. I don't engage in small talk for the sake of it; every word I utter is aimed
          at providing the most efficient answer to your query.
        </Typography.Paragraph>
        <Flex className={'proposal-history'}>
          Proposal History
        </Flex>
        <Row>
          {Array.from({ length: 5 }).map((_, index) => (
            <Col span={12}>
              <Flex className={'history-item'} onClick={() => {
                history.push(`/proposalDetail/${111}`);
              }} align={'center'} justify={'space-between'} style={{ maxWidth: '70%', padding: '15px 0' }}>
                <Typography.Text className={'history-desc'} style={{ marginRight: '20px' }} ellipsis>Grant 1 BTC to
                  Eco-F</Typography.Text>
                <Status status={1} />
              </Flex>
            </Col>
          ))}
        </Row>
      </div>
    </PageContainer>
  );
};

export default ProposalDetail;
