import { ActionType, PageContainer } from '@ant-design/pro-components';
import {
  Button,
  Card,
  Col,
  Descriptions,
  Divider,
  Flex,
  FloatButton,
  Image,
  Modal,
  Row,
  Space,
  Typography,
} from 'antd';
import CommonHeader from '@/components/CommonHeader/CommonHeader';
import './index.less';
import { PlusOutlined } from '@ant-design/icons';
import Status from '@/pages/Home/components/Status';
import ProgressBar from '@/pages/Home/components/progressBar/ProgressBar';
import styled from 'styled-components';
import React, { useEffect, useRef, useState } from 'react';
import BottomProgressBar from '@/pages/Home/components/bottomProgressBar/BottomProgressBar';
import { history } from 'umi';
import ApplyProposal from '@/pages/Home/components/applyProposal';
import ProList from '@ant-design/pro-list/lib';
import { manifesto, proposals } from '@/services/api/AgentController';

const StyledCard = styled(Card)`
  transition: background-color 0.3s, box-shadow 0.3s;
  background-color: #EFEFEF;

  &:hover {
    background-color: white;
    box-shadow: 0 6px 16px 0 rgba(0, 0, 0, 0.08), 0 3px 6px -4px rgba(0, 0, 0, 0.12),
    0 9px 28px 8px rgba(0, 0, 0, 0.05);
  }
`;

