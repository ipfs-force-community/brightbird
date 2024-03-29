<template>
  <div class="jm-workflow-editor-toolbar">
    <div class="left">
      <button class="jm-icon-button-left" @click="goBack"></button>
      <div class="title">{{ workflowData.name }}</div>
      <ElUpload :disabled="fileList.length > 0" :on-change="onUploadChange"  v-model:file-list="fileList" :auto-upload="false" :show-file-list="false" :multiple="false">
      <div :class="{'upload':true,'disabled':fileList.length > 0}">
        <ElButton class="add-button" :disabled="fileList.length > 0">添加插件</ElButton>
      </div>
      </ElUpload>
      
    </div>
    <div class="right">
      <div class="tools">
        <jm-tooltip content="缩小" placement="bottom" :appendToBody="false">
          <button class="jm-icon-workflow-zoom-out" @click="zoom(ZoomTypeEnum.OUT)"></button>
        </jm-tooltip>
        <div class="ratio">{{ zoomPercentage }}</div>
        <jm-tooltip content="放大" placement="bottom" :appendToBody="false">
          <button class="jm-icon-workflow-zoom-in" @click="zoom(ZoomTypeEnum.IN)"></button>
        </jm-tooltip>
        <jm-tooltip content="居中" placement="bottom" :appendToBody="false">
          <button class="jm-icon-workflow-zoom-center" @click="zoom(ZoomTypeEnum.CENTER)"></button>
        </jm-tooltip>
      </div>
      <div class="operations">
        <ElButton class="env" @click="environment()">环境变量</ElButton>
        <jm-button class="save-return" @click="save(true)" @keypress.enter.prevent>保存并返回</jm-button>
        <jm-button type="primary" @click="save(false)" @keypress.enter.prevent>保存</jm-button>
      </div>
    </div>
    <project-panel v-if="projectPanelVisible" @save="save(true)" v-model="projectPanelVisible" :workflow-data="workflowData" />
  </div>
</template>

<script lang="ts">
import { computed, defineComponent, getCurrentInstance, inject, onMounted, PropType, Ref, ref  } from 'vue';
import { Cell, Graph } from '@antv/x6';
import { ZoomTypeEnum } from '../../model/data/enumeration';
import { WorkflowTool } from '../../model/workflow-tool';
import ProjectPanel from './project-panel.vue';
import { IWorkflow } from '../../model/data/common';
import { WorkflowValidator } from '../../model/workflow-validator';
import { cloneDeep } from 'lodash';
import { compare } from '../../model/util/object';
import { ElButton, ElUpload, UploadUserFile } from 'element-plus';
import { mapMutations, mapState, useStore } from 'vuex';

export default defineComponent({
  components: { ProjectPanel, ElButton, ElUpload },
  props: {
    workflowData: {
      type: Object as PropType<IWorkflow>,
      required: true,
    },
    // projectPanelVisible: {
    //   type: Boolean,
    //   default: false,
    // },
  },
  // watch:{
  //   fileList:{
  //     handler(val) {
  //       this.setFileList(val);
  //     },
  //     deep: true,
  //   },
  // },
  methods:{
    ...mapMutations('worker-editor', [
      'setFileList',
      'setUploadCancel',
    ]),
  },
  computed:{
    ...mapState('worker-editor', {
      isUploadCancel:(state:any)=>{
        return state.isUploadCancel;
      },
    }),
    fileList:{
      get(){
        return this.store.state['worker-editor']['fileList'] as UploadUserFile[];
      },
      set(val:UploadUserFile[]){
        this.setFileList(val);
      },
    },
    
  },
  emits: ['back', 'save', 'open-cache-panel', 'openEnvironment'],
  setup(props, { emit }) {
    const store = useStore();
    const { proxy } = getCurrentInstance() as any;
    let workflowBackUp = cloneDeep(props.workflowData);
    const workflowForm = ref<IWorkflow>(props.workflowData);
    const projectPanelVisible = inject('projectPanelVisible') as boolean;
    const getGraph = inject('getGraph') as () => Graph;
    const graph = getGraph();
    const getWorkflowValidator = inject('getWorkflowValidator') as () => WorkflowValidator;
    const workflowValidator = getWorkflowValidator();
    const zoomVal = ref<number>(graph.zoom());
    const cacheIconVisible = ref<boolean>(false);
    const options = ref([
      {
        value: '1',
        label: '1',
      },
      {
        value: '10',
        label: '10',
      },
      {
        value: '30',
        label: '30',
      },
      {
        value: '50',
        label: '50',
      },
      {
        value: '70',
        label: '70',
      },
      {
        value: '100',
        label: '100',
      },
    ]);
    const tooltipVisible = ref<boolean>(false);
    const concurrentVal = ref<string>();
    const concurrentRef = ref();

    const workflowTool = new WorkflowTool(graph);

    const onUploadChange = () => {
      store.commit('worker-editor/setUploadCancel', false);
    };
    return {
      store,
      onUploadChange,
      ZoomTypeEnum,
      workflowForm,
      projectPanelVisible,
      zoomPercentage: computed<string>(() => `${Math.round(zoomVal.value * 100)}%`),
      goBack: async () => {
        const originData = workflowBackUp.data ? JSON.parse(workflowBackUp.data) : {};
        const targetData: any = graph.toJSON();
        if (targetData.cells.length === 0) {
          delete targetData.cells;
        }
        workflowTool.slimGraphData(originData);
        workflowTool.slimGraphData(targetData);
        if (
          workflowBackUp.name !== workflowForm.value.name ||
          workflowBackUp.description !== workflowForm.value.description ||
          workflowBackUp.groupId !== workflowForm.value.groupId ||
          !compare(originData, targetData)
          // !compare(JSON.stringify(workflowBackUp.global), JSON.stringify(workflowForm.value.global))
        ) {
          proxy
            .$confirm(' ', '保存此次修改', {
              confirmButtonText: '保存',
              cancelButtonText: '不保存',
              distinguishCancelAndClose: true,
              type: 'info',
            })
            .then(async () => {
              try {
                await workflowValidator.checkNodes();
              } catch ({ message }) {
                proxy.$error(message);
                return;
              }
            
              workflowForm.value.graph = JSON.stringify(targetData);
              emit('save', true, workflowTool.toDsl(workflowForm.value));
            })
            .catch((action: string) => {
              if (action === 'cancel') {
                emit('back');
              }
            });
        } else {
          emit('back');
        }
      },
      edit: () => {
        projectPanelVisible.value = true;
      },
      zoom: async (type: ZoomTypeEnum) => {
        workflowTool.zoom(type);
        zoomVal.value = graph.zoom();
      },
      save: async (back: boolean) => {
        try {
          await workflowValidator.checkNodes();

          const graphData = graph.toJSON();
          workflowTool.slimGraphData(graphData);
          workflowForm.value.graph = JSON.stringify(graphData);
          emit('save', back, workflowTool.toDsl(workflowForm.value));
          workflowBackUp = cloneDeep(workflowForm.value);
        } catch ({ message }) {
          proxy.$error(message);
        }
      },
      environment: () => {
        emit('openEnvironment');
      },
      cacheIconVisible,
      tooltipVisible,
      options,
      concurrentVal,
      concurrentRef,
    };
  },
});
</script>

