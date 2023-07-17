export interface ITaskCountVo extends Readonly<{
    failed: number;
    passed: number;
    total: number;
    passRate: string;
}> {
}

export interface ITaskData2WeekVo extends Readonly<{
    testData: Map<string, number[]>;
    dateArray: string[];
}> {
}

export interface ITodayPassRateVo extends Readonly<{
    jobNames: string[];
    passRates: string[];
}> { }


export interface IFailureRatiobLast2WeekVo extends Readonly<{
    failTask: Map<string, number>;
}> { }

export interface ITasktPassRateLast30DaysVo extends Readonly<{
    dateArray: string[];
    passRateArray: number[];
}> { }

export interface IJobPassCountLast30DaysVo extends Readonly<{
    testData: Map<string, number[]>;
    dateArray: string[];
}> { }

export interface IPluginsCountVo extends Readonly<{
    deployerCount: number;
    execCount: number;
}> { }