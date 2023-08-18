import { IPageDto } from '@/api/dto/common';
import {
  DslTypeEnum,
  PluginTypeEnum,
} from '@/api/dto/enumeration';

export interface Node
  extends Readonly<{
    name: string;
    instanceName: string;
    version: string;
    input: string;
    output: string;
  }> { }

export interface GlobalProperty
    extends Readonly<{
        name: string;
        type: string;
        value: string;
  }> { }

/**
 * 保存项目dto
 */
export interface ITestFlowDetail
  extends Readonly<{
    id?: string;
    name: string;
    createTime: string;
    modifiedTime: string;
    groupId: string;
    graph: string;
    globalProperties: GlobalProperty;
    description: string;
  }> { }

export interface IGetTestFlowParam
  extends Readonly<{
    id?: string;
    name?: string;
  }> { }

export interface ICountTestFlowParam
  extends Readonly<{
    groupId?: string;
    name?: string;
  }> { }



/**
 * 查询项目dto
 */
export interface IProjectQueryingDto
  extends Readonly<
    IPageDto & {
      name?: string;
      groupId: string;
      pageNum?: number;
      pageSize?: number;
    }
  > { }

/**
 * 项目id vo
 */
export interface ITestFlowIdVo
  extends Readonly<{
    id: string;
  }> { }

/**
 * 节点定义vo
 */
export interface INodeDefVo
  extends Readonly<{
    name: string;
    description?: string;
    icon?: string;
    ownerName: string;
    ownerType: string;
    ownerRef: string;
    creatorName: string;
    creatorRef: string;
    sourceLink?: string;
    documentLink?: string;
    type: string;
  }> { }


/**
 * 测试流组添加dto
 */
export interface IChangeTestflowGroupDto
  extends Readonly<{
    groupId: string;
    testflowIds: string[];
  }> { }