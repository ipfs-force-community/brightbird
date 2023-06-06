import {IWorkflowNode} from './data/common';
import {AsyncTask} from './data/node/async-task';
import {NodeTypeEnum, ParamTypeEnum} from '@/components/workflow/workflow-editor/model/data/enumeration';
import {fetchDeployPlugins, fetchExecPlugins, fetchPluginsMainfest} from "@/api/view-no-auth";
import { PluginTypeEnum } from '@/api/dto/enumeration';

interface IPageInfo {
  content: IWorkflowNode[];
}

export class WorkflowNode {

  constructor() {
  }

  async loadDeployPlugins(keyword?: string): Promise<IPageInfo> {
    const nodes = await fetchPluginsMainfest({pluginType:PluginTypeEnum.Deploy});
    const arr: IWorkflowNode[] = nodes.map(item => new AsyncTask(item.name, NodeTypeEnum.ASYNC_TASK, item.icon,
      item.pluginType, "", item.pluginType, [], [], 0,
      0, {
        name:"instance",
        value:"",
        type: item.pluginType,
        sockPath: "",
        require:true,
        description:"节点实例名称",
      }));

    return {
      content: keyword ? arr.filter(item => item.getDisplayName().includes(keyword)) : arr,
    };
  }

  async loadExecPlugins(keyword?: string): Promise<IPageInfo> {
    const nodes = await fetchPluginsMainfest({pluginType:PluginTypeEnum.Exec});
    const arr: IWorkflowNode[] = nodes.map(item => new AsyncTask(item.name, NodeTypeEnum.ASYNC_TASK, item.icon,
      item.pluginType, "", item.pluginType, [], [], 0,
      0, {
        name:"instance",
        value:"",
        type: item.pluginType,
        sockPath: "",
        require:true,
        description:"节点实例名称",
      }));
    return {
      content: keyword ? arr.filter(item => item.getDisplayName().includes(keyword)) : arr,
    };
  }
}
