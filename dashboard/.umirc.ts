import { defineConfig } from '@umijs/max';

export default defineConfig({
  antd: {},
  access: {},
  model: {},
  initialState: {},
  request: {},
  layout: {
    // title: '@umijs/max',
  },
  routes: [
    {
      path: '/',
      redirect: '/home',
    },
    {
      name: 'home',
      path: '/home',
      component: './Home',
    },
    {
      name: 'proposalDetail',
      path: '/proposalDetail/:id',
      component: './Proposal/proposalDetail.tsx',
    },
    {
      name: 'agentDetail',
      path: '/agentDetail/:id',
      component: './Agent/agentDetail.tsx',
    }
  ],
  npmClient: 'yarn',
  proxy: {
    '/api': {
      target: 'http://52.221.224.189:8631', // 你的后端域名
      changeOrigin: true,
      pathRewrite: { '^/api': '' }, // 根据需求决定是否重写路径
    },
  },
});

