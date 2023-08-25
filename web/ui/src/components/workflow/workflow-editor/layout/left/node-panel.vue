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
      <ElUpload :on-change="onUploadChange"  v-model:file-list="fileList" :auto-upload="false" :show-file-list="false" :multiple="true">➕</ElUpload>
    </div>
    <jm-scrollbar>
      <div class="groups" v-show="nodeCount>0">
        <node-group
          ref="nodeGroup1"
          :type="PluginTypeEnum.Deploy" :keyword="tempKeyword" @get-node-count="getNodeCount"/>
        <node-group
        ref="nodeGroup2"
          :type="PluginTypeEnum.Exec" :keyword="tempKeyword" @get-node-count="getNodeCount"/>
      </div>
      <div class="empty" v-if="nodeCount<=0">
        <jm-empty description="没有搜到相关结果" :image="noDataImage">
        </jm-empty>
      </div>
    </jm-scrollbar>
  </div>
  <div v-if="!isUploadCancel" class="jm-workflow-add-plugin-panel">
    <div class="item-wrap">
      <div :class="{'file-item-wrap':true}" v-for=" item,index in fileList" :key="index">
      <div :style="{width: '400px'}">
        <div :class="{'file-item':true}">
        <div class="name">{{ item.name }}</div>
        <div :class="{'close':true}"></div>
       </div>
      </div>
       <div>
        <ElSelect class="tag-select" size="small">
        </ElSelect>
       </div>
      </div>
    </div>
    <div>
      <ElButton type="info" @click="onUploadCancel">取消</ElButton>
      <ElButton type="primary" @click="onUpload">上传</ElButton>
    </div>
  </div>
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
import { ElButton, ElMessageBox, ElSelect, ElUpload, UploadUserFile } from 'element-plus';
import { uploadPlugin } from '@/api/plugin';

export default defineComponent({
  components: { NodeGroup, ElButton, ElUpload, ElSelect },
  emits: ['node-selected'],
  setup(props, { emit }) {
    const fileList: Ref<UploadUserFile[]> = ref([]);
    const isUploadCancel: Ref<boolean> = ref(true);

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
    const onUpload = async() => {
      console.log(fileList.value.length);

      try {
        // uploading.value = true;
        if (fileList.value.length > 0) {
          const formData = new FormData(); 
          fileList.value.forEach(file => {
            if (file.raw) formData.append('plugins', file.raw);
            
          });
          await uploadPlugin(formData);
          await nodeGroup1.value.loadNodes(tempKeyword.value, false);
          await nodeGroup2.value.loadNodes(tempKeyword.value, false);

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
        // uploading.value = false;
        fileList.value = [];
      }
    };
    const onUploadCancel = () => {
      isUploadCancel.value = true;
      fileList.value = [];
    };
    const onUploadChange = () => {
      isUploadCancel.value = false;
    };
    // 确定容器宽度
    onMounted(() => {
      // 初始化dnd
      workflowDnd = new WorkflowDnd(
        getGraph(),
        getWorkflowValidator(),
        container.value! as HTMLElement,
        (nodeId: string) => emit('node-selected', nodeId));
    });
    return {
      nodeGroup1,
      nodeGroup2,
      isUploadCancel,
      fileList,
      noDataImage,
      nodeCount,
      getNodeCount,
      PluginTypeEnum,
      collapsed,
      keyword,
      tempKeyword,
      container,
      onUpload,
      onUploadCancel,
      onUploadChange,
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
  display: grid;
  grid-template-columns: @node-panel-width calc(100vw - @node-panel-width);
  z-index: 2;
  height: calc(100% - @node-panel-top);
  top: @node-panel-top;
  left: 0;

  transition: grid-template-columns 0.3s ease-in-out;

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
  // 折叠动画
  transition: width 0.3s ease-in-out;
  width: @node-panel-width;
  border: 1px solid #E6EBF2;
  background: #FFFFFF;
  height: 100%;

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
      // 反转
      transform: scaleX(-1);
      border-radius: 50% 0 0 50%;
      right: calc(-@collapse-btn-width * 2 / 2);

      &::before {
        margin-left: 5.5px;
      }
    }

    .search {
      opacity: 0;
    }
  }

  .collapse-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    box-sizing: border-box;
    border: 1px solid #EBEEFB;
    z-index: 3;
    position: absolute;
    top: 78px;
    right: calc(-@collapse-btn-width / 2);

    width: @collapse-btn-width;
    height: 36px;
    line-height: 36px;
    text-align: center;
    color: #6B7B8D;
    font-size: 16px;
    background-color: #FFFFFF;
    border-radius: 50%;
    cursor: pointer;

    &::before {
      margin-left: 1.5px;
    }
  }

  .search {
    position: absolute;
    top: 0;
    width: 100%;
    transition: opacity 0.3s ease-in-out;
    padding: 30px 0 30px;
    display: flex;
    justify-content: center;
    z-index: 2;
    background-color: #FFFFFF;

    ::v-deep(.el-input) {
      width: calc(100% - 40px);
    }

    border-bottom: 1px solid #EBEEFB;

    .jm-icon-button-search {
      font-size: 16px;
      color: #7B8C9C;
    }
  }

  ::v-deep(.el-scrollbar) {
    height: calc(100% - 97px);
    margin-top: 97px;
  }

  .groups {
    width: @node-panel-width;
  }

  .empty {
    font-size: 14px;
    text-align: center;
    margin-top: 20px;

    .submit-issue {
      cursor: pointer;
      color: @primary-color;
    }
  }
}

.jm-workflow-add-plugin-panel {
  background: white;
  margin: 0px 20px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  row-gap: 10px;
  padding: 20px;
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
      font-size: 12px;
      white-space: nowrap;
      height: 100%;
      // background: yellowgreen;
      width: min-content;
      border-radius: 2px;

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
        height: 24px;
        width: 150px;
      }
    }
  }

  .file-item:hover {
    background: #f5f5f5;
  }
}
</style>
