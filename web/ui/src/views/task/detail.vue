<template>
    <div class="nav">
    <button class="jm-icon-button-left" @click="goBack"></button>
   </div>
  <div class="task-detail" v-loading="loading">
    <div class="task-list">
      <div class="task-item" v-for="pod in podList" :key="pod" @click="selectTask(pod)"
        :class="{ 'selected': selectedPod && selectedPod === pod }">
        {{ pod }}
      </div>
    </div>
    <div class="task-logs">
      <div v-if="!selectedPod">
        Please select a task to view its logs.
      </div>
      <div v-else>
        <div class="pod-log-header">{{ selectedPod }} Logs</div>
        <el-collapse v-if="selectedPod.indexOf('test-runner') > -1" class="pod-log  pod-height">
          <el-collapse-item v-for="step,index in podLog?.steps" :key="index" class="step-header">
            <template #title>
              <div class="header-title">
                <el-icon v-if="step.state == StepStateEnum.Success" size="20">
                  <SuccessFilled color="green" />
                </el-icon>

                <el-icon v-if="step.state == StepStateEnum.Fail" size="20">
                  <CircleCloseFilled color="rgb(255, 168, 168)" />
                </el-icon>

                <el-icon v-if="step.state == StepStateEnum.NotRunning" size="20">
                  <RemoveFilled color="gray" />
                </el-icon>

                <el-icon v-if="step.state == StepStateEnum.Running" size="20">
                  <QuestionFilled color="green" />
                </el-icon>

                <el-text class="title" size="small"> {{ step.name }} </el-text>
              </div>
            </template>
            <div class="steps">
              <el-text class="step-item" size="small" tag="p" v-for="log,index in step.logs" :key="index"> {{ log }} </el-text>
            </div>
          </el-collapse-item>
        </el-collapse>
        <el-input v-else class="pod-log pod-height" v-model="podLogString" aria-readonly="true" :autosize="{ minRows: 2 }"
          type="textarea" />
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, getCurrentInstance, onMounted, provide, ref } from 'vue';
import { listAllPod, getPodLog } from '@/api/log';
import { HttpError, TimeoutError } from '@/utils/rest/error';
import { LogResp } from '@/api/dto/log';
import { StepStateEnum } from '@/api/dto/enumeration';
import { useRouter } from 'vue-router';

export default defineComponent({
  props: {
    testId: {
      type: String,
      required: true,
    },
  },
  computed: {
    podLogString(): string {
      return this.podLog?.logs.join('\n') ?? '';
    },
  },
  setup(props: any) {
    const { proxy } = getCurrentInstance() as any;
    const podList = ref<string[]>([]);
    const selectedPod = ref<string>('');
    const podLog = ref<LogResp>();
    const loading = ref<boolean>(false);
    const router = useRouter();

    // 获取所有任务列表
    const loadPodList = async () => {
      try {
        const pods = await listAllPod(props.testId);
        podList.value = pods;
      } catch (error) {
        console.error(error);
      }
    };

    // 选中任务并获取其日志
    const selectTask = async (podName: string) => {
      try {
        selectedPod.value = podName;
        podLog.value = await getPodLog({ podName: podName, testID: props.testId });
      } catch (error) {
        console.error(error);
      }
    };

    const loadData = async (refreshing?: boolean) => {
      try {
        await proxy.listAllPod({
          testId: props.testId,
        });
      } catch (err) {
        if (!refreshing) {
          throw err;
        }

        if (err instanceof TimeoutError) {
          // 忽略超时错误
          console.warn(err.message);
        } else if (err instanceof HttpError) {
          const { response } = err as HttpError;

          if (response && response.status !== 502) {
            throw err;
          }

          // 忽略错误
          console.warn(err.message);
        }
      }
    };

    provide('loadData', loadData);
    // 初始化任务列表
    const loadFirstPodLog = async () => {
      if (podList.value.length > 0) {
        selectedPod.value = podList.value[0];
        podLog.value = await getPodLog({ podName:selectedPod.value, testID:props.testId });
      }
    };

    const goBack = () => {
      router.back();
    };

    onMounted(async () => {
      // 初始化任务列表
      await loadPodList();

      // 加载第一个 Pod 的日志
      await loadFirstPodLog();
    });

    return {
      StepStateEnum,
      loading,
      podList,
      selectedPod,
      podLog,
      selectTask,
      goBack,
    };
  },
});
</script>

<style scoped lang="less">
@primary-color: #096dd9;

.nav {
  position: fixed;
  z-index: 101;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 30px;
  width: 100%;
  height: 64px;
  background: white;
  color: #042749;
  button[class^="jm-icon-"] {
    width: 24px;
    height: 24px;
    border-width: 0;
    border-radius: 2px;
    background-color: transparent;
    color: #6b7b8d;
    text-align: center;
    font-size: 18px;
    cursor: pointer;

    &::before {
      font-weight: 500;
    }

    &:hover {
      background-color: #eff7ff;
      color: @primary-color;
    }
  }
}


.task-detail {
  padding-top: 64px;
  display: flex;
  justify-content: space-between;
  background-color: #ffffff;
  /* 设置背景色为白色 */

  .task-list {
    width: 30%;
    border-right: 1px solid #ccc;
    background-color: #f5f5f5;
    padding: 10px;
    overflow: auto;

    &::-webkit-scrollbar {
      width: 8px;
      height: 8px;
      background-color: #f5f5f5;
    }

    &::-webkit-scrollbar-thumb {
      background-color: #ccc;
      border-radius: 4px;
    }

    &::-webkit-scrollbar-track {
      background-color: #f5f5f5;
    }
  }

  .task-item {
    padding: 10px;
    cursor: pointer;
    margin-bottom: 10px;
    background-color: #fff;
    border-radius: 4px;
    box-shadow: 0 0 5px rgba(0, 0, 0, 0.1);

    &:hover {
      background-color: #f5f5f5;
    }

    &.selected {
      background-color: #e1e1e1;

      &:hover {
        background-color: #e1e1e1;
      }
    }
  }

  .task-logs {
    flex-grow: 1;
    padding: 10px;

    .pod-log-header {
      font-weight: bold;
      margin-bottom: 10px;
    }

    .pod-log {
      font-family: monospace;
      white-space: pre-wrap;
      background-color: #f0f0f0;
      padding: 10px;
      overflow: auto;
      width: 1050px;
      resize: none;
      border: none;
      outline: none;
      cursor: default;
      --el-collapse-header-bg-color: rgb(154, 154, 154);
      --el-collapse-content-bg-color: #e1e1e1;
      --el-collapse-border-color: rgb(154, 154, 154);
      --el-collapse-header-height: 30px;

      .step-header {
        margin-left: 8px;
        margin-bottom: 3px;
        border-radius: 8px;
        border: 3px solid rgb(154, 154, 154);

        .header-title {
          display: inline-flex;
          align-items: center;

          .title {
            color: rgb(255, 255, 255);
            padding-left: 10px;
            font-size: 15px;
          }
        }
      }

      .steps {
        .step-item {
          padding-left: 10px;
          word-break: break-word;
        }
      }
    }

    @media (max-width: 1800px) {
      .pod-height {
        height: 760px;
      }
    }

    @media (min-width: 1805px) {
      .pod-height {
        height: 1050px;
      }
    }
  }
}</style>
