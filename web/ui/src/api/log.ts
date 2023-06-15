
import { restProxy } from '@/api/index';
import { baseUrl } from '@/api/view-no-auth';
import { LogResp } from './dto/log';


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
  export function getPodLog(podName: string): Promise<LogResp> {
    return restProxy({
      url: `${baseUrl.log}/${podName}`,
      method: 'get',
    });
  }