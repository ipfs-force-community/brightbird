import {IWorkflowNode} from './data/common';
import {AsyncTask} from './data/node/async-task';
import {NodeGroupEnum, NodeTypeEnum, ParamTypeEnum} from '@/components/workflow/workflow-editor/model/data/enumeration';
import {IPropertyDto} from "@/api/dto/node-library";
import {fetchDeployPlugins, fetchExecPlugins} from "@/api/view-no-auth";

interface IPageInfo {
  content: IWorkflowNode[];
}

export class WorkflowNode {

  constructor() {
  }

  async loadDeployPlugins(keyword?: string): Promise<IPageInfo> {
    const nodes = await fetchDeployPlugins();
    const arr: IWorkflowNode[] = nodes.map(item => new AsyncTask(item.name, NodeTypeEnum.ASYNC_TASK, item.icon,
        NodeGroupEnum.DEPLOY, item.version, item.category, item.properties, item.svcProperties, item.createTime,
        item.modifiedTiem, item.isAnnotateOut, item.out));

    return {
      content: keyword ? arr.filter(item => item.getName().includes(keyword)) : arr,
    };
  }

  async loadExecPlugins(keyword?: string): Promise<IPageInfo> {
    const nodes = await fetchExecPlugins();
    const arr: IWorkflowNode[] = nodes.map(item => new AsyncTask(item.name, NodeTypeEnum.ASYNC_TASK, item.icon,
        NodeGroupEnum.DEPLOY, item.version, item.category, item.properties, item.svcProperties, item.createTime,
        item.modifiedTiem, item.isAnnotateOut, item.out));
    return {
      content: keyword ? arr.filter(item => item.getName().includes(keyword)) : arr,
    };
  }
}
