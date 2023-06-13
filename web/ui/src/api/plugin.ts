import { AddLabelReq, DeleteLabelReq, PluginDetail } from '@/api/dto/testflow';
import { restProxy } from '@/api/index';
import { PluginTypeEnum } from '@/api/dto/enumeration';
import { baseUrl } from '@/api/view-no-auth';


/**
 * 获取exec插件
 */
export function fetchExecPlugins(): Promise<PluginDetail[]> {
  return restProxy<PluginDetail[]>({
    url: `${baseUrl.plugin}/list`,
    method: 'get',
    payload: {
      'pluginType': PluginTypeEnum.Exec,
    },
  });
}

/**
 * 获取deploy插件
 */
export function fetchDeployPlugins(): Promise<PluginDetail[]> {
  return restProxy<PluginDetail[]>({
    url: `${baseUrl.plugin}/list`,
    method: 'get',
    payload: {
      'pluginType': PluginTypeEnum.Deploy,
    },
  });
}

/**
 * 根据name获取插件
 * @param name
 */
export function getPluginByName(name: string): Promise<PluginDetail> {
  return restProxy<PluginDetail>({
    url: `${baseUrl.plugin}`,
    method: 'get',
    payload: {
      'name': name,
    },
  });
}


/**
 * 根据name获取插件列表
 * @param name
 */
export function listPluginByName(name: string): Promise<PluginDetail[]> {
  return restProxy<PluginDetail[]>({
    url: `${baseUrl.plugin}/list`,
    method: 'get',
    payload: {
      'name': name,
    },
  });
}

/**
 * 删除插件
 */
export function deletePlugin(id: string, version: string): Promise<void> {
  return restProxy<void>({
    url: `${baseUrl.plugin}?id=${id}&version=${version}`,
    method: 'delete',
  });
}

/**
 * add labels
 */
export function addPluginLabel(req: AddLabelReq): Promise<void> {
  return restProxy<void>({
    url: `${baseUrl.plugin}/label`,
    method: 'post',
    payload: req,
  });
}

/**
 * delete labels
 */
export function delPluginLabel(req: DeleteLabelReq): Promise<void> {
  return restProxy<void>({
    url: `${baseUrl.plugin}/label`,
    method: 'delete',
    payload: req,
  });
}

/**
 * 上传插件
 */
export function uploadPlugin(files: File[]): Promise<void> {
  const formData = new FormData();
  for (let i = 0; i < files.length; i++) {
    formData.append('plugins', files[i]);
  }
  
  return restProxy({
    url: `${baseUrl.testflow}/plugin/upload`,
    method: 'post',
    payload: formData,
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
}
