import { IWorkflowNode } from './data/common';
import { Cron } from './data/node/cron';
import { Webhook } from './data/node/webhook';
import { Shell } from './data/node/shell';
import { AsyncTask } from './data/node/async-task';
import { INodeParameterVo } from '@/api/dto/node-definitions';
import { ParamTypeEnum } from '@/components/workflow/workflow-editor/model/data/enumeration';
import {fetch_deploy_plugins, fetch_exec_plugins} from "@/api/view-no-auth";

interface IPageInfo {
  content: IWorkflowNode[];
}

/**
 * push输入/输出参数
 * @param data 原数据
 * @param inputs
 * @param outputs
 */
export const pushParams = (data: AsyncTask, inputs: INodeParameterVo[], outputs: INodeParameterVo[], versionDescription: string) => {
  data.versionDescription = versionDescription;
  if (inputs) {
    inputs.forEach(item => {
      data.inputs.push({
        ref: item.ref,
        name: item.name,
        type: item.type as ParamTypeEnum,
        value: (item.value || '').toString(),
        required: item.required,
        description: item.description,
      });
    });
  }
  if (outputs) {
    outputs.forEach(item => {
      data.outputs.push({
        ref: item.ref,
        name: item.name,
        type: item.type as ParamTypeEnum,
        value: (item.value || '').toString(),
        required: item.required,
        description: item.description,
      });
    });
  }
};

export class WorkflowNode {

  constructor() {
  }

  loadInnerTriggers(keyword?: string): IWorkflowNode[] {
    const arr: IWorkflowNode[] = [new Cron(), new Webhook()];

    return keyword ? arr.filter(item => item.getName().includes(keyword)) : arr;
  }

  loadInnerNodes(keyword?: string): IWorkflowNode[] {
    const arr: IWorkflowNode[] = [new Shell()];

    return keyword ? arr.filter(item => item.getName().includes(keyword)) : arr;

  }

  async loadDeployPlugins(keyword?: string): Promise<IPageInfo> {
    const nodes = await fetch_deploy_plugins();
    const arr: IWorkflowNode[] = nodes.map(item => new AsyncTask(item.name, item.icon, item.version, item.category, "", "", ""));

    return {
      content: keyword ? arr.filter(item => item.getName().includes(keyword)) : arr,
    };
  }

  async loadExecPlugins(keyword?: string): Promise<IPageInfo> {
    const nodes = await fetch_exec_plugins();
    const arr: IWorkflowNode[] = nodes.map(item => new AsyncTask(item.name, item.icon, item.version, item.category, "", "", ""));
    return {
      content: keyword ? arr.filter(item => item.getName().includes(keyword)) : arr,
    };
  }
}
