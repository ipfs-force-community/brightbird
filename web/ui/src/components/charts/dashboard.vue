<template>
    <div>
        <el-row justify="center" style="margin-bottom: 20px;">
            <el-col :span="6">
                <el-row style="margin-bottom: 10px;">
                    <el-col>
                        <div style="color: rgba(16, 16, 16, 1); font-size: 20px; text-align: left;">总任务数</div>
                    </el-col>
                </el-row>
                <el-row style="margin-bottom: 10px;"> 
                    <el-col :span="3">
                        <el-icon :size="25" style="width: 37px; height: 37px; border-radius: 10%; background-color: rgb(255, 228, 186);">
                            <EditPen style="color: rgb(255, 130, 10)" />
                        </el-icon>
                    </el-col>
                    <el-col :span="3"> 
                        <div style="width: 109px; height: 40px; color: rgba(16, 16, 16, 1); font-size: 26px; text-align: left; font-weight: bold;">{{ taskCount.total }}</div>
                    </el-col>
                </el-row>
            </el-col>
            <el-col :span="6">
                <el-row style="margin-bottom: 10px;">
                    <el-col>
                        <div style="color: rgba(16, 16, 16, 1); font-size: 20px; text-align: left;">通过任务数</div>
                    </el-col>
                </el-row>
                <el-row style="margin-bottom: 10px;"> 
                    <el-col :span="3">
                        <el-icon :size="25" style="width: 37px; height: 37px; border-radius: 10%; background-color: rgb(232, 255, 251);">
                            <CircleCheck style="color: rgb(16, 198, 194)" />
                        </el-icon>
                    </el-col>
                    <el-col :span="3"> 
                        <div style="width: 109px; height: 40px; color: rgba(16, 16, 16, 1); font-size: 26px; text-align: left; font-weight: bold;">{{ taskCount.passed }}</div>
                    </el-col>
                </el-row>
            </el-col>
            <el-col :span="6">
                <el-row style="margin-bottom: 10px;">
                    <el-col>
                        <div style="color: rgba(16, 16, 16, 1); font-size: 20px; text-align: left;">失败任务数</div>
                    </el-col>
                </el-row>
                <el-row style="margin-bottom: 10px;"> 
                    <el-col :span="3">
                        <el-icon :size="25" style="width: 37px; height: 37px; border-radius: 10%; background-color: rgb(232, 243, 255);">
                            <CircleClose style="color: rgb(34, 101, 255)" />
                        </el-icon>
                    </el-col>
                    <el-col :span="3"> 
                        <div style="width: 109px; height: 40px; color: rgba(16, 16, 16, 1); font-size: 26px; text-align: left; font-weight: bold;">{{ taskCount.failed }}</div>
                    </el-col>
                </el-row>
            </el-col>
            <el-col :span="6">
                <el-row style="margin-bottom: 10px;">
                    <el-col>
                        <div style="color: rgba(16, 16, 16, 1); font-size: 20px; text-align: left;">通过率</div>
                    </el-col>
                </el-row>
                <el-row style="margin-bottom: 10px;"> 
                    <el-col :span="3">
                        <el-icon :size="25" style="width: 37px; height: 37px; border-radius: 10%; background-color: rgb(245, 232, 255);">
                            <PieChart style="color: rgb(129, 67, 214)" />
                        </el-icon>
                    </el-col>
                    <el-col :span="3"> 
                        <div style="width: 109px; height: 40px; color: rgba(16, 16, 16, 1); font-size: 26px; text-align: left; font-weight: bold;">{{ taskCount.passRate }}</div>
                    </el-col>
                </el-row>
            </el-col>
        </el-row>
        <el-row >
            <el-col :span="12" style="height: 500px;"> 
                <StackedAreaChart v-if="isTaskData2WeekReady && taskData2Week.dateArray" :testData="taskData2Week.testData" :dateArray="taskData2Week.dateArray"/>
            </el-col>
            <el-col :span="6">
                <BarChart v-if="isTodayPassRateReady" :job-names="todayPassRate.jobNames" :pass-rates="todayPassRate.passRates"/>
            </el-col>
            <el-col :span="6">
                <PieBorderRadius v-if="isFailureRatiobLast2WeekReady" :map="failureRatiobLast2Week?.failTask"/>
            </el-col>
        </el-row>
        <el-row>
            <el-col :span="8" style="height: 400px;">
                <LineGradient v-if="isTasktPassRateLast30DaysReady && tasktPassRateLast30Days.dateArray" :dateArray="tasktPassRateLast30Days.dateArray" :passRateArray="tasktPassRateLast30Days.passRateArray"/>
            </el-col>
            <el-col :span="8" style="height: 400px;">
                <PieSimple v-if="isPluginsCountReady" :deployPlugin="pluginsCount.deployerCount" :execPlugin="pluginsCount.execCount" />
            </el-col>
            <el-col :span="8" style="height: 400px;">
                <LineStack v-if="isJobPassCountLast30DaysReady && jobPassCountLast30Days.dateArray " :testData="jobPassCountLast30Days.testData" :dateArray="jobPassCountLast30Days.dateArray"/>
            </el-col>
        </el-row>
    </div>
  </template>
  
  <script lang="ts">
  import { defineComponent, onMounted, ref } from 'vue';
  import StackedAreaChart from '@/components/charts/stacked-area.vue';
  import LineGradient from '@/components/charts/line-gradient.vue';
  import PieSimple from '@/components/charts/pie-simple.vue';
  import BarChart from '@/components/charts/bar-chart.vue';
  import LineStack from '@/components/charts/line-stack.vue';
  import PieBorderRadius from '@/components/charts/pie-border-radius.vue';
  import { getTaskCount,
    getTodayPassRate, 
    getTasktPassRateLast30Days,
    getFailureRatiobLast2Week,
    getPluginsCount,
    getTaskData2Week,
    getJobPassCountLast30Days } from '@/api/dashboard';
  import { ITodayPassRateVo, 
    ITasktPassRateLast30DaysVo,
    IFailureRatiobLast2WeekVo,
    IPluginsCountVo,
    ITaskData2WeekVo,
    IJobPassCountLast30DaysVo,
    ITaskCountVo } from '@/api/dto/dashboard';
  
  export default defineComponent({
    components: { StackedAreaChart, LineGradient, PieSimple, BarChart, LineStack, PieBorderRadius },
    name: 'DashBoard',
    setup(props, { emit}) {
        const taskCount = ref<ITaskCountVo>({
            passed: 0,
            failed: 0,
            passRate: "",
            total: 0,
        })
        const todayPassRate = ref<ITodayPassRateVo>({
            jobNames: [],
            passRates: [],
        });
        const tasktPassRateLast30Days = ref<ITasktPassRateLast30DaysVo>({
            dateArray: [],
            passRateArray: [],
        });
        const failureRatiobLast2Week = ref<IFailureRatiobLast2WeekVo>({
            failTask: new Map<string, number>(),
        });
        const pluginsCount = ref<IPluginsCountVo>({
            deployerCount: 0,
            execCount: 0,
        })
        const taskData2Week = ref<ITaskData2WeekVo>({
            testData: new Map<string, number[]>(),
            dateArray: [],
        })
        const jobPassCountLast30Days = ref<IJobPassCountLast30DaysVo>({
            testData: new Map<string, number[]>(),
            dateArray: [],
        })

        const isTaskCountReady = ref(false);
        const isTodayPassRateReady = ref(false);
        const isTasktPassRateLast30DaysReady = ref(false);
        const isFailureRatiobLast2WeekReady = ref(false);
        const isPluginsCountReady = ref(false);
        const isTaskData2WeekReady = ref(false);
        const isJobPassCountLast30DaysReady = ref(false);

        onMounted(async () => {
            taskCount.value = await getTaskCount();
            isTaskCountReady.value = true;

            todayPassRate.value = await getTodayPassRate();
            isTodayPassRateReady.value = true;

            tasktPassRateLast30Days.value = await getTasktPassRateLast30Days();
            isTasktPassRateLast30DaysReady.value = true;

            failureRatiobLast2Week.value = await getFailureRatiobLast2Week();
            isFailureRatiobLast2WeekReady.value = true;
            console.log(failureRatiobLast2Week)
            console.log(failureRatiobLast2Week.value.failTask)

            pluginsCount.value = await getPluginsCount();
            isPluginsCountReady.value = true;

            taskData2Week.value = await getTaskData2Week();
            isTaskData2WeekReady.value = true;

            jobPassCountLast30Days.value = await getJobPassCountLast30Days();
            isJobPassCountLast30DaysReady.value = true;
        });
        return {
            taskCount,
            isTaskCountReady,
            todayPassRate,
            isTodayPassRateReady,
            tasktPassRateLast30Days,
            isTasktPassRateLast30DaysReady,
            failureRatiobLast2Week,
            isFailureRatiobLast2WeekReady,
            pluginsCount,
            isPluginsCountReady,
            taskData2Week,
            isTaskData2WeekReady,
            jobPassCountLast30Days,
            isJobPassCountLast30DaysReady,
        };
    },

    });
  </script>
  
  <style>
    .el-row {
        margin-bottom: 100px;
    }
    .el-row:last-child {
        margin-bottom: 0;
    }
    .el-col {
        border-radius: 4px;
    }
    
    .grid-content {
        border-radius: 4px;
        min-height: px;
    }
  </style>
  