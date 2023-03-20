import { restProxy } from '@/api';
import { INodeDefinitionVersionExampleVo, INodeDefVersionListVo} from '@/api/dto/node-definitions';

export const baseUrl = '/library';
const localVersionUrl = '/view/nodes';

const hubUrl = import.meta.env.VITE_JIANMU_HUB_API_BASE_URL;
const officialVersionUrl = 'hub/view/node_definitions';

/**
 * 删除节点
 * @param ownerRef
 * @param ref
 */
export function deleteNodeLibrary(ownerRef: string, ref: string): Promise<void> {
  return restProxy<void>({
    url: `${baseUrl}/${ownerRef}/${ref}`,
    method: 'delete',
    auth: true,
  });
}

/**
 * 同步节点
 * @param ownerRef
 * @param ref
 */
export function syncNodeLibrary(ownerRef: string, ref: string): Promise<void> {
  return restProxy<void>({
    url: `${baseUrl}/${ownerRef}/${ref}`,
    method: 'put',
    timeout: 60 * 1000,
    auth: true,
  });
}

/**
 * 通过ownerRef/ref获取版本列表
 * @param ref
 * @param ownerRef
 */
export function getOfficialVersionList(ref: string, ownerRef: string): Promise<INodeDefVersionListVo> {
  return restProxy<INodeDefVersionListVo>({
    url: `${hubUrl}/${officialVersionUrl}/${ownerRef}/${ref}/versions`,
    method: 'get',
  });
}

/**
 * 获取版本示例
 * @param ref
 * @param ownerRef
 * @param version
 */
export function getOfficialNodeParams(ref: string, ownerRef: string, version: string): Promise<INodeDefinitionVersionExampleVo> {
  return restProxy<INodeDefinitionVersionExampleVo>({
    url: `${hubUrl}/${officialVersionUrl}/${ownerRef}/${ref}/versions/${version}`,
    method: 'get',
  });
}