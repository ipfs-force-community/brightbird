import { restProxy } from '@/api/index';
import {
  IAsyncTaskInstanceVo,
  ITaskExecutionRecordVo,
  ITaskParamVo,
  IWorkflowExecutionRecordVo,
} from '@/api/dto/workflow-execution-record';
import { IProcessTemplateVo, IProjectDetailVo, IProjectQueryingDto, IProjectVo, IWorkflowVo } from '@/api/dto/project';
import { IPageDto, IPageVo } from '@/api/dto/common';
import { INodeVo } from '@/api/dto/node-library';
import { ITriggerEventVo, ITriggerWebhookVo } from '@/api/dto/trigger';
import { IProjectGroupVo } from '@/api/dto/project-group';
import { IProjectCacheVo, INodeCacheVo } from '@/api/dto/cache';

export const baseUrl = {
  projectGroup: '/group',
  workflow: '/view/workflow_instances',
  asyncTasks: '/view/async_task_instances',
  task: '/view/task_instance',
  log: '/view/logs',
  dsl: '/view/workflow',
  processLog: '/view/logs/workflow',
  deployPlugin: '/deploy',
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
    url: `${baseUrl.projectGroup}`,
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
export function queryProject(dto: IProjectQueryingDto): Promise<IPageVo<IProjectVo>> {
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
export function fetchProjectDetail(projectId: string): Promise<IProjectDetailVo> {
  return restProxy({
    url: `${baseUrl.testflow}/id/${projectId}`,
    method: 'get',
  });
}

/**
 * 获取流程执行记录列表
 * @param workflowRef
 */
export function listWorkflowExecutionRecord(workflowRef: string): Promise<IWorkflowExecutionRecordVo[]> {
  return restProxy<IWorkflowExecutionRecordVo[]>({
    url: `${baseUrl.workflow}/${workflowRef}`,
    method: 'get',
  });
}

/**
 * 获取异步任务实例列表
 * @param triggerId
 */
export function listAsyncTaskInstance(triggerId: string): Promise<IAsyncTaskInstanceVo[]> {
  return restProxy<IAsyncTaskInstanceVo[]>({
    url: `${baseUrl.asyncTasks}/${triggerId}`,
    method: 'get',
  });
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
 * 获取dsl
 * @param workflowRef
 * @param workflowVersion
 */
export function fetchWorkflow(workflowRef: string, workflowVersion: string): Promise<IWorkflowVo> {
  return restProxy<IWorkflowVo>({
    url: `${baseUrl.dsl}/${workflowRef}/${workflowVersion}`,
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

export function fetch_deploy_plugins(): Promise<INodeVo[]> {
  return restProxy<INodeVo[]>({
    url: `${baseUrl.deployPlugin}/plugins`,
    method: 'get',
  });
}

export function fetch_exec_plugins(): Promise<INodeVo[]> {
  return restProxy<INodeVo[]>({
    url: `${baseUrl.testflow}/plugins`,
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
