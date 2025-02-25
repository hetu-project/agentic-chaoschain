import { PageContainer } from '@ant-design/pro-components';
import CommonHeader from '@/components/CommonHeader/CommonHeader';
import React, { useEffect, useLayoutEffect, useRef, useState } from 'react';
import { useParams } from 'umi';
import { Flex, Image, Typography } from 'antd';
import './index.less';
import { DoubleRightOutlined, DownOutlined } from '@ant-design/icons';
import BottomProgressBar from '@/pages/Home/components/bottomProgressBar/BottomProgressBar';
import Status from '@/pages/Home/components/Status';
import Discussions from '@/pages/Proposal/components/discussions';
import { proposalDetail } from '@/services/api/AgentController';
import moment from 'moment/moment';
const Markdown = require('react-remarkable');

// @ts-ignore
const avatarContext = require.context(
  '@/assets/head', // 目录路径
  false, // 不递归子目录
  /\.(png|jpe?g|svg)$/, // 匹配格式
);
// @ts-ignore
const avatarList = avatarContext.keys().map(key => avatarContext(key));


const ProposalDetail: React.FC = () => {
  const { id } = useParams();
  const [showAll, setShowAll] = useState(false);
  const [resultData, setResultData] = useState<AGENTAPI.ProposalDetailRes>();
  const [isEllipsis, setIsEllipsis] = useState(false);
  const [head, setHead] = useState<React.ReactNode | undefined>();
  const getProposalDetail = async () => {
    const res = await proposalDetail({
      proposalId: Number(id || 0),
    });
    setResultData(res);
    setHead(res?.proposal?.head_photo?res?.proposal?.head_photo:avatarList[Math.floor(Math.random() * avatarList.length)] || undefined);
  };
  useEffect(() => {
    getProposalDetail();

  }, [id]);


  useEffect(() => {
    console.log('isEllipsis', isEllipsis);
  }, [isEllipsis]);

  return (
    <PageContainer
      pageHeaderRender={() => <CommonHeader name='Agentic Chaos Chain' />}
      ghost
      className='homeContent'
    >
      <div style={{ padding: '16px 70px', background: 'white' }}>
        <Flex style={{ marginTop: '50px' }} align={'center'}>
          <Image width={100} height={100} preview={false}
            // @ts-ignore
                 src={head} />
          <span className={'name'}>{resultData?.proposal?.proposer_name}</span>
        </Flex>
        <Flex align={'center'} justify={'space-between'}>
          <Flex vertical flex={1} style={{ marginRight: '20px' }}>
            <Flex style={{ padding: '0' }}>
              <Typography.Text ellipsis className={'desc'}>
                {resultData?.proposal?.title}
              </Typography.Text>
            </Flex>
            <Flex style={{ padding: '0' }}>
              <Typography.Text className={'expire'}>
                {`# Expire ${moment.unix(resultData?.proposal?.expire_timestamp || 0).format('YYYY-MM-DD HH:mm:ss')}`}
              </Typography.Text>
            </Flex>
          </Flex>
          <Status status={resultData?.proposal?.status} />
        </Flex>
        <Flex className={'markdown-base'} style={{ paddingTop: '10px',maxWidth: '100%' }}>
          <Markdown className={'content-markdown'}>
            {
              resultData?.proposal?.data
            }
          </Markdown>
          {/*<Typography.Paragraph*/}
          {/*  ellipsis={{*/}
          {/*    rows: 3,*/}
          {/*    expandable: 'collapsible',*/}
          {/*    expanded,*/}
          {/*    onExpand: (_, info) => {*/}
          {/*      setExpanded(info.expanded);*/}
          {/*    },*/}
          {/*    symbol: !expanded ? (*/}
          {/*      <span className='expanded-text'>*/}
          {/*  view full text <DownOutlined style={{ fontSize: 10 }} />*/}
          {/*</span>*/}
          {/*    ) : null,*/}
          {/*  }}*/}
          {/*  className={'content'}*/}
          {/*>*/}
          {/*  {resultData?.proposal?.data}*/}
          {/*</Typography.Paragraph>*/}
        </Flex>
        <Flex className={'progress-bar-base'} style={{ width: '550px' }}>
          <BottomProgressBar status={resultData?.proposal?.status} />
        </Flex>
        <Flex vertical>
          {resultData?.decisionSteps?.map((item, index) => {
            return showAll ? <Discussions key={`discussions-${index}`} decisionStep={item}
                                          showLine={(showAll && resultData?.decisionSteps && index < resultData?.decisionSteps?.length - 1) || false}
                                          status={index} />
              : (index === 0 ? <Discussions key={`discussions-${index}`} decisionStep={item}
                                            showLine={showAll && resultData?.decisionSteps && index < resultData?.decisionSteps?.length - 1}
                                            status={index} /> : null);
          })}
        </Flex>
        {(resultData?.decisionSteps && resultData?.decisionSteps?.length > 1 && !showAll) &&
          <Flex justify={'center'} align={'center'}>
            <a className={'show-all'} onClick={() => setShowAll(true)}>View Previous <DoubleRightOutlined
              style={{ transform: 'rotate(90deg)' }} /></a>
          </Flex>}
      </div>
    </PageContainer>
  );
};

export default ProposalDetail;
