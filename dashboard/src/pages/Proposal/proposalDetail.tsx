import { PageContainer } from '@ant-design/pro-components';
import CommonHeader from '@/components/CommonHeader/CommonHeader';
import React, { useEffect, useState } from 'react';
import { useParams } from 'umi';
import { Flex, Image, Typography } from 'antd';
import './index.less'
import { DoubleRightOutlined, DownOutlined } from '@ant-design/icons';
import BottomProgressBar from '@/pages/Home/components/bottomProgressBar/BottomProgressBar';
import Status from '@/pages/Home/components/Status';
import Discussions from '@/pages/Proposal/components/discussions';
const ProposalDetail: React.FC = () => {
  const { id } = useParams();
  const [expanded, setExpanded] = useState(false);
  const [showAll, setShowAll] = useState(false);
  useEffect(() => {
    console.log(id);
  },[id])
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
        <Flex align={'center'} justify={'space-between'}>
          <Flex vertical flex={1} style={{marginRight:'20px'}}>
            <Flex style={{ padding: '0' }}>
              <Typography.Text ellipsis className={'desc'}>
                Let's Build Spaceship To Mars！Let's Build Spaceship To Mars！Let's Build Spaceship To Mars！
              </Typography.Text>
            </Flex>
            <Flex style={{ padding: '0' }}>
              <Typography.Text className={'expire'}>
                # Expire 2025-01-21 23:59:59
              </Typography.Text>
            </Flex>
          </Flex>
          <Status status={1} />
        </Flex>
        <Flex style={{ paddingTop: '10px' }}>
          <Typography.Paragraph
            ellipsis={{
              rows:2,
              expandable: 'collapsible',
              expanded,
              onExpand: (_, info) => setExpanded(info.expanded),
              symbol: expanded ? '' : <span className={'expanded-text'}>view full text <DownOutlined /></span>,
            }}
            className={'content'}>
            I have heard of your paintings too, well enough; God has given you one face, and you make yourselves another: you jig and amble, and you lisp, and nickname God's creatures, and make your wantonness your ignorance. Go to, I'll no more on't; it hath made me mad. I say, we will have no more marriages: those that are married already, all but one, shall live; the rest shall keep as they are. To a nunnery, go.
          </Typography.Paragraph>
        </Flex>
        <Flex className={'progress-bar-base'} style={{width:'550px'}}>
          <BottomProgressBar status={3} />
        </Flex>
        <Flex vertical>
          {[1,3,4]?.map((item,index)=> {
            return showAll  ? <Discussions showLine={showAll && index < [1,2,3].length-1} status={index} /> : (index === 0 ? <Discussions showLine={showAll && index < [1,2,3].length-1} status={index} /> : null);
          })}
        </Flex>
        {([1,3,4].length > 1 && !showAll) && <Flex justify={'center'} align={'center'}>
          <a className={'show-all'} onClick={() => setShowAll(true)}>View Previous <DoubleRightOutlined style={{ transform: 'rotate(90deg)' }} /></a>
        </Flex>}
      </div>
    </PageContainer>
  );
};

export default ProposalDetail;
