import { BaseNode } from './base-node';
import { NodeTypeEnum, ParamTypeEnum } from '../enumeration';
import defaultIcon from '../../../svgs/shape/async-task.svg';
import { PluginTypeEnum, TaskStatusEnum } from '@/api/dto/enumeration';
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
  groupType: PluginTypeEnum;
  version: string;
  pluginType: string;
  input: any;
  output: any;
  createTime: number;
  modifiedTime: number;

  constructor(
    name: string,
    instanceName: string,
    type: NodeTypeEnum,
    icon = '',
    labels: string[] = [],
    groupType: PluginTypeEnum,
    version = '',
    pluginType = '',
    input = {},
    createTime = 0,
    modifiedTime = 0,
  ) {
    super(
      name,
      instanceName,
      NodeTypeEnum.ASYNC_TASK,
      checkDefaultIcon(icon) ? defaultIcon : icon, labels
    );
    this.groupType = groupType;
    this.version = version;
    this.pluginType = pluginType;
    this.input = input;
    this.createTime = createTime;
    this.modifiedTime = modifiedTime;
  }


  static build(
    { name, instanceName, type, icon, labels, groupType, version, pluginType, input, createTime, modifiedTime }: any,
  ): AsyncTask {
    return new AsyncTask(name, instanceName, type, icon, labels, groupType, version, pluginType, input, createTime, modifiedTime);
  }

  // eslint-disable-next-line @typescript-eslint/ban-types
  toDsl(): object {
    //todo rewrite all dsl data
    const { name, instanceName, version } = this;

    return {
      name: name,
      instanceName: instanceName,
      version: version,
      pluginType: this.pluginType,
      input: this.input,
    };
  }

  setInstanceName(name: string) {
    this.instanceName = name;
  }

  getInstanceName(): string {
    return this.instanceName;
  }
}
