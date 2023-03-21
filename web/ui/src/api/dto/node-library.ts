export interface IPropertyDto extends Readonly<{
  name: string;
  type: string;
  value: any;
  required: true;
  description: string;
}> {
}

export interface INodeVo extends Readonly<{
  icon: string;
  name: string;
  createTime: string;
  modifyTime: string;
  version: string;
  category: string;
  description: string;
  path: string;
  isAnnotateOut: boolean;
  properties: IPropertyDto[];
  svcProperties: IPropertyDto[];
  out?:IPropertyDto;
  deprecated: boolean;
}> {
}