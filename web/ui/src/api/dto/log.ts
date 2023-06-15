

export interface StepLog extends Readonly<{
    name: string;
    logs: string[];
    isSuccess: boolean;
}> { }


export interface LogResp extends Readonly<{
    podName: string;
    steps: StepLog[];
    logs: string[];
}> { }