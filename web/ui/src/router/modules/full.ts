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
    path: 'example',
    component: () => import('@/views/project/example.vue'),
  },
  {
    name:'/test-flow',
    path:'/test-flow',
    component:()=> import('@/views/test-flow/index.vue'),
    children:[
    
    ],
  },
  {
    name: 'task-detail',
    path: 'task/detail',
    component: () => import('@/views/task/detail.vue'),
    props: route => ({ testId: route.query.testId,id: route.query.id }), 
    meta: {
      title: 'Task详情',
    },
  },
] as RouteRecordRaw[];
