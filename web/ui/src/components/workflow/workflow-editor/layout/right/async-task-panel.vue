<template>
  <div class="jm-workflow-editor-async-task-panel">
    <jm-form label-width="auto" :model="form" label-position="top" ref="formRef" @submit.prevent>
      <div class="set-padding">
        <div class="name-item">
          <el-text size="large" tag="b">{{ form.name }}</el-text>
        </div>

        <jm-form-item label="实例名称" prop="instanceName" class="name-item" :rules="nodeData.getFormRules().instanceName">
          <jm-input v-model="form.instanceName" show-word-limit :maxlength="36" />
        </jm-form-item>
        <jm-form-item label="节点版本" prop="version" :rules="nodeData.getFormRules().version" class="node-item">
          <jm-select v-loading="versionLoading" :disabled="versionLoading" v-model="form.version" placeholder="请选择节点版本"
            @change="changeVersion">
            <jm-option v-for="item in versionList.versions" :key="item" :label="item" :value="item" />
          </jm-select>
          <div v-if="form.version ? !versionLoading : false" class="version-description">
            {{ form.version }}
          </div>
        </jm-form-item>
      </div>
      <div class="separate"></div>
      <div v-if="form.version">
        <div class="tab-container">
          <div :class="{ 'input-tab': true, 'selected-tab': tabFlag === 1 }" @click="tabFlag = 1">
            输入参数
            <div class="checked-underline" v-if="tabFlag === 1"></div>
          </div>
        </div>
        <div class="inputs-container set-padding" v-if="tabFlag === 1">
          <div v-if="!input">
            <jm-empty description="无输入参数" :image="noParamImage"></jm-empty>
          </div>
          <div v-else>
            <jm-form-item v-for="(value, key) in input.properties" :key="key" :prop="`inputs.${key}.value`"
              class="node-name">
              <template #label>
                {{value.title}}
                <jm-tooltip placement="top" v-if="value.description" :append-to-body="false" :content="value.description">
                  <i class="jm-icon-button-help"></i>
                </jm-tooltip>
              </template>
              <PropertySelect :instanceName="form.instanceName" :propName="key" :input="form.input" :property="value" :treeData="nodeNames">
              </PropertySelect>
            </jm-form-item>
          </div>
        </div>
      </div>
    </jm-form>
  </div>
</template>

<script lang="ts">
import { defineComponent, getCurrentInstance, inject, onMounted, PropType, ref } from 'vue';
import { AsyncTask } from '../../model/data/node/async-task';
import { ParamTypeEnum } from '../../model/data/enumeration';
import noParamImage from '../../svgs/no-param.svg';
import { INodeDefVersionListVo } from '@/api/dto/node-definitions';
import ExpressionEditor from './form/expression-editor.vue';
import { Graph, Node } from '@antv/x6';
import JmEmpty from '@/components/data/empty/index.vue';
import JmForm from '@/components/form/form';
import jmFormItem from '@/components/form/form-item';
import JmInput from '@/components/form/input';
import { getPluginByName, getPluginDef } from '@/api/plugin';
import { PluginDef } from '@/api/dto/testflow';
import PropertySelect from './property-select.vue'
import { CustomX6NodeProxy } from '@/components/workflow/workflow-editor/model/data/custom-x6-node-proxy';
import { TreeProp } from '@/components/workflow/workflow-editor/model/data/common';
import JmSelect from '@/components/form/select';
import {Try} from "json-schema-to-typescript/dist/src/utils";
import { JSONSchema } from 'json-schema-to-typescript';
import { schema } from '@antv/g2plot';

