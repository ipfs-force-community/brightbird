<template>
  <div class="task-detail" v-loading="loading">
    <div class="task-list">
      <div
          class="task-item"
          v-for="task in podList"
          :key="task"
          @click="selectTask(task)"
          :class="{ 'selected': selectedPod && selectedPod.id === task }"
      >
        {{ task }}
      </div>
    </div>
    <div class="task-logs">
      <div v-if="!selectedPod">
        Please select a task to view its logs.
      </div>
      <div v-else>
        <div class="pod-log-header">{{ selectedPod.name }} Logs</div>
        <textarea class="pod-log" v-model="podLogString"></textarea>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import {defineComponent, getCurrentInstance, onMounted, provide, ref} from 'vue';
import { listAllPod, getPodLog } from '@/api/view-no-auth';
import sleep from "@/utils/sleep";
import {HttpError, TimeoutError} from "@/utils/rest/error";

export default defineComponent({
  props: {
    testId: {
      type: String,
      required: true,
      // default: '39cb76fe',
    },
  },
  computed: {
    podLogString(): string {
      return this.podLog.join('\n');
    },
  },
  setup(props: any) {
    const { proxy } = getCurrentInstance() as any;
    const podList = ref<string[]>([]);
    const selectedPod = ref<string>('');
    const podLog = ref<string[]>([]);
    const loading = ref<boolean>(false);
    console.log(props.testId);

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
        podLog.value = await getPodLog(podName);
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
        podLog.value = await getPodLog(selectedPod.value);
      }
    };

    onMounted(async () => {
      // 初始化任务列表
      await loadPodList();

      // 加载第一个 Pod 的日志
      await loadFirstPodLog();
    });

    return {
      podList,
      selectedPod,
      podLog,
      selectTask,
      loading,
    };
  },
});
</script>

<style scoped lang="less">
.task-detail {
  display: flex;
  justify-content: space-between;
  height: 100%;
  background-color: #ffffff; /* 设置背景色为白色 */

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
  height: 700px;
  width: 1050px;
  resize: none;
  border: none;
  outline: none;
  cursor: default;
}
}
}
</style>
