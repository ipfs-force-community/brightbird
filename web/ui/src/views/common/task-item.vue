<template>
  <div class="task-item">
    <div :class="{
      'state-bar': true,
      [formatState(task.state).toLowerCase()]: true,
    }"></div>
    <div class="content">
      <router-link :to="{
        name: 'task-detail',
        query: { testId: task.testId, id: task.id },
      }">
        <div class="content-top">
          <jm-text-viewer :value="task.name" :class="{ title: true }" />
        </div>
      </router-link>
      <div class="content-center">
        <jm-text-viewer class="log" v-for="(log, index) of latestlog(task.logs)" :key="index" :value="`${log}`" />
      </div>

      <div class="content-bottom">
        <span class="podname">{{ task.podName }}</span>
        <div v-if="task.state == TaskStateEnum.Running" class="operation">
          <jm-tooltip content="停止运行" placement="bottom">
            <button class="cancel" @click="cancelTask(task)"></button>
          </jm-tooltip>
        </div>
        <div v-if="task.state == TaskStateEnum.Error" class="operation">
          <jm-tooltip content="重试" placement="bottom">
            <button class="retry" @click="retryTaskOnce(task.id)"></button>
          </jm-tooltip>
        </div>
      </div>
    </div>
    <div class="cover"></div>
  </div>
</template>

<script lang="ts">
import {
  defineComponent,
  getCurrentInstance,
  PropType,
  SetupContext,
  ref,
} from 'vue';
import JmTextViewer from '@/components/text-viewer/index.vue';
import { ITaskVo } from '@/api/dto/tasks';
import { TaskStateEnum } from '@/api/dto/enumeration';
import { stopTask, retryTask} from '@/api/tasks';


export default defineComponent({
  components: { JmTextViewer},
  props: {
    task: {
      type: Object as PropType<ITaskVo>,
      required: true,
    },
  },
  emits: [],
  setup(props: any, { emit }: SetupContext) {
    const { proxy } = getCurrentInstance() as any;

    const task = ref<ITaskVo>(props.task);
    const cancelTask = async (task: ITaskVo) => {
      try {
        await stopTask(task.id);
        proxy.$success('stop task success');
      } catch (err) {
        proxy.$throw(err, proxy);
      }
    };

    const retryTaskOnce = async (id: string) => {
      try {
        await retryTask(id);
        task.value = Object.assign(task.value, {state: TaskStateEnum.Init});
        proxy.$success('retry task success');
      } catch (err) {
        proxy.$throw(err, proxy);
      }
    };

    const formatState = (state: TaskStateEnum): string =>
      TaskStateEnum.toString(state);

    const latestlog = (log: string[]): string[] => {
      return log.slice(log.length - 3).reverse();
    };

    return {
      TaskStateEnum,
      cancelTask,
      retryTaskOnce,
      formatState,
      latestlog,
      task,
    };
  },
});
</script>

<style scoped lang="less">
@keyframes workflow-running {
  0% {
    background-position-x: -53.5px;
  }

  100% {
    background-position-x: 0;
  }
}

@-webkit-keyframes workflow-running {
  0% {
    background-position-x: -53.5px;
  }

  100% {
    background-position-x: 0;
  }
}

.task-item {
  box-sizing: border-box;
  margin: 0.8% 0.8% 0 0;
  width: 19.2%;
  min-width: 260px;
  background-color: #ffffff;
  min-height: 166px;
  border-radius: 0px 0px 4px 4px;

  &:hover {
    box-shadow: 0px 0px 12px 4px #edf1f8;

    .content {
      border: 1px solid transparent;
      border-top: none;
    }
  }

  .cover {
    display: none;
  }

  .state-bar {
    height: 6px;
    overflow: hidden;

    &.init {
      background-color: #979797;
    }

    &.running {
      background-image: repeating-linear-gradient(115deg,
          #10c2c2 0px,
          #58d4d4 1px,
          #58d4d4 10px,
          #10c2c2 11px,
          #10c2c2 16px);
      background-size: 106px 114px;
      animation: 3s linear 0s infinite normal none running workflow-running;
    }

    &.building {
      background-color: #ad82f7;
    }
    &.success {
      background-color: #3ebb03;
    }

    &.temperr {
      background-color: #e0b818;
    }

    &.error {
      background-color: #cf1524;
    }
  }

  .content {
    min-height: 80px;
    position: relative;
    padding: 16px 20px 10px 20px;
    border: 1px solid #dee4eb;
    border-top: none;
    border-radius: 0px 0px 4px 4px;

    .content-top {
      min-height: 24px;
      // display: flex;
      align-items: center;
    }

    .content-center {
      .log {
        font-size: 10px;
      }

      .status {
        margin-top: 10px;
        display: flex;
        align-items: center;
        justify-content: space-between;
        font-size: 14px;
        color: #082340;
        font-weight: 400;
      }
    }

    .content-bottom {
      padding: 10px 0 0;
      border-top: 1px solid #dee4eb;
      display: flex;
      align-items: center;
      justify-content: space-between;

      .podname {
        font-size: 10px;
      }

      .operation {
        min-height: 26px;
        display: flex;
        align-items: center;

        button+button {
          margin-left: 20px;
        }

        button {
          width: 26px;
          height: 26px;
          background-color: transparent;
          border: 0;
          background-position: center center;
          background-repeat: no-repeat;
          cursor: pointer;
          outline: none;

          &:active {
            background-color: #eff7ff;
            border-radius: 4px;
          }

          &.cancel {
            background-image: url("@/assets/svgs/btn/cancel.svg");
          }

          &.retry {
            background-image: url("@/assets/svgs/btn/refresh.svg");
          }
        }

        &.webhook {
          background-image: url("@/assets/svgs/btn/hook.svg");
        }

        &.git-label {
          background-image: url("@/assets/svgs/index/git-label.svg");
        }

        &.workflow-label {
          background-image: url("@/assets/svgs/index/workflow-label.svg");
        }

        &.pipeline-label {
          background-image: url("@/assets/svgs/index/pipeline-label.svg");
        }

        &.doing {
          opacity: 0.5;
          cursor: not-allowed;

          &:active {
            background-color: transparent;
          }
        }
      }
    }

    .more {
      opacity: 0.65;
      cursor: pointer;

      &:hover {
        opacity: 1;
      }

      .el-dropdown-link {
        .btn-group {
          &:active {
            background-color: #eff7ff;
            border-radius: 4px;
          }

          width: 24px;
          height: 24px;
          background-image: url("@/assets/svgs/btn/more2.svg");
          transform: rotate(90deg);
        }
      }
    }
  }
}
</style>
