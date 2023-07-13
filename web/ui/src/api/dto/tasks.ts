import { TaskStateEnum } from './enumeration';
import { IPageDto } from './common';
export interface ITaskVo extends Readonly<{
  id: string;
  name: string;
  podName: string;
  jobId: string;
  testflowId: string;
  testId: string;
  state: TaskStateEnum;
  logs: string[];
  versions: {};
  createTime: number;
  modifiedTime: number;
}> {
}

export interface IListTaskVo
  extends Readonly<
    IPageDto & {
      jobId: string;
    }
  > { }


export interface IGetTaskReq extends Readonly<{
}> {
  testID?: string;
  ID?: string;
}