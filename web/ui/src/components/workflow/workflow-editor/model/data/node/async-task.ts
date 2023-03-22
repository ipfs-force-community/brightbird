import {BaseNode} from './base-node';
import {NodeGroupEnum, NodeTypeEnum, ParamTypeEnum} from '../enumeration';
import defaultIcon from '../../../svgs/shape/async-task.svg';
import {TaskStatusEnum} from '@/api/dto/enumeration';
import {IPropertyDto} from "@/api/dto/node-library";

export interface IAsyncTaskParam {
  readonly ref: string;
  readonly name: string;
  readonly type: ParamTypeEnum;
  readonly required: boolean;
  value: string;
  readonly description?: string;
}

/**
 * 检查是否为默认图标
 * @param icon
 */
export function checkDefaultIcon(icon: string) {
  if (!icon) {
    return true;
  }

  const tags = Object.values(TaskStatusEnum).map(status => `/${status}.`);
  tags.push('/async-task.');
  for (const tag of tags) {
    if (icon.includes(tag)) {
      return true;
    }
  }

  return false;
}

export class AsyncTask extends BaseNode {
  groupType: NodeGroupEnum;
  version: string;
  category: string;
  readonly inputs: IPropertyDto[];
  readonly outputs: IPropertyDto[];
  out: IPropertyDto;
  createTime: string;
  modifiedTime: string;
  isAnnotateOut: boolean;

  constructor(
      name: string,
      type: NodeTypeEnum,
      icon = '',
      groupType: NodeGroupEnum,
      version = '',
      category = '',
      inputs: IPropertyDto[],
      outputs: IPropertyDto[],
      createTime = '',
      modifiedTime = '',
      isAnnotateOut = false,
      out: IPropertyDto,
  ) {
    super(
        name,
        NodeTypeEnum.ASYNC_TASK,
        checkDefaultIcon(icon) ? defaultIcon : icon,
    );
    this.groupType = groupType;
    this.version = version;
    this.category = category;
    this.inputs = inputs;
    this.outputs = outputs;
    this.createTime = createTime;
    this.modifiedTime = modifiedTime;
    this.isAnnotateOut = isAnnotateOut;
    this.out = out;
  }

  static build(
      { name, type, icon, groupType, version, category, inputs, outputs, createTime, modifiedTime, isAnnotateOut, out}: any,
  ): AsyncTask {
    return new AsyncTask(name, type, icon, groupType, version, category, inputs, outputs, createTime, modifiedTime, isAnnotateOut, out);
  }

}
