import { IJobCreateVo, IJobUpdateVo, IJobIdVo, IJobVo, IJobDetailVo } from "./dto/job";
import { restProxy } from '@/api';
import { JobEnum } from "./dto/enumeration";
import { ITaskVo, IListTaskVo} from "./dto/tasks";
import { IPageDto, IPageVo } from "./dto/common";

export const baseUrl = {
    task: '/task'
};

/**
 * 获取job
 * @param dto
 */
export async function getTask(id: String): Promise<ITaskVo> {
    const res = await restProxy({ 
        url:`${baseUrl.task}/${id}`,
        method:"get",
    });
  
    return res
}

/**
 * 获取job
 * @param dto
 */
export async function getTaskInJob(req: IListTaskVo): Promise<IPageVo<ITaskVo>> {
    const res = await restProxy({ 
        url:`${baseUrl.task}/list`,
        payload: req,
        method:"get",
    });
  
    return res
}

/**
 * 停止人物执行
 * @param dto
 */
export async function stopTask(id: String): Promise<ITaskVo[]> {
    const res = await restProxy({ 
        url:`${baseUrl.task}/stop/${id}`,
        method:"post",
    });
  
    return res
}