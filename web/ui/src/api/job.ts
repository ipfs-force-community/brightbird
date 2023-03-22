import { IJobCreateVo, IJobUpdateVo, IJobIdVo, IJobVo } from "./dto/job";
import { restProxy } from '@/api';
import { JobEnum } from "./dto/enumeration";


export const baseUrl = {
    job: '/job'
};

/**
 * 新建job
 * @param dto
 */
export async function createJob(dto: IJobCreateVo): Promise<IJobIdVo> {
    const res = await restProxy({ 
        url:baseUrl.job,
        method:"post",
        payload:dto
    });
  
    return {
        id: res
    };
}

/**
 * 更新job
 * @param dto
 */
export async function updateJob(id: String, dto: IJobUpdateVo): Promise<void> {
    await restProxy({ 
        url:`${baseUrl.job}/${id}`,
        method:"post",
        payload:dto
    });
}



/**
 * 获取job列表
 * @param dto
 */
export async function listJobs(): Promise<IJobVo[]> {
    const res = await restProxy({ 
        url:baseUrl.job,
        method:"get",
    });
  
    return res
}

/**
 * 获取job
 * @param dto
 */
export async function getJob(id: String): Promise<IJobVo> {
    const res = await restProxy({ 
        url:`${baseUrl.job}/${id}`,
        method:"get",
    });
  
    return res
}

/**
 * 获取job
 * @param dto
 */
export async function getJobTypes(): Promise<JobEnum[]> {
   return [JobEnum.CronJob, JobEnum.Fake]
}


/**
 * 获取job
 * @param dto
 */
export async function deleteJob(id: String): Promise<IJobVo> {
    const res = await restProxy({ 
        url:`${baseUrl.job}/${id}`,
        method:"delete",
    });
  
    return res
}