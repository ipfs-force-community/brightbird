import {PluginTypeEnum} from "@/api/dto/enumeration";

/**
 * 本地/官方-获取节点定义版本列表
 */
export interface INodeDefVersionListVo
    extends Readonly<{
        versions: string[]
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
    description: string;
    buildScript: string;
    repo: string;
    imageTarget: string;
    path: string;
    inputSchema : any;
    outputSchema: any;
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
    isDeleting: boolean;
}