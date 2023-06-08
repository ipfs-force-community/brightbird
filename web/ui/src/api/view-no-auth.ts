import { restProxy } from '@/api/index';
import {
  ITaskExecutionRecordVo,
  ITaskParamVo,
} from '@/api/dto/workflow-execution-record';
import {
  IProcessTemplateVo,
  ITestFlowDetail,
  IProjectQueryingDto,
  IChangeTestflowGroupDto,
  ITestFlowIdVo,
  IGetTestFlowParam,
  ICountTestFlowParam,
} from '@/api/dto/testflow';
import { IPageVo } from '@/api/dto/common';
import { ITriggerEventVo, ITriggerWebhookVo } from '@/api/dto/trigger';
import { ITestflowGroupVo } from '@/api/dto/testflow-group';
import { IProjectCacheVo, INodeCacheVo } from '@/api/dto/cache';

export const baseUrl = {
  projectGroup: '/group',
  workflow: '/view/workflow_instances',
  asyncTasks: '/view/async_task_instances',
  task: '/view/task_instance',
  log: '/logs',
  dsl: '/view/workflow',
  processLog: '/view/logs/workflow',
  plugin: '/plugin',
  testflow: '/testflow',
  parameter: '/view/parameters',
  triggerEvent: '/view/trigger_events',
  trigger: '/view/trigger',
  version: 'https://jianmu.dev/versions/ci',
  cache: '/view/caches',
};
const hubUrl = import.meta.env.VITE_JIANMU_HUB_API_BASE_URL;
const baseHubUrl = {
  processTemplate: '/hub/view/workflow_templates',
  node_definitions: '/hub/view/node_definitions',
};

/**
 * 获取项目组列表
 */
export function listTestflowGroup(): Promise<ITestflowGroupVo[]> {
  return restProxy<ITestflowGroupVo[]>({
    url: `${baseUrl.projectGroup}/list`,
    method: 'get',
  });
}

/**
 * 查询项目组详情
 */
export function getProjectGroup(groupId: string): Promise<ITestflowGroupVo> {
  return restProxy({
    url: `${baseUrl.projectGroup}/${groupId}`,
    method: 'get',
  });
}

/**
 * 查询测试流
 * @param dto
 */
export function queryTestFlow(dto: IProjectQueryingDto): Promise<IPageVo<ITestFlowDetail>> {
  return restProxy({
    url: `${baseUrl.testflow}/list`,
    method: 'get',
    payload: dto,
  });
}

/**
 * 查询测试流
 * @param dto
 */
export function countTestFlow(dto: ICountTestFlowParam): Promise<number> {
  return restProxy({
    url: `${baseUrl.testflow}/count`,
    method: 'get',
    payload: dto,
  });
}

/**
 * 项目组添加项目
 */
export function changeTestflowGroup(dto: IChangeTestflowGroupDto): Promise<void> {
  return restProxy({
    url: `${baseUrl.testflow}/changegroup`,
    method: 'post',
    auth: true,
    payload: dto,
  });
}


/**
 * 获取流程模版
 * @param dto
 */
export function getProcessTemplate(dto: number): Promise<IProcessTemplateVo> {
  return restProxy({
    url: `${hubUrl}${baseHubUrl.processTemplate}/${dto}`,
    method: 'get',
  });
}

/**
 * 获取测试流详情
 * @param params
 */
export function fetchTestFlowDetail(params: IGetTestFlowParam): Promise<ITestFlowDetail> {
  return restProxy({
    url: `${baseUrl.testflow}`,
    method: 'get',
    payload: params,
  });
}

export function deleteTestFlow(projectId: string): Promise<void> {
  return restProxy({
    url: `${baseUrl.testflow}/${projectId}`,
    method: 'delete',
  });
}

/**
 * 保存测试流
 * @param dto
 */
export async function saveTestFlow(dto: ITestFlowDetail): Promise<ITestFlowIdVo> {
  const res = await restProxy({
    url: `${baseUrl.testflow}`,
    method: 'post',
    payload: dto,
  });

  return dto.id ? {
    id: dto.id,
  } : res;
}

/**
 * 获取任务实例列表
 * @param businessId
 */
export function listTaskInstance(businessId: string): Promise<ITaskExecutionRecordVo[]> {
  return restProxy<ITaskExecutionRecordVo[]>({
    url: `${baseUrl.task}/${businessId}`,
    method: 'get',
  });
}

/**
 * 获取任务参数列表
 * @param taskId
 */
export function listTaskParam(taskId: string): Promise<ITaskParamVo[]> {
  return restProxy<ITaskParamVo[]>({
    url: `${baseUrl.task}/${taskId}/parameters`,
    method: 'get',
  });
}

/**
 * 获取参数类型列表
 */
export function fetchParameterType(): Promise<string[]> {
  return restProxy({
    url: `${baseUrl.parameter}/types`,
    method: 'get',
  });
}

/**
 * 获取触发器事件
 * @param triggerId
 */
export function fetchTriggerEvent(triggerId: string): Promise<ITriggerEventVo> {
  return restProxy<ITriggerEventVo>({
    url: `${baseUrl.triggerEvent}/${triggerId}`,
    method: 'get',
  });
}

/**
 * 获取触发器webhook
 * @param projectId
 */
export function fetchTriggerWebhook(projectId: string): Promise<ITriggerWebhookVo> {
  return restProxy({
    url: `${baseUrl.trigger}/webhook/${projectId}`,
    method: 'get',
  });
}

/**
 * 获取版本列表
 */
export function fetchVersion(): Promise<string> {
 return restProxy({
    url: `${baseUrl.version}`,
    method: 'get',
    timeout: 1000,
  });
}

/**
 * 获取项目缓存
 */

export function fetchProjectCache(workflowRef: string) {
  return restProxy<IProjectCacheVo>({
    url: `${baseUrl.cache}/${workflowRef}`,
    method: 'post',
  });
}

/**
 * 获取节点缓存
 */
export function fetchNodeCache(asyncTaskId: string) {
  return restProxy<INodeCacheVo[]>({
    url: `${baseUrl.cache}/async_task_instances/${asyncTaskId}`,
    method: 'post',
  });
}

/**
 * 获取task中的Pod
 * @param taskId
 */
export function listAllPod(taskId: string): Promise<string[]> {
  return restProxy({
    url: `${baseUrl.log}/pods/${taskId}`,
    method: 'get',
  });
}

/**
 * 获取Pod中的Log
 * @param podName
 */
export function getPodLog(podName: string): Promise<string[]> {
  return restProxy({
    url: `${baseUrl.log}/${podName}`,
    method: 'get',
  });
}