import { RouteLocationNormalizedLoaded, RouteRecordRaw } from 'vue-router';

export default [
  {
    name: 'create-pipeline',
    path: 'project/pipeline-editor',
    component: () => import('@/views/project/pipeline-editor.vue'),
    meta: {
      title: '创建管道项目',
    },
  },
  {
    name: 'update-pipeline',
    path: 'project/pipeline-editor/:id',
    component: () => import('@/views/project/pipeline-editor.vue'),
    meta: {
      title: '编辑管道项目',
    },
    props: ({ params: { id } }: RouteLocationNormalizedLoaded) => ({ id }),
  },
  {
    name: 'task-detail',
    path: 'task/detail',
    component: () => import('@/views/task/detail.vue'),
    props: ({
      query: { testId },
    }: RouteLocationNormalizedLoaded) => ({
      testId,
    }),
    meta: {
      title: '执行记录',
    },
  },
  {
    path: 'example',
    component: () => import('@/views/project/example.vue'),
  },
] as RouteRecordRaw[];
