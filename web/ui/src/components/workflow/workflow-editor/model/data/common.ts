import { RuleItem } from 'async-validator';
import { NodeTypeEnum } from './enumeration';
import { ISelectableParam } from '../../../workflow-expression-editor/model/data';
import { Case, Node} from "@/api/dto/project";

type TriggerValue = 'blur' | 'change';

export interface CustomRuleItem extends RuleItem {
  trigger?: TriggerValue;
}

export type CustomRule = CustomRuleItem | CustomRuleItem[];

/**
 * 节点数据
 */
export interface IWorkflowNode {
  getName(): string;

  getType(): NodeTypeEnum;

  getIcon(): string;

  buildSelectableParam(nodeId: string): ISelectableParam | undefined;

  getFormRules(): Record<string, CustomRule>;

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
  cases?: Case[];
  nodes?: Node[];
  graph?: string;
  description?: string;
}

export type ValidateParamFn = (value: string) => void;

export type ValidateCacheFn = (name: string) => void;
