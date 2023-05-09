import {BaseNode} from './base-node';
import {NodeGroupEnum, NodeTypeEnum, ParamTypeEnum} from '../enumeration';
import defaultIcon from '../../../svgs/shape/async-task.svg';
import {TaskStatusEnum} from '@/api/dto/enumeration';
import {IPropertyDto} from "@/api/dto/testflow";

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
  readonly svcProperties: IPropertyDto[];
  out: IPropertyDto;
  createTime: string;
  modifiedTime: string;

  constructor(
      name: string,
      type: NodeTypeEnum,
      icon = '',
      groupType: NodeGroupEnum,
      version = '',
      category = '',
      inputs: IPropertyDto[],
      svcProperties: IPropertyDto[],
      createTime = '',
      modifiedTime = '',
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
    this.svcProperties = svcProperties;
    this.createTime = createTime;
    this.modifiedTime = modifiedTime;
    this.out = out;
  }

  static build(
      { name, type, icon, groupType, version, category, inputs, svcProperties, createTime, modifiedTime, out}: any,
  ): AsyncTask {
    return new AsyncTask(name, type, icon, groupType, version, category, inputs, svcProperties, createTime, modifiedTime, out);
  }

  // eslint-disable-next-line @typescript-eslint/ban-types
  toDsl(): object {
    const { name, version, inputs, svcProperties, out } = this;
    const param: {
      [key: string]: string | number | boolean;
    } = {};
    if (inputs && inputs.length > 0) {
      inputs.forEach(({name, type, required, value}) => {
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

        if (!required && !value && type !== ParamTypeEnum.STRING) {
          delete param[name];
        }
      });
    }

    const svc: {
      [key: string]: string | number | boolean;
    } = {};
    if (svcProperties && svcProperties.length > 0) {
      svcProperties.forEach(({name, type, required, value}) => {
        switch (type) {
          case ParamTypeEnum.NUMBER: {
            const val = parseFloat(value);
            if (!isNaN(val)) {
              svc[name] = val;
              return;
            }
            break;
          }
          case ParamTypeEnum.BOOL: {
            switch (value) {
              case 'true':
                svc[name] = true;
                return;
              case 'false':
                svc[name] = false;
                return;
            }
            break;
          }
        }

        if (!svc[name]) {
          svc[name] = value;
        }

        if (!required && !value && type !== ParamTypeEnum.STRING) {
          delete svc[name];
        }
      });
    }

    const output: {
      [key: string]: string | number | boolean;
    } = {};
    if (out) {
      switch (out.type) {
        case ParamTypeEnum.NUMBER: {
          const val = parseFloat(out.value);
          if (!isNaN(val)) {
            output[out.name] = val;
          }
          break;
        }
        default: {
          if (!output[out.name]) {
            output[out.name] = out.value;
          }

          if (!out.required && !out.value && out.type !== ParamTypeEnum.STRING) {
            delete output[out.name];
          }
        }
      }
    }


    return {
      type: `${this.name}:${version}`,
      param: inputs && inputs.length === 0 ? undefined : param,
      svc: svcProperties && svcProperties.length == 0 ? undefined: svc,
      output: !out ? undefined: output,
    };
  }

}
