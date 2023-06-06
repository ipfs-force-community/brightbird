import {BaseNode} from './base-node';
import {NodeTypeEnum, ParamTypeEnum} from '../enumeration';
import defaultIcon from '../../../svgs/shape/async-task.svg';
import {PluginTypeEnum, TaskStatusEnum} from '@/api/dto/enumeration';
import { DependencyProperty, Property } from "@/api/dto/testflow";
import { CustomRule } from '../common';
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
  inputs: Property[];
  dependencies: DependencyProperty[];
  instance: DependencyProperty;
  createTime: number;
  modifiedTime: number;

  constructor(
      name: string,
      type: NodeTypeEnum,
      icon = '',
      groupType: PluginTypeEnum,
      version = '',
      pluginType = '',
      inputs: Property[],
      dependencies: DependencyProperty[],
      createTime = 0,
      modifiedTime = 0,
      instance: DependencyProperty,
  ) {
    super(
        name,
        NodeTypeEnum.ASYNC_TASK,
        checkDefaultIcon(icon) ? defaultIcon : icon,
    );
    this.groupType = groupType;
    this.version = version;
    this.pluginType = pluginType;
    this.inputs = inputs;
    this.dependencies = dependencies;
    this.createTime = createTime;
    this.modifiedTime = modifiedTime;
    this.instance = instance;
  }



  static build(
      { name, type, icon, groupType, version, pluginType, inputs, dependencies, createTime, modifiedTime, instance}: any,
  ): AsyncTask {
    return new AsyncTask(name, type, icon, groupType, version, pluginType, inputs, dependencies, createTime, modifiedTime, instance);
  }

  // eslint-disable-next-line @typescript-eslint/ban-types
  toDsl(): object {
    const { name, version, inputs, dependencies, instance } = this;
    const param: {
      [key: string]: string | number | boolean;
    } = {};
    if (inputs && inputs.length > 0) {
      inputs.forEach(({name, type, require, value}) => {
        switch (type) {
          case ParamTypeEnum.NUMBER: {
            const val = parseFloat(value);
            if (!isNaN(val)) {
              param[name] = val;
              return;
            }
            break;
          }
          case ParamTypeEnum.BOOL: {
            switch (value) {
              case 'true':
                param[name] = true;
                return;
              case 'false':
                param[name] = false;
                return;
            }
            break;
          }
        }

        if (!param[name]) {
          param[name] = value;
        }

        if (!require && !value && type !== ParamTypeEnum.STRING) {
          delete param[name];
        }
      });
    }

    const svc: {
      [key: string]: string | number | boolean;
    } = {};
    if (dependencies && dependencies.length > 0) {
      dependencies.forEach(({name, type, require, value}) => {
        if (!svc[name]) {
          svc[name] = value;
        }

        if (!require && !value) {
          delete svc[name];
        }
      });
    }

    const output: {
      [key: string]: string | number | boolean;
    } = {};
    if (instance) {
      if (!output[instance.name]) {
        output[instance.name] = instance.value;
      }

      if (!instance.require && !instance.value) {
        delete output[instance.name];
      }
    }


    return {
      type: `${this.name}:${version}`,
      param: inputs && inputs.length === 0 ? undefined : param,
      dependencies: dependencies && dependencies.length == 0 ? undefined: svc,
      instance: !instance ? undefined: output,
    };
  }

  setInstanceName(name: string) {
    this.instance.value = name;
  }

  getDisplayName(): string {
    if (this.instance.value) {
      return this.instance.value
    } else {
      return this.name
    }
  }

}
