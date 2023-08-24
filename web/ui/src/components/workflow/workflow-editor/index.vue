<template>
  <div class="jm-workflow-editor">
    <template v-if="graph">
<!--      工具栏-->
      <toolbar :workflow-data="workflowData" @back="handleBack" @save="handleSave" @open-environment="onOpenEnvironment" :project-panel-visible="projectPanelVisible" />
<!--      节点配置面板：选中某个节点时，显示该节点的配置面板，允许用户进行配置-->
      <node-config-panel
        v-if="selectedNodeId"
        v-model="nodeConfigPanelVisible"
        :node-id="selectedNodeId"
        :node-waring-clicked="nodeWaringClicked"
        :workflow-data="workflowData"
        modal-class="node-config-panel-overlay"
        @closed="handleNodeConfigPanelClosed"
      />
      <environment-panel v-if="environmentPanelVisible" v-model="environmentPanelVisible" v-model:workflow-data="workflowData" ></environment-panel>
    </template>
    <div class="main">
<!--      节点面板：显示流程中全部节点，允许用户从中选择一个节点进行编辑-->
      <node-panel v-if="graph" @node-selected="nodeId => handleNodeSelected(nodeId, true)" />
<!--      图形面板：显示图形化视图，允许用户通过拖拽、连线等方式编辑工作流-->
      <graph-panel
        :workflow-data="workflowData"
        @graph-created="handleGraphCreated"
        @node-selected="nodeId => handleNodeSelected(nodeId, false)"
      />
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, getCurrentInstance, PropType, provide, ref, inject } from 'vue';
import { cloneDeep } from 'lodash';
import Toolbar from './layout/top/toolbar.vue';
import NodePanel from './layout/left/node-panel.vue';
import NodeConfigPanel from './layout/right/node-config-panel.vue';
// import CachePanel from './layout/right/cache-panel.vue';
import GraphPanel from './layout/main/graph-panel.vue';
import { IWorkflow } from './model/data/common';
import { Graph, Node } from '@antv/x6';
import registerCustomVueShape from './shape/custom-vue-shape';
import { WorkflowValidator } from './model/workflow-validator';
import EnvironmentPanel from './layout/right/environment-panel.vue';
import { GlobalProperty } from '@/api/dto/testflow';

// 注册自定义x6元素
registerCustomVueShape();

export default defineComponent({
  name: 'jm-workflow-editor',
  components: { Toolbar, NodePanel, NodeConfigPanel, GraphPanel, EnvironmentPanel },
  props: {
    modelValue: {
      type: Object as PropType<IWorkflow>,
      required: true,
    },
  },
  emits: ['update:model-value', 'back', 'save'],
  setup(props, { emit }) {
    const { proxy } = getCurrentInstance() as any;
    const workflowData = ref<IWorkflow>(cloneDeep(props.modelValue));

    if (workflowData.value.globalProperties === undefined) {
      workflowData.value.globalProperties = [JSON.parse(JSON.stringify({
        name: 'logLevel',
        type: '0',
        value:'INFO',
      }))];
    }
    const graph = ref<Graph>();
    const nodeConfigPanelVisible = ref<boolean>(false);
    const environmentPanelVisible = ref<boolean>(false);

    const selectedNodeId = ref<string>('');
    const nodeWaringClicked = ref<boolean>(false);
    let workflowValidator: WorkflowValidator;

    provide('getGraph', (): Graph => graph.value!);
    provide('getWorkflowValidator', (): WorkflowValidator => workflowValidator!);
    const projectPanelVisible = inject('projectPanelVisible') as boolean;
    const handleNodeSelected = async (nodeId: string, waringClicked: boolean) => {
      nodeConfigPanelVisible.value = true;
      selectedNodeId.value = nodeId;
      nodeWaringClicked.value = waringClicked;
    };
    return {
      workflowData,
      graph,
      nodeConfigPanelVisible,
      selectedNodeId,
      nodeWaringClicked,
      projectPanelVisible,
      environmentPanelVisible,
      handleBack: () => {
        emit('back');
      },
      handleSave: async (back: boolean, graph: string) => {
        // 必须克隆后发事件，否则外部的数据绑定会受影响
        emit('update:model-value', cloneDeep(workflowData.value));
        emit('save', back, graph);
      },
      onOpenEnvironment: () => {
        environmentPanelVisible.value = true;
      },
      handleGraphCreated: (g: Graph) => {
        workflowValidator = new WorkflowValidator(g, proxy, workflowData.value);
        graph.value = g;
      },
      handleNodeSelected,
      handleNodeConfigPanelClosed: (valid: boolean) => {
        const selectedNode = graph.value!.getCellById(selectedNodeId.value) as Node;
        if (valid) {
          workflowValidator.removeWarning(selectedNode);
        } else {
          workflowValidator.addWarning(selectedNode, nodeId => {
            handleNodeSelected(nodeId, true);
          });
        }
        // 取消选中
        graph.value!.unselect(selectedNodeId.value);
        selectedNodeId.value = '';
      },
    };
  },
});
</script>

<style lang="less">
@import './vars';

.jm-workflow-editor {
  @import './theme/x6';
  @import './theme/el';

  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  background-color: #f0f2f5;
  user-select: none;
  -moz-user-select: none;
  -webkit-user-select: none;
  -ms-user-select: none;

  .main {
    position: relative;
    z-index: 1;
    height: calc(100% - @tool-bar-height);
  }
}
</style>
