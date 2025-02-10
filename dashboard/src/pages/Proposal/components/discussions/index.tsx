import React, { useEffect } from 'react';
import { Divider, Flex, Image, Typography } from 'antd';
import EChartsReact from '@/components/BaseCharts';
import { EChartsOption } from 'echarts';
import './style.less'

interface Props {
  status: number;
  showLine: boolean;
}


const Discussions: React.FC<Props> = (props) => {
  const { status,showLine = false } = props;
  useEffect(() => {
  }, [status]);

  const itemCiaOption: EChartsOption = {
    tooltip: {
      trigger: 'item',
      formatter:(obj ) => {
        return `${'name' in obj ? obj?.name :''}: ${'value' in obj ? obj?.value :''}%`
      }
    },
    series: [
      {
        type: 'pie',
        radius: '70%',
        itemStyle: {
          color: function (params: { dataIndex: number }) {
            const colorList: string[] = ['#5798F7','#E36681' ];
            return colorList[params.dataIndex];
          },
        },
        data: [
          { value: 67, name: `Pass` },
          { value: 33, name: 'Reject' },
        ],
        label: {
          formatter: (obj) => {
            return `${obj?.name} ${obj?.value}%`
          }
        },
      }
    ]
  };

  return (
    <div>
      <Flex className={'discussions-title'}>Discussions</Flex>
      {[1,2,3].map((item, index) => <Flex className={'discussions-item'} vertical key={`discussions-${index}`}>
        <Flex align={'center'} className={'discussions-item-header'}>
          <Image width={80} height={80} preview={false}
                 src={'https://zos.alipayobjects.com/rmsportal/jkjgkEfvpUPVyRjUImniVslZfWPnJuuZ.png'} />
          <span>Alice</span>
        </Flex>
        <Typography.Paragraph className={'discussions-item-content'}>
          Since the late 20th century, Mars has been explored by uncrewed spacecraft and rovers, with the first flyby by the Mariner 4 probe in 1965, the first orbit by the Mars 2 probe in 1971, and the first landing by the Viking 1 probe in 1976.
        </Typography.Paragraph>
      </Flex>)}
      <Flex className={'discussions-title'}>Decision</Flex>
      <Flex>
        <EChartsReact
          option={itemCiaOption}
          style={{ height: 238, width: '380px' }}
        />
      </Flex>
      {showLine && <Divider style={{background:'#000000'}} />}
    </div>
  );
};

export default Discussions;