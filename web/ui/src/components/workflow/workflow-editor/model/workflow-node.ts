import { IWorkflowNode } from './data/common';
import { Cron } from './data/node/cron';
import { Webhook } from './data/node/webhook';
import { Shell } from './data/node/shell';
import { AsyncTask } from './data/node/async-task';
import { INodeParameterVo } from '@/api/dto/node-definitions';
import { ParamTypeEnum } from '@/components/workflow/workflow-editor/model/data/enumeration';

interface IPageInfo {
  pageNum: number;
  totalPages: number;
  content: IWorkflowNode[];
}

/**
 * push输入/输出参数
 * @param data 原数据
 * @param inputs
 * @param outputs
 */
export const pushParams = (data: AsyncTask) => {
/*  if (data.inputProperties) {
    inputProperties.forEach(item => {
      data.inputProperties.push({
        ref: item.ref,
        name: item.name,
        type: item.type as ParamTypeEnum,
        value: (item.value || '').toString(),
        required: item.required,
        description: item.description,
      });
    });
  }*/
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
}
