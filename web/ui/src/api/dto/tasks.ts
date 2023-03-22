import { TaskStateEnum } from "./enumeration";

  export interface ITaskVo extends Readonly<{
    id: string
    jobId: string;
    testflowId: string
    testId: string
    State: TaskStateEnum
    logs: string[],
    versions: {},
    createTime:number,
    modifiedTime:number,
  }> {
  }