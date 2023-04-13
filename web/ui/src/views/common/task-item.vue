<template>
  <div class="task-item">
    <div :class="{
      'state-bar': true,
      [task.state.toLowerCase()]: true,
    }"></div>
    <div class="content">
      <div class="content-top">
<!--        <router-link-->
<!--            :to="{-->
<!--            name: 'task-detail',-->
<!--            query: { taskId: task.id },-->
<!--          }"-->
<!--        >-->
        <jm-text-viewer :value="task.name" :class="{ title: true }" />
<!--        </router-link>-->
      </div>

      <div class="content-center">
        <jm-text-viewer class="status" :value="`${task.state || '无'}`" />
      </div>

      <div class="content-bottom">
        <div class="operation">
          <jm-tooltip content="停止" placement="bottom">
            <button class="del" @click="stopTask(task.id)"></button>
          </jm-tooltip>
        </div>
      </div>
    </div>
    <div class="cover"></div>
  </div>
</template>

<script lang="ts">
import { defineComponent, getCurrentInstance, PropType, ref, SetupContext } from 'vue';
import JmTextViewer from "@/components/text-viewer/index.vue";
import { ITaskVo } from '@/api/dto/tasks';

export default defineComponent({
  components: { JmTextViewer },
  props: {
    task: {
      type: Object as PropType<ITaskVo>,
      required: true,
    },
  },
  emits: [],
  setup(props: any, { emit }: SetupContext) {
    const { proxy } = getCurrentInstance() as any;
    const stopTask = async (id: string) => {

    }
    return {
      stopTask,
      props,
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

          &.del {
            background-image: url('@/assets/svgs/btn/del.svg');
          }
        }

        &.webhook {
          background-image: url('@/assets/svgs/btn/hook.svg');
        }

        &.git-label {
          background-image: url('@/assets/svgs/index/git-label.svg');
        }

        &.workflow-label {
          background-image: url('@/assets/svgs/index/workflow-label.svg');
        }

        &.pipeline-label {
          background-image: url('@/assets/svgs/index/pipeline-label.svg');
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
          background-image: url('@/assets/svgs/btn/more2.svg');
          transform: rotate(90deg);
        }
      }
    }
  }
}
</style>
