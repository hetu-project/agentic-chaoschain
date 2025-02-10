import { Flex } from 'antd';
import React, { useEffect } from 'react';
import { CheckCircleFilled } from '@ant-design/icons';
import './index.less';

interface Props {
  status: 1 | 2 | 3 | 4;
}

interface ItemClass {
  text: string;
  icon?: React.ReactNode | null;
  textColor: string;
  progressColor: string;
}

const BottomProgressBar: React.FC<Props> = (props) => {
  const { status } = props;
  const [itemClasses, setItemClasses] = React.useState<ItemClass[]>([]);
  useEffect(() => {
    //忽略（状态码1）、讨论中（状态码2）、接受（状态码3）、拒绝（状态码4）
    setItemClasses([
      {
        text:'Draft',
        icon: <CheckCircleFilled />,
        textColor: 'text-red',
        progressColor: 'progress-red',
      },
      {
        text:'Discussion',
        icon: null,
        textColor: 'text-yellow',
        progressColor: 'progress-yellow',
      },
      {
        text:'Decision',
        icon: null,
        textColor: 'text-green',
        progressColor: 'progress-green-3',
      }]);
  }, [status]);

  return (
    <Flex justify={'space-between'} vertical align={'center'} style={{ width: `100%` }}>

      <Flex className={'progress-item-base'} style={{ width: '100%' }}>
        {itemClasses?.map((item, index) => <Flex key={`progress-name-${index}`} className={`${item?.textColor}`} justify={'center'} flex={1}>
          <span>{`${item?.text}  `}{item?.icon}</span>
        </Flex>)}
      </Flex>
      <Flex className='progress-bar'>
        {itemClasses?.map((item, index) => <div key={`progress-${index}`} className={`progress-segment ${item?.progressColor}`}></div>)}
      </Flex>
    </Flex>
  );
};

export default BottomProgressBar;