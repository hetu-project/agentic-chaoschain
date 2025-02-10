import { Flex, Layout, Row, Typography } from 'antd';
import React from 'react';
import { MinusCircleOutlined } from '@ant-design/icons';

interface Props {
  status: 1 | 2 | 3;
}

// 脚手架示例组件
const Status: React.FC<Props> = (props) => {
  const { status } = props;
  return (
    <Flex justify={'space-between'} align={'center'} style={{padding:'10px',background:'gold',borderRadius:'14px',height:'28px',fontSize:'15px',color:'#777575'}} >
      <MinusCircleOutlined style={{marginRight:'4px'}} />
      Failed
    </Flex>
  );
};

export default Status;