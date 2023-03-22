import { RouteLocationNormalizedLoaded, RouteRecordRaw } from 'vue-router';

export default [
  // 首页
  {
    name: 'index',
    path: '',
    component: () => import('@/views/index.vue'),
    props: ({
      query: { searchName, groupId },
    }: RouteLocationNormalizedLoaded) => ({
      searchName,
      groupId,
    }),
    meta: {
      title: '首页',
    },
  },
  // 组件库路由
  {
    path: 'component-lib',
    component: () => import('@/views/component-lib/index.vue'),
    meta: {
      title: '组件库',
    },
  },
  // 节点库路由
  {
    name: 'node-library',
    path: 'node-library',
    component: () => import('@/views/node-library/node-library-manager.vue'),
    meta: {
      title: '本地节点库',
    },
  },
  // job路由
  {
    name: 'job',
    path: 'job',
    component: () => import('@/views/job/job-manager.vue'),
    meta: {
      title: 'Job管理',
    },
  },
  {
    name: 'project-group',
    path: 'project-group',
    component: () => import('@/views/project-group/project-group-manager.vue'),
    meta: {
      title: '分组管理',
    },
    children: [
      {
        name: 'project-group-detail',
        path: 'detail/:id',
        component: () =>
          import('@/views/project-group/project-group-detail.vue'),
        props: ({ params: { id } }: RouteLocationNormalizedLoaded) => ({ id }),
        meta: {
          title: '详情',
        },
      },
    ],
  },
] as RouteRecordRaw[];
