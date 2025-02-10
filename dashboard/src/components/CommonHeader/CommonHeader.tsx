import { Layout, Row, Typography } from 'antd';
import React from 'react';
import './style.less'

interface Props {
  name: string;
  rightNode?: React.ReactNode;
}

// 脚手架示例组件
const CommonHeader: React.FC<Props> = (props) => {
  const { name } = props;
  return (
    <Layout>
      <div className='baseHeader'>
        <span className='title'>{props?.name ?? ''}</span>
        {props?.rightNode ?? ""}
      </div>
    </Layout>
  );
};

export default CommonHeader;