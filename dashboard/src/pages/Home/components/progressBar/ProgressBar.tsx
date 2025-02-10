import { Flex } from 'antd';
import React from 'react';
import './index.less'

interface Props {
  passPercent: number;
}

const ProgressBar: React.FC<Props> = (props) => {
  const { passPercent } = props;
  return (
    <Flex justify={'space-between'} align={'center'} style={{height:'40px',width:'100%',position:'relative'}} >
      <div className={'pass-base'} style={{right:`calc(${100-passPercent}% ${passPercent === 0 ?'- 120px': (passPercent === 100 ?'+ 120px': '+ 0px')} )`}}>
        <Flex justify={'space-between'} align={'center'} style={{ width:`100%`,height:'40px' }}>
          <span className={`pass-base-text ${passPercent >= 50 && 'reject-base-text-blue'}`}>Pass</span>
          <span className={'pass-percent-text'}>{`${passPercent}%`}</span>
        </Flex>
      </div>
      <div className={'reject-base'} style={{left:`calc(${passPercent}% ${passPercent === 100 ?'- 120px': (passPercent === 0 ?'+ 120px': '+ 0px')})`}}>
        <Flex justify={'space-between'} align={'center'} style={{ width:`100%`,height:'40px' }}>
          <span className={'reject-percent-text'}>{`${100 - passPercent}%`}</span>
          <span className={`reject-base-text ${passPercent < 50 && 'reject-base-text-red'}`}>Reject</span>
        </Flex>
      </div>
    </Flex>
  );
};

export default ProgressBar;