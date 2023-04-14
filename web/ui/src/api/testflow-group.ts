import { restProxy } from '@/api';
import {
  IProjectGroupCreatingDto,
  IProjectGroupEditingDto,
  IProjectGroupSortUpdatingDto,
  IChangeTestflowGroupDto,
} from '@/api/dto/testflow-group';
const baseUrl = '/group';
/**
 * 创建项目组
 */
export function createProjectGroup(
  dto: IProjectGroupCreatingDto,
): Promise<void> {
  return restProxy({
    url: `${baseUrl}`,
    method: 'post',
    payload: dto,
    auth: true,
  });
}

/**
 * 编辑项目组
 */
export function editProjectGroup(
    groupId: string,
  dto: IProjectGroupEditingDto,
): Promise<void> {
  return restProxy({
    url: `${baseUrl}/${groupId}`,
    method: 'post',
    payload: dto,
    auth: true,
  });
}

/**
 * 删除项目组
 * @param groupId 项目组id
 */
export function deleteProjectGroup(groupId: string): Promise<void> {
  return restProxy({
    url: `${baseUrl}/${groupId}`,
    method: 'delete',
    auth: true,
  });
}

/**
 * 修改项目组排序
 * @param dto
 */
export function updateProjectGroupSort(
  dto: IProjectGroupSortUpdatingDto,
): Promise<void> {
  return restProxy({
    url: `${baseUrl}/sort`,
    method: 'patch',
    payload: dto,
    auth: true,
  });
}

/**
 * 修改项目组排序
 * @param dto
 */
export function projectGroupAddProject(
  dto: IChangeTestflowGroupDto,
): Promise<void> {
  return restProxy({
    url: `${baseUrl}/projects`,
    method: 'post',
    payload: dto,
    auth: true,
  });
}

/**
 * 项目组删除项目
 */
export function deleteProjectGroupProject(
  projectLinkGroupId: string,
): Promise<void> {
  return restProxy({
    url: `${baseUrl}/projects/${projectLinkGroupId}`,
    method: 'delete',
    auth: true,
  });
}

/**
 * 修改项目组是否展示
 */
export function updateProjectGroupShow(groupId: string): Promise<void> {
  return restProxy({
    url: `${baseUrl}/${groupId}/is_show`,
    method: 'put',
    auth: true,
  });
}