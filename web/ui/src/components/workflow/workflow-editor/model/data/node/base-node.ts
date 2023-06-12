import Schema, { Value } from 'async-validator';
import { CustomRule, IWorkflowNode } from '../common';
import { NodeTypeEnum } from '../enumeration';
import { ISelectableParam } from '../../../../workflow-expression-editor/model/data';

export abstract class BaseNode implements IWorkflowNode {
  name: string;
  labels: string[];
  private readonly type: NodeTypeEnum;
  private readonly icon: string;

  protected constructor(name: string,
    type: NodeTypeEnum, icon: string, labels: string[]) {
    this.name = name;
    this.type = type;
    this.icon = icon;
    this.labels = labels;
  }

  getDisplayName(): string {
    return this.name;
  }

  getType(): NodeTypeEnum {
    return this.type;
  }

  getIcon(): string {
    return this.icon;
  }

  getLabels(): string[] {
    return this.labels;
  }

  buildSelectableParam(nodeId: string): ISelectableParam | undefined {
    return undefined;
  }

  getFormRules(): Record<string, CustomRule> {
    return {
      name: [
        {
          required: true,
          message: '节点名称不能为空',
          trigger: 'blur',
        },
      ],
    };
  }

  async validate(): Promise<void> {
    const validator = new Schema(this.getFormRules());

    const source: Record<string, Value> = {};
    Object.keys(this).forEach(key => (source[key] = (this as any)[key]));

    await validator.validate(source, {
      first: true,
    });
  }

  toDsl(): object {
    return {};
  }
  setInstanceName(name: string) {}
}