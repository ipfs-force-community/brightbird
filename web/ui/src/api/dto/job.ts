import { JobEnum } from "./enumeration"

export interface IJobIdVo
  extends Readonly<{
    id: string;
  }> {}

export interface IPRMergedEventMatch
extends Readonly<{
  repo: string;
  basePattern: string;
  sourcePattern: string;
}> {}

export interface ITagCreateEventMatch
extends Readonly<{
  repo: string;
  tagPattern: string;
}> {}
  
export interface ICountJobParam
  extends Readonly<{
    id?: string;
    name?: string;
  }> { }

export interface IJobVo extends Readonly<{
    id: string;
    testFlowId: string;
    name: string;
    jobType: JobEnum;
    description: string;
    versions: Map<string, string>;

    cronExpression: string;
    prMergedEventMatchs: IPRMergedEventMatch[];
    tagCreateEventMatchs: ITagCreateEventMatch[];

    createTime:string;
    modifiedTime:string;
  }> {
  }
  
  export interface IJobDetailVo extends Readonly<{
    id: string;
    testFlowId: string;
    name: string;
    testFlowName:string;
    groupName:string;
    jobType: JobEnum;
    description: string;
    versions: Map<string, string>;

    cronExpression: string;
    prMergedEventMatchs: IPRMergedEventMatch[];
    tagCreateEventMatchs: ITagCreateEventMatch[];

    createTime:string;
    modifiedTime:string;
  }> {
  }


export interface IJobCreateVo extends Readonly<{
    testFlowId: string;
    name: string;
    jobType: JobEnum;
    description: string;
    versions: any;

    cronExpression: string;
    prMergedEventMatchs: IPRMergedEventMatch[];
    tagCreateEventMatchs: ITagCreateEventMatch[];
    }> {
}

export interface IJobUpdateVo extends Readonly<{
    name: string;
    testFlowId: string;
    description: string;
    versions:  any;
   
    cronExpression: string;
    prMergedEventMatchs: IPRMergedEventMatch[];
    tagCreateEventMatchs: ITagCreateEventMatch[];
    }> {
}