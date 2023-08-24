import { RuleItem } from 'async-validator';
import { NodeTypeEnum } from './enumeration';
import { ISelectableParam } from '../../../workflow-expression-editor/model/data';
import { JSONSchema } from 'json-schema-to-typescript';
import { GlobalProperty } from '@/api/dto/testflow';

type TriggerValue = 'blur' | 'change';

export interface CustomRuleItem extends RuleItem {
  trigger?: TriggerValue;
}

export type CustomRule = CustomRuleItem | CustomRuleItem[];

/**
 * 节点数据
 */
export interface IWorkflowNode {
  getInstanceName(): string;

  getName(): string;
  
  getType(): NodeTypeEnum;

  getIcon(): string;

  getLabels(): string[];

  buildSelectableParam(nodeId: string): ISelectableParam | undefined;

  getFormRules(): Record<string, CustomRule>;

  setInstanceName(name:string): void;
  /**
   * 校验
   * @throws Error
   */
  validate(): Promise<void>;

  toDsl(): object;
}

export interface IGlobal {
  concurrent: number | boolean;
}

/**
 * 工作流数据
 */
export interface IWorkflow {
  name: string;
  groupId: string;
  createTime: string;
  modifiedTime: string;
  graph?: string;
  description?: string;
  data: string;
  globalProperties?: GlobalProperty[];

}


export interface TreeProp {
  name: string;
  isLeaf: boolean;
  type: string;
  index: number|undefined;
  defs:Map<string, JSONSchema>;
  schema: JSONSchema|null;
  children: TreeProp[]
}
