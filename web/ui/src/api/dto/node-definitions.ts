/**
 * 本地/官方-获取节点定义版本列表
 */
export interface INodeDefVersionListVo
    extends Readonly<{
        versions: string[]
    }> {
}

/**
 * 获取节点定义版本-输入/输出参数
 */
export interface INodeParameterVo
    extends Readonly<{
        name: string;
        ref: string;
        type: string;
        description: string;
        value: object;
        required: boolean;
    }> {
}

/**
 * 官方-获取节点版本定义
 */
export interface INodeDefinitionVersionExampleVo
    extends Readonly<{
        id: number;
        versionNumber: string;
        description: string;
        workflowExample: string;
        pipelineExample: string;
        inputParams: INodeParameterVo[];
        outputParams: INodeParameterVo[];
        creatorName: string;
        creatorRef: string;
        creatorPortrait: string;
        createTime: string;
    }> {
}