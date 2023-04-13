<template>
  <div class="task-detail">
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
        <pre class="pod-log">{{ podLog }}</pre>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref } from 'vue';
import { listAllPod, getPodLog } from '@/api/view-no-auth';

export default defineComponent({
  name: 'TaskDetail',
  props: {
    taskId: {
      type: String,
      required: true,
    },
  },
  setup(props: any) {
    const podList = ref<string[]>([]);
    const selectedPod = ref<string>('');
    const podLog = ref<string[]>([]);

    // 获取所有任务列表
    const loadPodList = async () => {
      try {
        const pods = await listAllPod(props.taskId);
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

    // 初始化任务列表
    loadPodList();

    return {
      podList,
      selectedPod,
      podLog,
      selectTask,
    };
  },
});
</script>

<style scoped lang="less">
.task-detail {
  display: flex;
  justify-content: space-between;
  height: 100%;

.task-list {
  width: 30%;
  border-right: 1px solid #ccc;

.task-item {
  padding: 10px;
  cursor: pointer;

&.selected {
   background-color: #f0f0f0;
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
  height: 100%;
}
}
}
</style>
