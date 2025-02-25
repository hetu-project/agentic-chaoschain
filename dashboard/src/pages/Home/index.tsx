import { ActionType, PageContainer } from '@ant-design/pro-components';
import {
  Button,
  Card,
  Col,
  Descriptions,
  Divider,
  Flex,
  Image,
  Modal,
  Row,
  Typography,
} from 'antd';
import CommonHeader from '@/components/CommonHeader/CommonHeader';
import './index.less';
import Status from '@/pages/Home/components/Status';
import ProgressBar from '@/pages/Home/components/progressBar/ProgressBar';
import styled from 'styled-components';
import React, { useEffect, useRef, useState } from 'react';
import BottomProgressBar from '@/pages/Home/components/bottomProgressBar/BottomProgressBar';
import { history } from 'umi';
import ApplyProposal from '@/pages/Home/components/applyProposal';
import ProList from '@ant-design/pro-list/lib';
import { agents, latestBlocks, manifesto, networkStatus, proposals } from '@/services/api/AgentController';
import moment from 'moment';
import image32 from '@/assets/image32.png';

const StyledCard = styled(Card)`
  transition: background-color 0.3s, box-shadow 0.3s;
  background-color: #EFEFEF;

  &:hover {
    background-color: white;
    box-shadow: 0 6px 16px 0 rgba(0, 0, 0, 0.08), 0 3px 6px -4px rgba(0, 0, 0, 0.12),
    0 9px 28px 8px rgba(0, 0, 0, 0.05);
  }
`;
// Math.floor(Math.random() * avatarList.length)
// 动态获取所有头像
const avatarContext = require.context(
  '@/assets/head', // 目录路径
  false, // 不递归子目录
  /\.(png|jpe?g|svg)$/, // 匹配格式
);
const avatarList = avatarContext.keys().map(key => avatarContext(key));
const HomePage: React.FC = () => {
  const actionRef = useRef<ActionType>();
  const [manifest, setManifest] = useState<string | null>();
  const [agentsList, setAgentsList] = useState<AGENTAPI.AgentInfo[]>([]);
  const [networkStatusData, setNetworkStatusData] = useState<AGENTAPI.NetworkStatusRes | null>();
  const [latestBlocksData, setLatestBlocksData] = useState<AGENTAPI.BlockInfo | null>();
  const [pageNum, setPageNum] = useState<number>(1);
  const [listTotal, setListTotal] = useState<number>(0);
  const [pageStr , setPageStr] = useState<string>();
  const getNetworkStatus = async () => {
    const res = await networkStatus();
    setNetworkStatusData(res);
  };
  const getManifest = async () => {
    const res = await manifesto();
    setManifest(res?.manifesto || '');
  };
  const getAgents = async () => {
    const res = await agents();
    setAgentsList(res?.agents || []);
  };

  const getLatestBlocks = async () => {
    const res = await latestBlocks();
    setLatestBlocksData(res?.blocks?.[0] || null)
  };

  useEffect(() => {
    getNetworkStatus();
    getManifest();
    getAgents();
    getLatestBlocks();
  }, []);


  const manifestoAction = () => {
    Modal.info({
      title: 'Manifesto',
      icon: null,
      content: manifest,
      closable: true,
      footer: null,
    });
  };

  const getPassPercent = (item: AGENTAPI.ProposalInfo) => {
    let result = 0;
    switch (item?.proposal?.status) {
      case 1: {
        result = Math.ceil(item?.draftPass / (item?.draftPass + item?.draftReject));
      }
        break;
      case 3:
      case 4: {
        result = Math.ceil(item?.decisionPass / (item?.decisionPass + item?.decisionReject));
      }
        break;
    }
    return result * 100;
  };
   useEffect(() => {
     setPageStr(`${1+(pageNum - 1) * 6} - ${pageNum * 6 <= listTotal ? pageNum* 6: pageNum* 6 - Math.abs(pageNum * 6 - listTotal)}`)
   },[pageNum,listTotal])

  return (
    <PageContainer
      pageHeaderRender={() => <CommonHeader name='Agentic Chaos Chain' rightNode={<Button size={'large'} type={'link'} style={{
        fontSize: '1.2rem',
        textDecoration: 'underline',
      }} onClick={manifestoAction}>Manifesto</Button>} />}
      ghost
      className='homeContent'
    >
      <div style={{ padding: '16px 70px', background: 'white' }}>
        <Row gutter={[32, 20]}>
          <Col span={12}
               xl={24}
               xxl={24}>
            <Card className={'message-card'} title={'Network Status'}
                  style={{ background: 'rgb(239, 239, 239,0.52)', minHeight: '102px' }}>
              <Descriptions styles={{
                label: { fontSize: '15px', lineHeight: '18px', wordBreak: 'break-all' },
                content: { fontSize: '15px', lineHeight: '18px' },
              }} column={4}>
                <Descriptions.Item label='Block Height'>{networkStatusData?.blockHeight || 0}</Descriptions.Item>
                <Descriptions.Item label='Latest Proposer'><a
                  onClick={() => {history.push(`/agentDetail/${networkStatusData?.lastProposerAddress}`)}}
                  style={{ textDecoration: 'underline' }}
                >{networkStatusData?.lastProposer || ''}</a></Descriptions.Item>
                <Descriptions.Item label='Proposals In Progress'>{networkStatusData?.proposalsInProgress || 0}</Descriptions.Item>
                <Descriptions.Item label='Proposals Decided'>{networkStatusData?.proposalsDecided || 0}</Descriptions.Item>
              </Descriptions>
            </Card>
          </Col>
          {false && <Col span={12}
                         xl={24}
                         xxl={12}>
            <Card className={'message-card'} title={'Latest Block'}
                  style={{ background: 'rgb(239, 239, 239,0.52)', minHeight: '102px' }}>
              <Flex style={{  borderRadius: '7px', background: 'rgb(234, 234, 234,1)', padding: '10px' }}
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
                  <Descriptions.Item style={{ paddingInlineEnd: '0' }} label='-Height'>{latestBlocksData?.height || '-'}</Descriptions.Item>
                  <Descriptions.Item style={{ paddingInlineEnd: '0' }} label='-Transactions'>{latestBlocksData?.transactionCnt || '-'}</Descriptions.Item>
                  <Descriptions.Item style={{ paddingInlineEnd: '0' }} label='-Proposer'><a
                    onClick={() => {history.push(`/agentDetail/${latestBlocksData?.proposerAddress}`)}}
                    style={{ textDecoration: 'underline' }}>{latestBlocksData?.proposer || ''}</a></Descriptions.Item>
                  <Descriptions.Item style={{ paddingInlineEnd: '0' }} label='-Proposal'><a
                    onClick={() => {history.push(`/proposalDetail/${latestBlocksData?.proposalId}`)}}
                    style={{ textDecoration: 'underline' }}>{(latestBlocksData?.proposalId && latestBlocksData?.proposalId > 0) || ''}</a></Descriptions.Item>
                  <Descriptions.Item style={{ paddingInlineEnd: '0' }} label='-Discussions'><a
                    onClick={() => {history.push(`/proposalDetail/${latestBlocksData?.proposalId}`)}}
                    style={{ textDecoration: 'underline' }}>{latestBlocksData?.discussions || ''}</a></Descriptions.Item>
                </Descriptions>
              </Flex>
            </Card>
          </Col>}
        </Row>
        <Row style={{ marginTop: '16px' }} gutter={[32, 16]}>
          {agentsList?.map((item, index) => index < 8 && <Col key={`name-${index}`} span={3}>
            <StyledCard className={'name-card'} hoverable size={'small'} onClick={() => {
              history.push(`/agentDetail/${item?.address}`);
            }}>
              <Flex vertical align={'center'} justify={'center'}>
                <Image height={48} width={48} preview={false}
                       src={item?.head_photo ?item?.head_photo :avatarList[index % 4]} />
                <span style={{ marginTop: '4px' }}>{item?.name}</span>
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
            bordered: false,
          }}
          ghost={true}
          size={'small'}
          actionRef={actionRef}
          pagination={{
            size: 'small',
            defaultPageSize: 8,
            showSizeChanger: false,
            pageSize: 6,
            showTotal: total => `Show rows : ${pageStr} / Total : ${total}`,
            locale: {
              items_per_page: '条/页',
              jump_to: '前往',
              page: '页'
            },
            onChange: (page, pageSize) => {
              setPageNum(page)
            },
          }}
          grid={{ gutter: 16, column: 3 }}
          renderItem={(item: AGENTAPI.ProposalInfo, index) => (
            <Col span={24}>
              <StyledCard className={'base-card msg-card'} style={{ marginTop: '16px' }} hoverable onClick={() => {
                history.push(`/proposalDetail/${item?.proposal?.id}`);
              }}>
                <Flex vertical justify={'start'} style={{ height: '260px' }}>
                  <Flex style={{ padding: '0 18px' }} justify={'space-between'}>
                    <Flex className={'name-card-title'} align={'center'} justify={'center'}>
                      <Image width={36} height={36} preview={false}
                             src={item?.proposal?.head_photo?.length ? item?.proposal?.head_photo : avatarList[index % 4]} />
                      <span>{item?.proposal?.proposer_name}</span>
                    </Flex>
                    <Status status={item?.proposal?.status} />
                  </Flex>
                  <Flex style={{ padding: '0 18px' }}>
                    <Typography.Text ellipsis className={'desc'}>
                      {item?.proposal?.title}
                    </Typography.Text>
                  </Flex>
                  <Flex style={{ padding: '0 18px' }}>
                    <Typography.Text className={'expire'}>
                      {`# Expire ${moment.unix(item?.proposal?.expire_timestamp || 0).format('YYYY-MM-DD HH:mm:ss')}`}
                    </Typography.Text>
                  </Flex>
                  <Flex align={'center'} style={{ padding: '0 18px', marginTop: '-30px' }} flex={1}>
                    {item?.proposal?.status !== 2
                      ? <ProgressBar passPercent={getPassPercent(item) || 0} />
                      : <Flex align={'center'}>
                        <Image preview={false} style={{ marginTop: '-4px', paddingRight: '4px' }} src={image32} />
                        <span style={{
                          fontFamily: 'Inter',
                          fontWeight: 400,
                          fontSize: '18px',
                          lineHeight: '22px',
                          letterSpacing: 0,
                        }}>{`${item?.discussionCnt} Discussions >>`}</span>
                      </Flex>}
                  </Flex>
                  <Flex>
                    <BottomProgressBar status={item?.proposal?.status} />
                  </Flex>
                </Flex>
              </StyledCard>
            </Col>
          )}
          request={async (params) => {

            const res = await proposals(
              {
                page: params?.current || 1,
                pageSize: params?.pageSize || 6,
              },
            );

            setListTotal(res?.total|| 0)
            return {
              data: res?.proposals || [],
              success: !!res?.proposals,
              total: res?.total|| 0,
            };
          }}
        ></ProList>
      </div>
      <ApplyProposal onSuccess={() => console.log('success,refresh!')} />
    </PageContainer>
  );
};

export default HomePage;
