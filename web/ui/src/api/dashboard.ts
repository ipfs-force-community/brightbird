import { restProxy } from '@/api/index';
import {
  ITaskData2WeekVo,
  ITodayPassRateVo,
  IFailureRatiobLast2WeekVo,
  ITasktPassRateLast30DaysVo,
  IJobPassCountLast30DaysVo,
  IPluginsCountVo,
  ITaskCountVo,
} from '@/api/dto/dashboard';

const dashboardApiPrefix = '/dashboard';

export const baseUrl = {
  TaskCount: '/task-count',
  TestData: '/test-data',
  TodayPassRateRanking: '/today-pass-rate-ranking',
  FailedTask: '/failed-tasks',
  PassRateTrends: '/pass-rate-trends',
  CountPlugins: '/count-plugins',
  SuccessQuantityTrends: '/success-quantity-trends',
};

// Add the prefix to the base URLs
for (const key in baseUrl) {
  if (baseUrl.hasOwnProperty(key)) {
    const propKey = key as keyof typeof baseUrl;
    baseUrl[propKey] = `${dashboardApiPrefix}${baseUrl[propKey]}`;
  }
}

// Create a generic function for making API requests
async function makeRequest<T>(url: string): Promise<T> {
  return restProxy<T>({
    url,
    method: 'get',
  });
}

export function getTaskCount(): Promise<ITaskCountVo> {
  return makeRequest<ITaskCountVo>(baseUrl.TaskCount);
}

export function getTaskData2Week(): Promise<ITaskData2WeekVo> {
  return makeRequest<ITaskData2WeekVo>(baseUrl.TestData);
}

export function getTodayPassRate(): Promise<ITodayPassRateVo> {
  return makeRequest<ITodayPassRateVo>(baseUrl.TodayPassRateRanking);
}

export function getFailureRatiobLast2Week(): Promise<IFailureRatiobLast2WeekVo> {
  return makeRequest<IFailureRatiobLast2WeekVo>(baseUrl.FailedTask);
}

export function getTasktPassRateLast30Days(): Promise<ITasktPassRateLast30DaysVo> {
  return makeRequest<ITasktPassRateLast30DaysVo>(baseUrl.PassRateTrends);
}

export function getPluginsCount(): Promise<IPluginsCountVo> {
  return makeRequest<IPluginsCountVo>(baseUrl.CountPlugins);
}

export function getJobPassCountLast30Days(): Promise<IJobPassCountLast30DaysVo> {
  return makeRequest<IJobPassCountLast30DaysVo>(baseUrl.SuccessQuantityTrends);
}
