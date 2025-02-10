import React, { useEffect, useRef, useState } from 'react';
import { PlusOutlined, QuestionCircleOutlined } from '@ant-design/icons';
import { Flex, FloatButton, Form, Tooltip } from 'antd';
import { ModalForm, ProFormInstance } from '@ant-design/pro-form';
import { ProFormSelect, ProFormText } from '@ant-design/pro-components';
import './style.less'

interface Props {
  onSuccess: () => void;
}

const ApplyProposal: React.FC<Props> = (props) => {
  const { onSuccess } = props;
  const [isOpen, setIsOpen] = useState(false);
  const [form] = Form.useForm<any>();
  const formRef = useRef<ProFormInstance>(null);
  useEffect(() => {
    if (!isOpen) {
      form.resetFields();
    }
  },[isOpen])
  return (
    <>
      <FloatButton type='primary' icon={<PlusOutlined />} onClick={() => setIsOpen(true)} />
      {isOpen &&
        <ModalForm
          form={form}
          formRef={formRef}
          title={<Flex><span>Apply Proposal</span><Tooltip color={'#00000040'} title={'Chose a certain agent to submit Proposal for you'}><QuestionCircleOutlined className={'question'} /></Tooltip></Flex>}
          modalProps={{
            okText: 'Ok',
            cancelText: 'Cancel',
          }}
          width={400}
          open={isOpen}
          onOpenChange={(visible) => {
            setIsOpen(visible);
          }}
          colon={false}
          layout={'horizontal'}
          labelCol={{
            style: {
              width: 60,
            },
          }}
      >
          <ProFormText  placeholder={'Please input'} name="title" label="Title"></ProFormText>
          <ProFormText placeholder={'Please input'} name="content" label="Content"></ProFormText>
          <ProFormSelect placeholder={'Please choose'} name="agent" label="Agent" valueEnum={{
            Alice: 'Alice',
            Bob: 'Bob',
            Charlie: 'Charlie',
            agent4: 'agent 4',
          }}></ProFormSelect>
        </ModalForm>}
    </>
  );
};

export default ApplyProposal;
