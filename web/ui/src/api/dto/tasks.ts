import { TaskStateEnum } from "./enumeration";

  export interface ITaskVo extends Readonly<{
    id: string;
    name:string;
    jobId: string;
    testflowId: string;
    testId: string;
    state: TaskStateEnum;
    logs: string[];
    versions: {};
    createTime:number;
    modifiedTime:number;
  }> {
  }

  export interface ListTaskVo extends Readonly<{
    jobId: string;
    }> {
}