<style scoped lang="less">
@import '../../vars';

.jm-workflow-editor-toolbar {
  height: @tool-bar-height;
  background: #ffffff;
  z-index: 3;
  font-size: 14px;
  color: #042749;
  padding: 0 30px;

  display: flex;
  justify-content: space-between;
  align-items: center;

  button[class^='jm-icon-'] {
    border-radius: 2px;
    border-width: 0;
    background-color: transparent;
    color: #6b7b8d;
    cursor: pointer;
    text-align: center;
    width: 24px;
    height: 24px;
    font-size: 18px;

    &::before {
      font-weight: 500;
    }

    &:hover {
      background-color: #eff7ff;
      color: @primary-color;
    }
  }

  .left {
    display: flex;
    align-items: center;

    .title {
      margin-left: 20px;
      margin-right: 10px;
      max-width: 253px;
      text-overflow: ellipsis;
      overflow-x: hidden;
      white-space: nowrap;
    }
  }

  .right {
    display: flex;
    justify-content: right;
    align-items: center;
    column-gap: 15px;

    .tools {
      display: flex;
      align-items: center;

      .ratio {
        width: 40px;
        margin: 0 10px;
        text-align: center;
      }

      .jm-icon-workflow-zoom-in {
        margin-right: 10px;
      }
    }

    .cache {
      height: 20px;
      font-weight: 400;
      font-size: 14px;
      line-height: 20px;
      color: #042749;
      display: flex;
      align-items: center;
      margin: 0 0 0 50px;
      cursor: pointer;
      position: relative;

      .jm-icon-workflow-cache {
        margin-right: 6px;

        &::before {
          color: #6b7b8d;
        }
      }

      .cache-icon {
        display: flex;
        width: 12px;
        height: 12px;
        background: url('../../svgs/cache-waring.svg');
        position: absolute;
        right: -8px;
        top: -4px;
      }
    }

    .configs {
      display: flex;
      align-items: center;
      margin: 0 60px 0 44px;
      position: relative;

      ::v-deep(.el-select) {
        width: 88px;
        height: 36px;

        .el-input__icon {
          display: none;
        }

        .el-input {
          &.is-focus {
            .el-input__inner {
              border-color: #096dd9;
            }
          }

          .el-input__inner {
            border-color: #dde3ee;
          }
        }
      }

      .jm-icon-button-help {
        width: 24px;
        height: 24px;
        margin-right: 8px;
        color: #6b7b8d;
        text-align: center;
        line-height: 24px;
        font-size: 14px;
      }

      .tooltip-popper {
        width: 436px;
        height: 295px;
        padding: 16px;
        background: #ffffff;
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
        border-radius: 4px;
        box-sizing: border-box;

        position: absolute;
        top: 51px;
        right: 100px;

        .popper-description {
          font-weight: 500;
          font-size: 18px;
          color: #042749;
        }

        .concurrent-example {
          border: 10px solid #f0f2f5;
          margin-top: 8px;
        }
      }

      > div + div {
        margin-left: 10px;
      }
    }
  }

  .env,.save-return,.add-button {
        background: #f5f5f5;
        border-radius: 2px;
        color: #082340;
        border: none;
        box-shadow: none;

        &:hover {
          background: #d9d9d9;
        }
      }
}
</style>
