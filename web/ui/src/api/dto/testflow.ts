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
 * 流程模板vo
 */
export interface IProcessTemplateVo
  extends Readonly<{
    id: number;
    name: string;
    type: string;
    dsl: string;
    nodeDefs: [
      {
        name: string;
        description: string;
        type: string;
        icon: string;
        ownerRef: string;
        sourceLink: string;
        documentLink: string;
        workType: string;
      },
    ];
  }> { }

/**
 * 任务参数vo
 */
export interface ITaskParameterVo
  extends Readonly<{
    ref: string;
    expression: string;
  }> { }

/**
 * 流程节点vo
 */
export interface IWorkflowNodeVo
  extends Readonly<{
    /**
     * 节点定义名称
     */
    name: string;
    /**
     * 节点定义描述
     */
    description?: string;
    /**
     * 节点定义
     */
    metadata?: string;
    ref: string;
    type: string;
    taskParameters: ITaskParameterVo[];
    sources: string[];
    targets: string[];
  }> { }

/**
 * 全局参数vo
 */
export interface IGlobalParameterVo
  extends Readonly<{
    name: string;
    type: string;
    value: string | number | boolean;
  }> { }

/**
 * 流程vo
 */
export interface IWorkflowVo
  extends Readonly<{
    name: string;
    ref: string;
    type: DslTypeEnum;
    description?: string;
    version: string;
    nodes: IWorkflowNodeVo[];
    globalParameters: IGlobalParameterVo[];
    dslText: string;
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

export interface Property  {
    name: string;
    type: string;
    description: string;
    require: true;
    children: Property[]
}

export interface GetPluginReq extends Readonly<{
  name?: string;
  pluginType?: PluginTypeEnum;
  version?: string;
}> {
}

export interface GetPlugibMainfestReq extends Readonly<{
  name?: string;
  pluginType?: PluginTypeEnum;
}> {
}

export interface AddLabelReq extends Readonly<{
  name: string;
  label: string;
}> {
}

export interface DeleteLabelReq extends Readonly<{
  name: string;
  label: string;
}> {
}

export interface PluginDef extends Readonly<{
    name: string;
    instanceName: string;
    version: string;
    pluginType: PluginTypeEnum;
    description: string,
    repo: string,
    imageTarget: string,
    path: string;
    inputProperties: Property[];
    outputProperties: Property[];
}> {
}

export interface PluginDetail {
  id: string;
  name: string;
  pluginType: PluginTypeEnum;
  description: string,
  labels:string[];
  pluginDefs: PluginDef[]|undefined;
  createTime: number;
  modifiedTime: number;
  icon: string;
}

/**
 * 测试流组添加dto
 */
export interface IChangeTestflowGroupDto
  extends Readonly<{
    groupId: string;
    testflowIds: string[];
  }> { }