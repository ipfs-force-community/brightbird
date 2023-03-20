<template>
  <div class="project-item">
    <div class="state-bar"></div>
    <div class="content">
      <div class="content-top">
        <router-link
            :to="{
            name: 'workflow-execution-record-detail',
            query: { projectId: project.id },
          }"
        >
          <jm-text-viewer :value="project.name" :class="{ title: true}" />
        </router-link>
      </div>

      <div class="content-center">
        <jm-text-viewer class="status" :value="`${project.description || '无'}`"/>
      </div>

      <div class="content-bottom">
        <div class="operation">
          <jm-tooltip content="编辑" placement="bottom">
            <button class="edit" @click="edit(project.id)"></button>
          </jm-tooltip>
          <jm-tooltip content="删除" placement="bottom">
            <button class="del" @click="del(project.id)"></button>
          </jm-tooltip>
        </div>
      </div>
    </div>
    <cache-drawer
      v-if="cacheDrawerFlag"
      v-model="cacheDrawerFlag"
      :current-project-name="project.name"
      :current-project-workflow-ref="project.workflowRef"
    ></cache-drawer>
    <project-preview-dialog v-if="dslDialogFlag" :project-id="project.id" @close="dslDialogFlag = false" />
    <div class="cover"></div>
  </div>
</template>

<script lang="ts">
import { defineComponent, getCurrentInstance, PropType, ref, SetupContext } from 'vue';
import { IProjectVo } from '@/api/dto/project';
import { del } from '@/api/project';
import ProjectPreviewDialog from './project-preview-dialog.vue';
import WebhookDrawer from './webhook-drawer.vue';
import CacheDrawer from '@/views/common/cache-drawer.vue';
import { useRouter } from 'vue-router';
import JmTextViewer from "@/components/text-viewer/index.vue";

export default defineComponent({
  components: {JmTextViewer, ProjectPreviewDialog, WebhookDrawer, CacheDrawer },
  props: {
    project: {
      type: Object as PropType<IProjectVo>,
      required: true,
    },
    // 控制item是否加上hover样式，根据对比id值判断
    move: {
      type: Boolean,
      default: false,
    },
    // 控制是否处于拖拽模式
    moveMode: {
      type: Boolean,
      default: false,
    },
  },
  emits: ['triggered', 'synchronized', 'deleted', 'terminated'],
  setup(props: any, { emit }: SetupContext) {
    const { proxy } = getCurrentInstance() as any;
    const router = useRouter();
    const deleting = ref<boolean>(false);
    const dslDialogFlag = ref<boolean>(false);  //todo replace with config
    const cacheDrawerFlag = ref<boolean>(false);
    return {
      deleting,
      dslDialogFlag,
      cacheDrawerFlag,
      edit: (id: string) => {
        router.push({ name: 'update-pipeline', params: { id } });
      },
      del: (id: string) => {
        if (deleting.value) {
          return;
        }
        //todo delete
      },
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

.project-item {
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

  &.move {
    position: relative;
    cursor: move;

    .cover {
      display: block;
      position: absolute;
      box-sizing: border-box;
      width: 100%;
      height: 100%;
      border: 2px solid #096dd9;
      background-color: rgba(140, 140, 140, 0.3);
      top: 0;
      left: 0;

      &::after {
        content: '';
        position: absolute;
        bottom: 0;
        right: 0;
        display: inline-block;
        width: 30px;
        height: 30px;
        background-image: url('@/assets/svgs/sort/drag.svg');
        background-repeat: no-repeat;
      }
    }
  }

  .cover {
    display: none;
  }

  .state-bar {
    height: 6px;
    overflow: hidden;
    background-color: #3ebb03;
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
      display: flex;
      align-items: center;

      a {
        flex: 1;
      }

      .concurrent {
        height: 20px;
        line-height: 20px;
        background: #fff7e6;
        border-radius: 2px;
        padding: 3px;
        font-size: 12px;
        font-weight: 400;
        color: #6d4c41;
        margin-right: 5px;
      }

      .alarm {
        width: 24px;
        height: 24px;
        background: url('@/assets/svgs/index/alarm.svg') 100% no-repeat;
      }
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

        button + button {
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

          &.execute {
            background-image: url('@/assets/svgs/btn/execute.svg');
          }

          &.edit {
            background-image: url('@/assets/svgs/btn/edit.svg');
          }

          &.sync {
            background-image: url('@/assets/svgs/btn/sync.svg');

            &.doing {
              animation: rotating 2s linear infinite;
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
}
</style>