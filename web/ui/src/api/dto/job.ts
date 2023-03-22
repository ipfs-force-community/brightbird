import { JobEnum } from "./enumeration"

export interface IJobIdVo
  extends Readonly<{
    id: string;
  }> {}

export interface IJobVo extends Readonly<{
    id: string
    testFlowId: string
    name: string
    jobType: JobEnum
    description: string,
    cronExpression: string,
    versions: Map<string, string>
    createTime:string,
    modifiedTime:string,
  }> {
  }

export interface IJobCreateVo extends Readonly<{
    testFlowId: string
    name: string
    jobType: JobEnum
    description: string,
    versions: Map<string, string>
    cronExpression: string,
    }> {
}

export interface IJobUpdateVo extends Readonly<{
    testFlowId: string
    name: string
    description: string,
    versions:  Map<string, string>
    cronExpression: string,
    }> {
}