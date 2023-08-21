import { AddLabelReq, DeleteLabelReq, PluginDef, PluginDetail } from '@/api/dto/node-definitions';
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
 * 根据name和version获取插件定义
 * @param name
 */
export function getPluginDef(name: string, version: string): Promise<PluginDef> {
  return restProxy<PluginDef>({
    url: `${baseUrl.plugin}/def`,
    method: 'get',
    payload: {
      'name': name,
      'version':version,
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
 * 删除插件的全部版本
 */
export function deletePluginAllVersion(id: string): Promise<void> {
  return restProxy<void>({
    url: `${baseUrl.plugin}/all?id=${id}`,
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
export function uploadPlugin(formData: object): Promise<void> {
  return restProxy({
    url: `${baseUrl.plugin}/upload`,
    method: 'post',
    payload: formData,
    payloadType:"file",
    timeout: 200 * 1000,
  });
}

/**
 * Download the compressed 'public' folder
 */
export function downloadPublicZip(): Promise<void> {
  return restProxy({
    url: `${baseUrl.plugin}/download`,
    method: 'get',
  });
}