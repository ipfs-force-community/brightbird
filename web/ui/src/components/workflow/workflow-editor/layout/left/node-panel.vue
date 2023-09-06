<template>
 <div :class="{'jm-workflow-editor-node-panel-wrap':true,collapsed,'uploadCancel' : isUploadCancel }">
  <div :class="{ 'jm-workflow-editor-node-panel': true, collapsed,}" ref="container">
    <div class="collapse-btn jm-icon-button-left" @click="collapse"/>
    <div class="search">
      <jm-input placeholder="搜索" v-model="keyword" @change="changeKeyword" :clearable="true">
        <template #prefix>
          <i class="jm-icon-button-search"></i>
        </template>
      </jm-input>
      <!-- <ElUpload :disabled="fileList.length > 0" :on-change="onUploadChange"  v-model:file-list="fileList" :auto-upload="false" :show-file-list="false" :multiple="false">
      <div :class="{'upload':true,'disabled':fileList.length > 0}">+</div>
      </ElUpload> -->
    </div>
    <jm-scrollbar>
      <div class="groups" v-show="nodeCount>0">
        <node-group
          ref="nodeGroup1"
          :type="PluginTypeEnum.Deploy" 
          :keyword="tempKeyword"
           @get-node-count="getNodeCount"
           @on-node-click="onNodeClick"
           />
        <node-group
        ref="nodeGroup2"
          :type="PluginTypeEnum.Exec" 
          :keyword="tempKeyword"
           @get-node-count="getNodeCount"
           @on-node-click="onNodeClick"
           />
      </div>
      <div class="empty" v-if="nodeCount<=0">
        <jm-empty description="没有搜到相关结果" :image="noDataImage">
        </jm-empty>
      </div>
    </jm-scrollbar>
  </div>
  <div
   v-loading="uploading" 
   v-if="!isUploadCancel" 
   class="jm-workflow-add-plugin-panel">
    <div class="item-wrap">
      <div :class="{'file-item-wrap':true}" v-for=" item,index in fileList" :key="index">
      <div :style="{width: '400px'}">
        <div :class="{'file-item':true}">
        <div class="name">{{ item.name }}</div>
        <div @click="onFileListDelete(index)" :class="{'close':true}"></div>
       </div>
      </div>
       <div>
        <ElSelect allow-create filterable v-model="selectLabels" class="tag-select" size="small" multiple>
          <ElOption 
          v-for="(item,index) in labels"
           :key="index" 
           :label="item" 
           :value="item"
           />
        </ElSelect>
       </div>
      </div>
    </div>
    <div>
      <ElButton type="info" @click="onUploadCancel">取消</ElButton>
      <ElButton type="primary" @click="onUpload">上传</ElButton>
    </div>
  </div>
  <ElDrawer v-if="visible" v-model="visible" size="60%">
      <PluginDetail @on-upload-success="onDeletePluginSuccess" :name="pluginNode?.getName()">
      </PluginDetail>
  </ElDrawer>
 </div>
</template>

