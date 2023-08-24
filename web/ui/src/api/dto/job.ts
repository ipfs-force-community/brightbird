import { JobEnum } from './enumeration';
import { GlobalProperty } from './testflow';

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


export interface IJobNextParam
extends Readonly<{
  id: string;
  n: number;
}> { }

export interface IJobVo extends Readonly<{
    id: string;
    testFlowId: string;
    name: string;
    jobType: JobEnum;
    description: string;
    versions: Map<string, string>;
    cronExpression: string;
    prMergedEventMatches: IPRMergedEventMatch[];
    tagCreateEventMatches: ITagCreateEventMatch[];

    createTime:string;
    modifiedTime: string;
    globalProperties?: GlobalProperty[];
    globalParams?: { [key: string]: any };

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
    prMergedEventMatches: IPRMergedEventMatch[];
    tagCreateEventMatches: ITagCreateEventMatch[];

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
    prMergedEventMatches: IPRMergedEventMatch[];
    tagCreateEventMatches: ITagCreateEventMatch[];
    globalProperties?: GlobalProperty[];
    globalParams?: { [key: string]: any };
    }> {
}

export interface IJobUpdateVo extends Readonly<{
    name: string;
    testFlowId: string;
    description: string;
    versions:  any;
   
    cronExpression: string;
    prMergedEventMatches: IPRMergedEventMatch[];
    tagCreateEventMatches: ITagCreateEventMatch[];
    globalProperties?: GlobalProperty[];
    globalParams?: { [key: string]: any };
    }> {
}