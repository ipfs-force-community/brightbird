

export interface StepLog extends Readonly<{
    name: string;
    logs: string[];
    success: boolean;
}> { }


export interface LogResp extends Readonly<{
    podName: string;
    steps: StepLog[];
    logs: string[];
}> { }