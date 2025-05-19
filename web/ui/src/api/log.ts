
import { restProxy } from '@/api/index';
import { baseUrl } from '@/api/view-no-auth';
import { LogReq, LogResp, ListPodsReq } from './dto/log';


/**
 * 获取task中的Pod
 * @param taskId
 */
export function listAllPod(req: ListPodsReq): Promise<string[]> {
  return restProxy({
    url: `${baseUrl.log}/pods`,
    method: 'get',
    payload: req,
  });
}

/**
 * 获取Pod中的Log
 * @param podName
 */
export function getPodLog(req: LogReq): Promise<LogResp> {
  return restProxy({
    url: `${baseUrl.log}/logs`,
    method: 'get',
    payload: req,
  });
}