<template>
  <div class="project-preview-dialog">
    <jm-dialog
      :title="title"
      v-model="dialogVisible"
      width="1200px"
      @close="close"
    >
      <div class="content" v-loading="loading">
        <jm-workflow-viewer
          :dsl="dsl"
          readonly
          :node-infos="nodeDefs"/>
      </div>
    </jm-dialog>
  </div>
</template>

<script lang="ts">
import { defineComponent, getCurrentInstance, onBeforeMount, ref, SetupContext } from 'vue';
import { fetchTestFlowDetail } from '@/api/view-no-auth';
import { Node } from '@/api/dto/testflow.js';

export default defineComponent({
  props: {
    projectId: {
      type: String,
      required: true,
    },
  },
  // 覆盖dialog的close事件
  emits: ['close'],
  setup(props: any, { emit }: SetupContext) {
    const { proxy } = getCurrentInstance() as any;
    const dialogVisible = ref<boolean>(true);
    const title = ref<string>('');
    const loading = ref<boolean>(false);
    const dsl = ref<string>();
    const nodeDefs = ref<Node[]>([]);
    const close = () => emit('close');

    const loadDsl = async () => {
      if (dsl.value) {
        return;
      }

      try {
        loading.value = true;
        const { name, nodes, graph} = await fetchTestFlowDetail({id:props.projectId});
        title.value = name
        dsl.value = graph
        nodeDefs.value = nodes

      } catch (err) {
        close();

        proxy.$throw(err, proxy);
      } finally {
        loading.value = false;
      }
    };

    onBeforeMount(() => loadDsl());

    return {
      dialogVisible,
      title,
      loading,
      dsl,
      nodeDefs,
      close,
    };
  },
});
</script>

<style scoped lang="less">
.project-preview-dialog {
  ::v-deep(.el-dialog) {
    // 图标
    .el-dialog__header {
      .el-dialog__title::before {
        font-family: 'jm-icon-input';
        content: '\e803';
        margin-right: 10px;
        color: #6b7b8d;
        font-size: 20px;
        vertical-align: bottom;
        position: relative;
        top: 1px;
      }
    }

    .el-dialog__body {
      padding: 0;
    }
  }

  .content {
    height: 60vh;
  }
}
</style>