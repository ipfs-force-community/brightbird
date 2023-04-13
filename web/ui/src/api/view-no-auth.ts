import { restProxy } from '@/api/index';
import {
  ITaskExecutionRecordVo,
  ITaskParamVo,
} from '@/api/dto/workflow-execution-record';
import {
  INodeVo, IProcessTemplateVo, ITestFlowDetail,
  IProjectQueryingDto, IWorkflowVo, ITestFlowIdVo, IGetTestFlowParam
} from '@/api/dto/project';
import { IPageVo } from '@/api/dto/common';
import { ITriggerEventVo, ITriggerWebhookVo } from '@/api/dto/trigger';
import { IProjectGroupVo } from '@/api/dto/project-group';
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
export function listProjectGroup(): Promise<IProjectGroupVo[]> {
  return restProxy<IProjectGroupVo[]>({
    url: `${baseUrl.projectGroup}/list`,
    method: 'get',
  });
}

/**
 * 查询项目组详情
 */
export function getProjectGroup(groupId: string): Promise<IProjectGroupVo> {
  return restProxy({
    url: `${baseUrl.projectGroup}/${groupId}`,
    method: 'get',
  });
}

/**
 * 查询项目
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
 * 获取组中测试流的数量
 * @param dto
 */
export function countTestFlows(groupId: string): Promise<number> {
  return restProxy({
    url: `${baseUrl.testflow}/count/${groupId}`,
    method: 'get',
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
 * 获取项目详情
 * @param projectId
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
 * 保存项目
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
 * 获取deploy插件
 */
export function fetchDeployPlugins(): Promise<INodeVo[]> {
  return restProxy<INodeVo[]>({
    url: `${baseUrl.plugin}/deploy/list`,
    method: 'get',
  });
}

/**
 * 根据name获取deploy插件
 * @param name
 */
export function fetchDeployByName(name: string): Promise<INodeVo> {
  return restProxy<INodeVo>({
    url: `${baseUrl.plugin}/${name}`,
    method: 'get',
  });
}

/**
 * 获取exec插件
 */
export function fetchExecPlugins(): Promise<INodeVo[]> {
  return restProxy<INodeVo[]>({
    url: `${baseUrl.plugin}/exec/list`,
    method: 'get',
  });
}

/**
 * 根据name获取exec插件
 * @param name
 */
export function fetchExecByName(name: string): Promise<INodeVo> {
  return restProxy<INodeVo>({
    url: `${baseUrl.plugin}/${name}`,
    method: 'get',
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