<script lang="ts">
import { Ref, defineComponent, inject, onMounted, provide, ref } from 'vue';
import { Graph } from '@antv/x6';
import { WorkflowDnd } from '../../model/workflow-dnd';
import { WorkflowValidator } from '../../model/workflow-validator';
import NodeGroup from './node-group.vue';
import noDataImage from '../../svgs/no-data.svg';
import { PluginTypeEnum } from '@/api/dto/enumeration';
import { ElButton, ElDrawer, ElMessageBox, ElOption, ElSelect, UploadUserFile } from 'element-plus';
import { uploadPlugin } from '@/api/plugin';
import { IWorkflowNode } from '../../model/data/common';
import  PluginDetail  from '@/views/plugin-library/plugin-detail.vue';
import { mapMutations, mapState, useStore, mapActions } from 'vuex';
export default defineComponent({
  components: { NodeGroup, ElButton, ElSelect, ElDrawer, PluginDetail, ElOption },
  emits: ['node-selected'],
  methods:{
    ...mapMutations('worker-editor', [
      'setFileList',
    ]),
    ...mapActions('worker-editor', [
      'getLabels',
    ]),
    async onUpload() {
      try {
        this.uploading = true;
        if (this.fileList.length > 0) {
          const formData = new FormData(); 
          this.fileList.forEach(file => {
            if (file.raw) formData.append('plugin', file.raw);
            
          });

          this.selectLabels.forEach(value=>{
            formData.append('labels', value);
          });
          await uploadPlugin(formData);
          await this.nodeGroup1.loadNodes(this.tempKeyword, false);
          await this.nodeGroup2.loadNodes(this.tempKeyword, false);
          this.store.commit('worker-editor/setUploadCancel', true);
          ElMessageBox.alert('上传成功', '提示', {
            confirmButtonText: '确定',
            type: 'success',
          });
        } else {
          ElMessageBox.alert('文件列表为空，请先选择文件', '提示', {
            confirmButtonText: '确定',
            type: 'warning',
          });
        }
      } catch (error) {
        ElMessageBox.alert(`上传失败: ${error}`, '错误', {
          confirmButtonText: '确定',
          type: 'error',
        });
      } finally {
        this.uploading = false;
      }
    },
    onFileListDelete(idx:number){
      const object = JSON.parse(JSON.stringify(this.fileList));
      object.splice(idx, 1);
      this.setFileList(object);
    },
  },
  computed:{
    ...mapState('worker-editor', {
      isUploadCancel:(state:any)=>{
        return state.isUploadCancel;
      },
      fileList:(state:any)=>{
        return state.fileList as UploadUserFile[];
      },
      labels:state=>{
        return state.labels;
      },
    }),
  },
  mounted() {
    this.getLabels();
  },
  setup(props, { emit }) {
    const store = useStore();
    // const fileList: Ref<UploadUserFile[]> = ref([]);
    // const isUploadCancel: Ref<boolean> = ref(true);
    const visible =  ref<boolean>(false);
    const pluginNode = ref<IWorkflowNode>();
    const uploading =  ref<boolean>(false);
    const selectLabels = ref<string[]>([]);


    const nodeGroup1: Ref<any> = ref();
    const nodeGroup2: Ref<any> = ref();

    const collapsed = ref<boolean>(false);
    const keyword = ref<string>('');
    // 输入框触发change事件后传递给组件的keyword
    const tempKeyword = ref<string>('');
    const getGraph = inject('getGraph') as () => Graph;
    const getWorkflowValidator = inject('getWorkflowValidator') as () => WorkflowValidator;
    let workflowDnd: WorkflowDnd;
    provide('getWorkflowDnd', () => workflowDnd);
    const container = ref<HTMLElement>();
    // 控制节点拖拽面板是否显示
    const nodeCount = ref<number>(0);
    const getNodeCount = (count: number) => {
      if (!count) {
        return;
      }
      // 如果node-group中都找不到节点拖拽面板不展示
      nodeCount.value += count;
    };

    const onUploadCancel = () => {
      store.commit('worker-editor/setUploadCancel', true);
      store.commit('worker-editor/setFileList', []);
    };
    const onNodeClick = (item: IWorkflowNode) => {
      pluginNode.value = item;
      visible.value = true;
    };

    const onDeletePluginSuccess = async ()=>{
      visible.value = false;
      await nodeGroup1.value.loadNodes(tempKeyword.value, false);
      await nodeGroup2.value.loadNodes(tempKeyword.value, false);
      ElMessageBox.alert('删除成功', '提示', {
        confirmButtonText: '确定',
        type: 'success',
      });
    };
    // 确定容器宽度
    onMounted(() => {
      // 获取label 标签
      // 初始化dnd
      workflowDnd = new WorkflowDnd(
        getGraph(),
        getWorkflowValidator(),
        container.value! as HTMLElement,
        (nodeId: string) => emit('node-selected', nodeId));
      console.log('+++++===========', this);
      // this.getLabels();

    });
    return {
      store,
      selectLabels,
      uploading,
      visible,
      nodeGroup1,
      nodeGroup2,
      noDataImage,
      nodeCount,
      pluginNode,
      getNodeCount,
      PluginTypeEnum,
      collapsed,
      keyword,
      tempKeyword,
      container,
      onUploadCancel,
      onNodeClick,
      onDeletePluginSuccess,
      collapse: () => {
        collapsed.value = container.value!.clientWidth > 0;
      },
      changeKeyword(key: string) {
        // 解决用户在输入框中输入了关键字但未进行搜索且点击清空内容，出现未搜索到的情况
        if (tempKeyword.value === key) {
          return;
        }
        tempKeyword.value = key;
        nodeCount.value = 0;
      },
    };
  },
});
</script>

<style scoped lang="less">
@import '../../vars';

@node-panel-top: 20px;
@collapse-btn-width: 36px;

