import { BaseVo } from '@/api/dto/common';
/**
 * 项目组vo
 */
export interface ITestflowGroupVo
  extends Readonly<
    BaseVo & {
      id: string;
      name: string;
      isShow: boolean;
      testFlowCount: number;
      description?: string;
      modifiedTime?:string;
      createTime?:string
    }
  > { }
/**
 * 创建项目组dto
 */

export interface IProjectGroupCreatingDto
  extends Readonly<{
    name: string;
    isShow: boolean;
    description?: string;
  }> { }
/**
 * 编辑项目组dto
 */

export interface IProjectGroupEditingDto
  extends Readonly<{
    name: string;
    isShow: boolean;
    description?: string;
  }> { }
/**
 * 创建项目组dto
 */

export interface IProjectGroupDto
  extends Readonly<{
    name: string;
    description?: string;
  }> { }

/**
 *修改项目组排序dto
 */

export interface IProjectGroupSortUpdatingDto
  extends Readonly<{
    originGroupId: string;
    targetGroupId: string;
  }> { }


/**
 * 修改项目组中的项目排序dto
 */
export interface IProjectSortUpdatingDto
  extends Readonly<{
    originProjectId: string;
    targetProjectId: string;
  }> { }

export interface ICountGroupParam
  extends Readonly<{
    id?: string;
    name?: string;
  }> { }