import { IJobCreateVo, IJobUpdateVo, IJobIdVo, IJobVo, IJobDetailVo } from './dto/job';
import { restProxy } from '@/api';
import { JobEnum } from './dto/enumeration';
import { ITaskVo, IListTaskVo, IGetTaskReq } from './dto/tasks';
import { IPageDto, IPageVo } from './dto/common';

export const baseUrl = {
  task: '/task',
};

/**
 * 获取job
 * @param dto
 */
export async function getTask(req: IGetTaskReq): Promise<ITaskVo> {
  const res = await restProxy({ 
    url:`${baseUrl.task}`,
    method:'get',
    payload:req 
  });
  
  return res;
}

/**
 * 获取job
 * @param dto
 */
export async function getTaskInJob(req: IListTaskVo): Promise<IPageVo<ITaskVo>> {
  const res = await restProxy({ 
    url:`${baseUrl.task}/list`,
    payload: req,
    method:'get',
  });
  
  return res;
}

/**
 * 停止任务执行
 * @param dto
 */
export async function stopTask(id: string): Promise<ITaskVo[]> {
  const res = await restProxy({ 
    url:`${baseUrl.task}/stop/${id}`,
    method:'post',
  });
  
  return res;
}

/**
 * 重试任务
 * @param dto
 */
export async function retryTask(id: string): Promise<ITaskVo[]> {
  const res = await restProxy({ 
    url:`${baseUrl.task}/retry`,
    method:'post',
    payload:{ID:id}
  });
  
  return res;
}