.jm-workflow-editor-node-panel-wrap {
  position: absolute;
  top: @node-panel-top;
  bottom: 0;
  left: 0;
  z-index: 2;
  display: grid;
  height: calc(100% - @node-panel-top);
  transition: grid-template-columns 0.3s ease-in-out;

  grid-template-columns: @node-panel-width calc(100vw - @node-panel-width);

  &.collapsed {
    grid-template-columns: 0px 100vw;
  }
}

.jm-workflow-editor-node-panel-wrap.uploadCancel {
  grid-template-columns: @node-panel-width;
  &.collapsed {
    grid-template-columns: 0px;
  }
}

.jm-workflow-editor-node-panel {
  position: relative;
  overflow-y: auto;
  width: @node-panel-width;
  height: 100%;
  border: 1px solid #E6EBF2;
  background: #FFFFFF;
  // 折叠动画
  transition: width 0.3s ease-in-out;

  .search {
    display: flex;
    align-items: center;

    &>:last-child {
      margin-left: 10px;
    }
  }

  &.collapsed {
    width: 0;
    .collapse-btn {
      right: calc(-@collapse-btn-width * 2 / 2);
      border-radius: 50% 0 0 50%;
      // 反转
      transform: scaleX(-1);

      &::before {
        margin-left: 5.5px;
      }
    }

    .search {
      opacity: 0;
    }
  }

  .collapse-btn {
    position: absolute;
    top: 78px;
    right: calc(-@collapse-btn-width / 2);
    z-index: 3;
    display: flex;
    align-items: center;
    justify-content: center;
    box-sizing: border-box;
    width: @collapse-btn-width;
    height: 36px;
    border: 1px solid #EBEEFB;
    border-radius: 50%;
    background-color: #FFFFFF;
    color: #6B7B8D;
    text-align: center;
    font-size: 16px;
    line-height: 36px;
    cursor: pointer;

    &::before {
      margin-left: 1.5px;
    }
  }

  .search {
    position: absolute;
    top: 0;
    z-index: 2;
    display: flex;
    justify-content: center;
    padding: 30px 0 30px;
    width: 100%;
    border-bottom: 1px solid #EBEEFB;
    background-color: #FFFFFF;
    transition: opacity 0.3s ease-in-out;

    ::v-deep(.el-input) {
      width: calc(100% - 40px);
    }

    .jm-icon-button-search {
      color: #7B8C9C;
      font-size: 16px;
    }
  }

  ::v-deep(.el-scrollbar) {
    margin-top: 97px;
    height: calc(100% - 97px);
  }

  .groups {
    width: @node-panel-width;
  }

  .empty {
    margin-top: 20px;
    text-align: center;
    font-size: 14px;

    .submit-issue {
      color: @primary-color;
      cursor: pointer;
    }
  }
}

.jm-workflow-editor-node-panel.collapsed {
  overflow: visible;
}

.jm-workflow-add-plugin-panel {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  margin: 0px 20px;
  padding: 20px;
  background: white;

  row-gap: 10px;
  .item-wrap {
    display: flex;
    flex-direction: column;

    row-gap: 10px;
    .file-item-wrap {
      display: flex;
      width: min-content;
    .file-item {
      display: flex;
      align-items: center;
      justify-content: flex-start;
      // background: yellowgreen;
      width: min-content;
      height: 100%;
      border-radius: 2px;
      white-space: nowrap;
      font-size: 12px;

      .close {
        display: none;
      }

      &:hover {
        .close {
          display: block;
        }
      }
      .close {
        width: 15px;
        height: 15px;
        background-image: url('@/assets/svgs/btn/cancel-grey.svg');
      }
      .close:hover {
        background-image: url('@/assets/svgs/btn/cancel-blue.svg');
      }
    }

    :deep(.el-input--small) {
        // height: 24px;
        width: 150px;
        .disabled {
          cursor: not-allowed;
        }
      }
    }

    // :deep(.el-loading-spinnerbefore){
    //   font-size: 20px;
    // }  
  }
  
  .file-item:hover {
    background: #f5f5f5;
  }
}

:deep(.upload.disabled) {
      cursor: not-allowed !important;
}
::v-deep(.el-loading-spinner) {
        &::before {
          display: inline-block;
          color: @primary-color;
          content: '\e806';
          font-size: 30px !important;
          font-family: 'jm-icon-button';
          animation: rotating 2s linear infinite;
        }

        .circular {
          display: none;
        }
    }
</style>