export default defineComponent({
  components: { JmEmpty, ExpressionEditor, PropertySelect, JmForm, jmFormItem, JmInput, JmSelect },
  props: {
    nodeData: {
      type: Object as PropType<AsyncTask>,
      required: true,
    },
    caches: {
      type: [Array, String],
    },
  },
  emits: ['form-created'],
  setup(props, { emit }) {
    const { proxy } = getCurrentInstance() as any;
    const formRef = ref();
    const form = ref<AsyncTask>(props.nodeData);
    // 依赖组件列表
    const getGraph = inject('getGraph') as () => Graph;
    const graph = getGraph();

    const instanceName = props.nodeData.getInstanceName();
    const nodeNames: TreeProp[] = [];

    console.log(nodeNames)
    // 版本列表
    const plugins = new Map<string, PluginDef>();
    const versionList = ref<INodeDefVersionListVo>({ versions: [] });
    const nodeId = ref<string>('');
    const getNode = inject('getNode') as () => Node;
    nodeId.value = getNode().id;
    const versionLoading = ref<boolean>(false);
    const failureVisible = ref<boolean>(false);
    const tabFlag = ref<number>(1);
    const optionalFlag = ref<boolean>(false);
    const outputTabSelected = ref<boolean>(false);
    const input = ref<JSONSchema>();

    const convertToSchema = (input: any):JSONSchema =>{
      return Try<JSONSchema>(
        () => input,
        () => { throw new TypeError(`Error parsing JSON`)});
    }
    const changeVersion = async () => {
      try {
        versionLoading.value = true;
        failureVisible.value = false;
        const selectPlugin = plugins.get(form.value.version);
        input.value  = convertToSchema(selectPlugin?.inputSchema);
          
      } catch (err) {
        proxy.$throw(err, proxy);
      } finally {
        versionLoading.value = false;
        failureVisible.value = true;
      }
    };

    const prepareNodeParams = async () => {
      try {
        //prepare plugin paramsters
        const nodes = graph.getNodes()
        for (var i = 0; i < nodes.length; i++) {
          const proxy = new CustomX6NodeProxy(nodes[i]);
          const nodeData = proxy.getData(graph);
          const anode = nodeData as AsyncTask
          if (anode.instanceName != instanceName) {
            const pluginDef = await getPluginDef(anode.name, anode.version)
            nodeNames.push({
              name: anode.instanceName,
              type:"object",
              index:-1,
              defs: pluginDef.outputSchema.definitions || {},
              schema: convertToSchema(pluginDef.outputSchema) || {},
              isLeaf: false,
              children:[],
            });
          }
        }
      } catch (err) {
        proxy.$throw(err, proxy);
      } finally {
        versionLoading.value = false;
        failureVisible.value = true;
      }
    }

    const loadVersionOrDefault = async () => {
      if (form.value.version) {
        failureVisible.value = true;
      }
      try {
        const pluginDetail = await getPluginByName(props.nodeData.name);
        pluginDetail.pluginDefs?.forEach(a => {
          plugins.set(a.version, a);
          versionList.value.versions.push(a.version);
        });

        if (!pluginDetail.pluginDefs || pluginDetail.pluginDefs.length == 0) {
          return
        }

        if (form.value.version) {
          input.value =  convertToSchema(plugins.get(form.value.version)?.inputSchema );
        } else {
          form.value.version = versionList.value.versions[0];
          input.value = convertToSchema(pluginDetail.pluginDefs[0]?.inputSchema); 
        }


      } catch (err) {
        proxy.$throw(err, proxy);
      } finally {
        versionLoading.value = false;

      }
    }
    onMounted(async () => {
      await prepareNodeParams()
      await loadVersionOrDefault()
      // 等待异步数据请求结束才代码form创建成功（解决第一次点击警告按钮打开drawer没有表单验证）
      emit('form-created', formRef.value);
    });

    return {
      formRef,
      form,
      versionList,
      ParamTypeEnum,
      nodeId,
      versionLoading,
      failureVisible,
      changeVersion,
      tabFlag,
      optionalFlag,
      outputTabSelected,
      noParamImage,
      nodeNames,
      input,
    };

  },
});
</script>

<style scoped lang="less">
.jm-workflow-editor-async-task-panel {
  .set-padding {
    padding: 0 20px;

    ::v-deep(.cache-selector) {
      margin-bottom: 20px;
    }

    .add-select-cache-btn {
      height: 24px;
      font-weight: 400;
      font-size: 14px;
      line-height: 24px;
      color: #096dd9;
      margin-bottom: 26px;

      .add-link {
        cursor: pointer;
      }
    }
  }

  .name-item {
    margin-top: 20px;
  }

  .node-item {
    padding-top: 10px;

    &:last-child {
      margin-bottom: 20px;
    }
  }

  .jm-icon-button-help::before {
    margin: 0;
  }

  .node-name {
    padding-top: 10px;
  }

  .version-description {
    font-size: 12px;
    color: #7b8c9c;
    line-height: 20px;
    margin-top: 10px;
  }

  .separate {
    height: 6px;
    background: #fafbfc;
    margin-top: 20px;
  }

  .tab-container {
    display: flex;
    font-size: 14px;
    color: #7b8c9c;
    height: 50px;
    border-bottom: 1px solid #e6ebf2;
    margin-bottom: 10px;
    padding-left: 20px;
    align-items: flex-start;
    width: 100%;

    .input-tab,
    .output-tab,
    .optional-tab {
      line-height: 50px;
      width: 56px;
      display: flex;
      flex-direction: column;
      align-items: center;
      cursor: pointer;

      .checked-underline {
        width: 37px;
        border: 1px solid #096dd9;
        position: relative;
        top: -1px;
      }
    }

    .input-tab,
    .output-tab,
    .optional-tab {
      margin-right: 40px;
    }

    .selected-tab {
      color: #096dd9;
    }
  }

  .inputs-container,
  .outputs-container,
  .optional-container {
    font-size: 14px;

    .required-icon {
      display: inline-block;
      width: 6px;
      height: 6px;
      background: url('../../svgs/required-icon.svg');
      position: relative;
      top: -5px;
    }

    .label {
      color: #3f536e;
      margin-bottom: 10px;
      padding-top: 10px;
    }

    .content {
      color: #082340;
      background: #f6f8fb;
      border-radius: 2px;
      padding: 8px 17px 8px 14px;
      margin-bottom: 10px;
    }

    .el-empty {
      padding-top: 50px;
    }
  }
}
</style>