const HomePage: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const [ manifest, setManifest] = useState<string|null>();
  const getManifest = async () => {
    const res = await manifesto();
    setManifest(res?.manifesto || '')
  };
  useEffect(() => {
    getManifest();
  } ,[])


  const manifestoAction = () => {
    Modal.info({
      title: 'Manifesto',
      icon: null,
      content: manifest,
      closable: true,
      footer: null,
    });
  };

  return (
    <PageContainer
      pageHeaderRender={() => <CommonHeader name='HAC' rightNode={<Button size={'large'} type={'link'} style={{
        fontSize: '1.2rem',
        textDecoration: 'underline',
      }} onClick={manifestoAction}>Manifesto</Button>} />}
      ghost
      className='homeContent'
    >
      <div style={{ padding: '16px 114px', background: 'white' }}>
        <Row gutter={[32, 20]}>
          <Col span={12}>
            <Card className={'message-card'} title={'Network Status'}
                  style={{ background: 'rgb(239, 239, 239,0.52)', minHeight: '102px' }}>
              <Descriptions styles={{
                label: { fontSize: '15px', lineHeight: '18px', wordBreak: 'break-all' },
                content: { fontSize: '15px', lineHeight: '18px' },
              }} column={4}>
                <Descriptions.Item label='Block Height'>105</Descriptions.Item>
                <Descriptions.Item label='BLatest Proposer'>Alice</Descriptions.Item>
                <Descriptions.Item label='Proposals In Progress'>2</Descriptions.Item>
                <Descriptions.Item label='Proposals Decided'>34</Descriptions.Item>
              </Descriptions>
            </Card>
          </Col>
          <Col span={12}>
            <Card className={'message-card'} title={'Latest Block'}
                  style={{ background: 'rgb(239, 239, 239,0.52)', minHeight: '102px' }}>
              <Flex style={{ height: '45px', borderRadius: '7px', background: 'rgb(234, 234, 234,1)', padding: '10px' }}
                    gap={16} align={'center'} justify={'space-between'}>
                <span>Block #01</span>
                <Divider type='vertical' />
                <Descriptions style={{ flex: '1' }} column={5}
                              styles={
                                {
                                  label: { fontSize: '15px', lineHeight: '18px', wordBreak: 'break-all' },
                                  content: { fontSize: '15px', lineHeight: '18px' },
                                }
                              }
                >
                  <Descriptions.Item style={{ paddingInlineEnd: '0' }} label='-Height'>105</Descriptions.Item>
                  <Descriptions.Item style={{ paddingInlineEnd: '0' }} label='-Transactions'>1</Descriptions.Item>
                  <Descriptions.Item style={{ paddingInlineEnd: '0' }} label='-Proposer'><a
                    style={{ textDecoration: 'underline' }}>Alice</a></Descriptions.Item>
                  <Descriptions.Item style={{ paddingInlineEnd: '0' }} label='-Proposal'><a
                    style={{ textDecoration: 'underline' }}>34</a></Descriptions.Item>
                  <Descriptions.Item style={{ paddingInlineEnd: '0' }} label='-Discussions'><a
                    style={{ textDecoration: 'underline' }}>3</a></Descriptions.Item>
                </Descriptions>
              </Flex>
            </Card>
          </Col>
        </Row>
        <Row style={{ marginTop: '16px' }} gutter={[32, 16]}>
          {[1, 2, 3, 4, 5, 6, 7, 8].map((item, index) => <Col key={`name-${index}`} span={3}>
            <StyledCard className={'name-card'} hoverable size={'small'} onClick={() => {
              history.push(`/agentDetail/${111}`);
            }}>
              <Flex vertical align={'center'} justify={'center'}>
                <Image height={48} width={48} preview={false}
                       src={'https://zos.alipayobjects.com/rmsportal/jkjgkEfvpUPVyRjUImniVslZfWPnJuuZ.png'} />
                <span>name</span>
              </Flex>
            </StyledCard>
          </Col>)}
        </Row>
        <ProList
          cardProps={{
            bodyStyle: {
              paddingInline: 0,
              paddingBlock: 0,
            },
            ghost: true,
          }}
          itemCardProps={{
            ghost: true,
            bordered:false,
          }}
          ghost={true}
          size={'small'}
          actionRef={actionRef}
          pagination={{
            size: 'small',
            defaultPageSize: 8,
            showSizeChanger: false,
          }}
          grid={{ gutter: 16, column: 3}}
          renderItem={(item) => (
            <Col span={24}>
              <StyledCard className={'base-card msg-card'} style={{marginTop:'16px'}}  hoverable onClick={() => {
                history.push(`/proposalDetail/${111}`);
              }}>
                <Flex vertical justify={'start'} style={{ height: '260px' }}>
                  <Flex style={{ padding: '0 18px' }} justify={'space-between'}>
                    <Flex className={'name-card-title'} align={'center'} justify={'center'}>
                      <Image width={36} height={36} preview={false}
                             src={'https://zos.alipayobjects.com/rmsportal/jkjgkEfvpUPVyRjUImniVslZfWPnJuuZ.png'} />
                      <span>name</span>
                    </Flex>
                    <Status status={1} />
                  </Flex>
                  <Flex style={{ padding: '0 18px' }}>
                    <Typography.Text ellipsis className={'desc'}>
                      Let's Build Spaceship To Mars！Let's Build Spaceship To Mars！Let's Build Spaceship To Mars！
                    </Typography.Text>
                  </Flex>
                  <Flex style={{ padding: '0 18px' }}>
                    <Typography.Text className={'expire'}>
                      # Expire 2025-01-21 23:59:59
                    </Typography.Text>
                  </Flex>
                  <Flex align={'center'} style={{ padding: '0 18px', marginTop: '-30px' }} flex={1}>
                    <ProgressBar passPercent={37.5} />
                  </Flex>
                  <Flex>
                    <BottomProgressBar status={3} />
                  </Flex>
                </Flex>
              </StyledCard>
            </Col>
          )}
          request={async (params) => {
            const res = await proposals(
              {
                page: params?.current || 1,
                pageSize: 10,
              }
            );
            return {
              data: res?.proposals || [],
              success: !!res?.proposals,
              total: res?.total || 0,
            };
            // return {
            //   data: res?.data?.records || [],
            //   success: res?.success || false,
            //   total: res?.data?.total || 0,
            // };
          }}
        ></ProList>
      </div>
      <ApplyProposal onSuccess={() => console.log('success,refresh!')} />
    </PageContainer>
  );
};

export default HomePage;
