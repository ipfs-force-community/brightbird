import { IPageDto } from '@/api/dto/common';
import {
  DslTypeEnum,
  ProjectImporterTypeEnum,
} from '@/api/dto/enumeration';

/**
 * 保存项目dto
 */
export interface ITestFlowDetail
    extends Readonly<{
        id?: string;
        name: string;
        createTime: string;
        modifiedTime: string;
        cases: Case[];
        nodes: Node[];
        groupId: string;
        graph: string;
        description: string;
  }> {}

  export interface IGetTestFlowParam
      extends Readonly<{
          id: string;
          name: string;
    }> {}

export interface Node
    extends Readonly<{
        name: string;
        type: string;
        isAnnotateOut: boolean;
        properties: IPropertyDto[];
        svcProperties: IPropertyDto[];
        out:IPropertyDto;
    }> {}

export interface Case
    extends Readonly<{
        name: string;
        type: string;
        properties: IPropertyDto[];
        svcProperties: IPropertyDto[];
    }> {}

/**
 * 克隆Git库dto
 */
export interface IGitCloningDto
  extends Readonly<{
    uri: string;
      groupId: string;
    credential: {
      type?: ProjectImporterTypeEnum;
      namespace?: string;
      userKey?: string;
      passKey?: string;
      privateKey?: string;
    };
    branch: string;
  }> {}

/**
 * git值对象
 */
export interface IGitVo
  extends Readonly<{
    id: string;
    uri: string;
    branch: string;
  }> {}

/**
 * 导入项目dto
 */
export interface IProjectImportingDto
  extends Readonly<
    IGitCloningDto & {
      id: string;
      dslPath: string;
    }
  > {}

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
  > {}

/**
 * 项目id vo
 */
export interface ITestFlowIdVo
  extends Readonly<{
    id: string;
  }> {}

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
  }> {}

/**
 * 任务参数vo
 */
export interface ITaskParameterVo
  extends Readonly<{
    ref: string;
    expression: string;
  }> {}

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
  }> {}

/**
 * 全局参数vo
 */
export interface IGlobalParameterVo
  extends Readonly<{
    name: string;
    type: string;
    value: string | number | boolean;
  }> {}

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
  }> {}

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
  }> {}

export interface IPropertyDto extends Readonly<{
    name: string;
    type: string;
    value: any;
    required: true;
    description: string;
}> {
}

export interface INodeVo extends Readonly<{
    icon: string;
    name: string;
    createTime: string;
    modifiedTiem: string;
    version: string;
    category: string;
    description: string;
    path: string;
    isAnnotateOut: boolean;
    properties: IPropertyDto[];
    svcProperties: IPropertyDto[];
    out:IPropertyDto;
    deprecated: boolean;
}> {
}


/**
 * 测试流组添加测试流dto
 */
export interface IChangeTestflowGroupDto
  extends Readonly<{
      groupId: string;
      testflowIds: string[];
  }> {}