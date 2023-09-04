import { IWorkflowNode } from './data/common';
import { AsyncTask } from './data/node/async-task';
import { NodeTypeEnum, ParamTypeEnum } from '@/components/workflow/workflow-editor/model/data/enumeration';
import { fetchDeployPlugins, fetchExecPlugins } from '@/api/plugin';
import { PluginTypeEnum } from '@/api/dto/enumeration';

interface IPageInfo {
  content: IWorkflowNode[];
}

export class WorkflowNode {

  constructor() {
  }

  async loadDeployPlugins(keyword?: string): Promise<IPageInfo> {
    const nodes = await fetchDeployPlugins();
    const arr: IWorkflowNode[] = nodes.map(item => new AsyncTask(item.name, item.name, NodeTypeEnum.ASYNC_TASK, item.icon, item.labels,
      item.pluginType, '', item.pluginType, {}, 0, 0));

    return {
      content: keyword ? arr.filter(item => item.getInstanceName().includes(keyword) || item.getLabels().filter(a => a.includes(keyword)).length > 0) : arr,
    };
  }

  async loadExecPlugins(keyword?: string): Promise<IPageInfo> {
    const nodes = await fetchExecPlugins();
    const arr: IWorkflowNode[] = nodes.map(item => new AsyncTask(item.name, item.name, NodeTypeEnum.ASYNC_TASK, item.icon, item.labels,
      item.pluginType, '', item.pluginType, {}, 0,
      0));
    return {
      content: keyword ? arr.filter(item => item.getInstanceName().includes(keyword) || item.getLabels().filter(a => a.includes(keyword)).length > 0) : arr,
    };
  }
}
