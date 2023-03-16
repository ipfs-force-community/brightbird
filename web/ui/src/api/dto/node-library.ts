export interface IPropertyDto extends Readonly<{
  name: string;
  type: string;
  value: any;
  require:true;
  description: string;
}> {
}

export interface INodeVo extends Readonly<{
  icon: string;
  name: string;
  version: string
  category: string
  description: string
  path: string
  isAnnotateOut: boolean
  properties: IPropertyDto[]
  SvcProperties: IPropertyDto[]
  Out?:IPropertyDto
  deprecated: boolean;
}> {
}