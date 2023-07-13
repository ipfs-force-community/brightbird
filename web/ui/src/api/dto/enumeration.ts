/**
 * 流程执行记录状态枚举
 */
export enum WorkflowExecutionRecordStatusEnum {
  INIT = 'INIT',
  RUNNING = 'RUNNING',
  FINISHED = 'FINISHED',
  TERMINATED = 'TERMINATED',
  SUSPENDED = 'SUSPENDED',
}

/**
 * 项目状态枚举
 */
export enum ProjectStatusEnum {
  INIT = 'INIT',
  RUNNING = 'RUNNING',
  FAILED = 'FAILED',
  SUCCEEDED = 'SUCCEEDED',
  SUSPENDED = 'SUSPENDED',
}

/**
 * 任务状态枚举
 */
export enum TaskStatusEnum {
  INIT = 'INIT',
  WAITING = 'WAITING',
  RUNNING = 'RUNNING',
  SKIPPED = 'SKIPPED',
  FAILED = 'FAILED',
  SUCCEEDED = 'SUCCEEDED',
  SUSPENDED = 'SUSPENDED',
  IGNORED = 'IGNORED',
}

/**
 * 项目导入类型枚举
 */
export enum ProjectImporterTypeEnum {
  SSH = 'SSH',
  HTTPS = 'HTTPS',
}

/**
 * DSL来源枚举
 */
export enum DslSourceEnum {
  GIT = 'GIT',
  LOCAL = 'LOCAL',
}

/**
 * DSL类型枚举
 */
export enum DslTypeEnum {
  WORKFLOW = 'WORKFLOW',
  PIPELINE = 'PIPELINE',
}

/**
 * 任务参数类型枚举
 */
export enum TaskParamTypeEnum {
  INPUT = 'INPUT',
  OUTPUT = 'OUTPUT',
}

/**
 * 触发类型枚举
 */
export enum TriggerTypeEnum {
  WEBHOOK = 'WEBHOOK',
  CRON = 'CRON',
  MANUAL = 'MANUAL',
}

/**
 * 节点类型枚举
 */
export enum NodeTypeEnum {
  LOCAL = 'LOCAL',
}


/**
 * webhook请求状态枚举
 */
export enum WebhookRequstStateEnum {
  OK = 'OK',
  NOT_ACCEPTABLE = 'NOT_ACCEPTABLE',
  UNAUTHORIZED = 'UNAUTHORIZED',
  NOT_FOUND = 'NOT_FOUND',
  UNKNOWN = 'UNKNOWN',
}

/**
 * 项目排序枚举
 */
export enum SortTypeEnum {
  DEFAULT_SORT = 'DEFAULT_SORT',
  LAST_MODIFIED_TIME = 'LAST_MODIFIED_TIME',
  LAST_EXECUTION_TIME = 'LAST_EXECUTION_TIME'
}

/**
 * 失败处理模式枚举
 */
export enum FailureModeEnum {
  IGNORE = 'IGNORE',
  SUSPEND = 'SUSPEND',
}


/**
 * job枚举类型
 */
export enum JobEnum {
  CronJob = 'cron_job',
  PRMerged = 'pr_merged_hook',
  TagCreated = 'tag_created_hook',
}


/**
 * task state枚举类型
 */
export enum TaskStateEnum {
  Init = 1,
  Running = 2,
  TempError = 3,
  Error = 4,
  Successful = 5,
}

export namespace TaskStateEnum {
  export function toString(st: TaskStateEnum): string {
    switch (st) {
      case TaskStateEnum.Init:
        return 'init';
      case TaskStateEnum.Running:
        return 'running';
      case TaskStateEnum.TempError:
        return 'temperr';
      case TaskStateEnum.Error:
        return 'error';
      case TaskStateEnum.Successful:
        return 'success';
    }
  }
}

/**
 * task state枚举类型
 */
export enum PluginTypeEnum {
  Deploy = 'Deployer',
  Exec = 'Exec',
}


export enum StepStateEnum {
  NotRunning = "notrunning",
  Running = "running",
  Success = "success",
  Fail = "fail",
}

