import { StepStateEnum } from "./enumeration";


export interface StepLog extends Readonly<{
    name: string;
    instanceName: string;
    state: StepStateEnum;
    logs: string[];
}> { }

export interface LogResp extends Readonly<{
    podName: string;
    steps: StepLog[];
    logs: string[];
}> { }

export interface LogReq extends Readonly<{
    podName: string;
    testID: string;
}> { }