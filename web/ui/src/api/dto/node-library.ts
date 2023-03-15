import { NodeTypeEnum } from './enumeration';

export interface INodeCreatingDto extends Readonly<{
  name: string;
  description?: string;
  dsl: string;
}> {
}

export interface IPropertyDto extends Readonly<{
  name: string;
  type: string;
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
  properties:[]IPropertyDto
  deprecated: boolean;
}> {
}


type PluginOut struct {
  BaseTime
  PluginInfo
  Properties    []Property `json:"properties"`
  IsAnnotateOut bool       `json:"isAnnotateOut"`
  SvcProperties []Property `json:"svcProperties"`
  Out           *Property  `json:"out"`
}