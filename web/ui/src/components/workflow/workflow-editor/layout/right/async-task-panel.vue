<template>
  <div class="jm-workflow-editor-async-task-panel">
    <jm-form :model="form" label-position="top" ref="formRef" @submit.prevent>
      <div class="set-padding">
        <jm-form-item label="节点名称" prop="name" class="name-item" :rules="nodeData.getFormRules().name">
          <jm-input v-model="form.name" show-word-limit :maxlength="36" />
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
          <div :class="{ 'output-tab': true, 'selected-tab': tabFlag === 2 }" @click="tabFlag = 2">
            输出参数
            <div class="checked-underline" v-if="tabFlag === 2"></div>
          </div>
          <div :class="{ 'optional-tab': true, 'selected-tab': tabFlag === 3 }" @click="tabFlag = 3">
            依赖参数
            <div class="checked-underline" v-if="tabFlag === 3"></div>
          </div>
        </div>
        <div class="inputs-container set-padding" v-if="tabFlag === 1">
          <div v-if="!form.inputs.toString()">
            <jm-empty description="无输入参数" :image="noParamImage"></jm-empty>
          </div>
            <div v-else>
            <jm-form-item v-for="(item, index) in form.inputs" :key="item.name" :prop="`inputs.${index}.value`"
              :rules="nodeData.getFormRules().version" class="node-name">
              <template #label>
                {{ item.name }} ({{ item.type }})
                <jm-tooltip placement="top" v-if="item.description" :append-to-body="false" :content="item.description">
                  <i class="jm-icon-button-help"></i>
                </jm-tooltip>
              </template>
              <jm-input v-model="item.value" :node-id="nodeId"
                :placeholder="item.description ? item.description : '请输入' + item.name" show-word-limit :maxlength="36" />
            </jm-form-item>
          </div>
        </div>
        <div class="outputs-container set-padding" v-else-if="tabFlag === 2">
          <div v-if="!form.instance.toString()">
              <jm-empty description="无输出参数" :image="noParamImage"></jm-empty>
          </div>
          <div v-else>
            <div class="label">
              <i class="required-icon" v-if="form.instance.require"></i>
              组件实例名称
              <jm-tooltip placement="top" v-if="form.instance.description" :append-to-body="false"
                :content="form.instance.description">
                <i class="jm-icon-button-help"></i>
              </jm-tooltip>
            </div>
            <jm-input v-model="form.instance.value" :node-id="nodeId"
              :placeholder="form.instance.description ? form.instance.description : '请输入' + form.instance.type"
              show-word-limit :maxlength="36" />
          </div>
        </div>
        <div class="optional-container set-padding" v-else-if="tabFlag === 3">
          <div v-if="!form.dependencies.toString()">
              <jm-empty description="无依赖参数" :image="noParamImage"></jm-empty>
          </div>
          <div v-else>
            <jm-form-item v-for="(item, index) in form.dependencies" :key="item.name"
              :prop="form.dependencies.length ? `dependencies.${index}.value` : null" class="node-name">
              <template #label>
                {{ item.name }}
                <jm-tooltip placement="top" v-if="item.description" :append-to-body="false" :content="item.description">
                  <i class="jm-icon-button-help"></i>
                </jm-tooltip>
              </template>
              <jm-select
                v-model="item.value"
                :node-id="nodeId"
                :placeholder="item.description ? item.description : '请输入' + item.name"
                show-word-limit
                :maxlength="36"
              >
                <jm-option v-for="nodeName in nodeNames" :key="nodeName" :label="nodeName" :value="nodeName" />
              </jm-select>
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
import {Graph, Node} from '@antv/x6';
import JmEmpty from "@/components/data/empty/index.vue";
import JmForm from "@/components/form/form";
import jmFormItem from "@/components/form/form-item";
import JmInput from "@/components/form/input";
import { getPluginByName } from '@/api/plugin';
import { Plugin } from '@/api/dto/testflow';
import {CustomX6NodeProxy} from "@/components/workflow/workflow-editor/model/data/custom-x6-node-proxy";
import JmSelect from "@/components/form/select";

export default defineComponent({
  components: { JmEmpty, ExpressionEditor, JmForm, jmFormItem, JmInput, JmSelect },
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

    const instanceName = props.nodeData.getDisplayName();
    const nodeNames: string[] = [];
    graph.getNodes().forEach(node=>{
        const proxy = new CustomX6NodeProxy(node);
        // 不能为ref，否则，表单内容的变化影响数据绑定
        const nodeData = proxy.getData(graph);
       const  displayName = nodeData.getDisplayName();
        if (displayName&&displayName!=instanceName) {
            nodeNames.push(displayName)
        }
    })

    // 版本列表
    const plugins = new Map<string, Plugin>();
    const versionList = ref<INodeDefVersionListVo>({ versions: [] });
    const nodeId = ref<string>('');
    const getNode = inject('getNode') as () => Node;
    nodeId.value = getNode().id;
    const versionLoading = ref<boolean>(false);
    const failureVisible = ref<boolean>(false);
    const tabFlag = ref<number>(1);
    const optionalFlag = ref<boolean>(false);
    const outputTabSelected = ref<boolean>(false);

    const changeVersion = async () => {
      form.value.inputs.length = 0;
      form.value.dependencies.length = 0;
      try {
        versionLoading.value = true;
        failureVisible.value = false;
        const selectPlugin = plugins.get(form.value.version);
        form.value.inputs = selectPlugin?.properties ?? [];
        form.value.dependencies = selectPlugin?.dependencies ?? [];
      } catch (err) {
        proxy.$throw(err, proxy);
      } finally {
        versionLoading.value = false;
        failureVisible.value = true;
      }
    };

    const generateRandomNumber = () => {
      const randomNumber = Math.floor(1000 + Math.random() * 9000);
      return randomNumber.toString();
    };

    onMounted(async () => {
      if (form.value.version) {
        failureVisible.value = true;
      }
      try {
        const pluginDetail = await getPluginByName(props.nodeData.name);
        pluginDetail.plugins.forEach(a => {
          plugins.set(a.version, a);
          versionList.value.versions.push(a.version)
        });
        if (!form.value.version) {
          form.value.version = versionList.value.versions[0];
          form.value.inputs = pluginDetail.plugins[0].properties ?? [];
          form.value.dependencies = pluginDetail.plugins[0].dependencies ?? [];
        }

        if (!form.value.instance || !form.value.instance.value) {
          const defaultInstanceName = `${form.value.name}-${generateRandomNumber()}`;
          form.value.instance = {
            name: '组件实例名称',
            value: defaultInstanceName,
            type: form.value.instance.type,
            sockPath: '',
            require: true,
            description: '',
          };
        }

      } catch (err) {
        proxy.$throw(err, proxy);
      } finally {
        versionLoading.value = false;
        // 等待异步数据请求结束才代码form创建成功（解决第一次点击警告按钮打开drawer没有表单验证）
        emit('form-created', formRef.value);
      }
    });

    return {
      formRef,
      form,
      versionList,
      ParamTypeEnum,
      nodeId,
      versionLoading,
      failureVisible,
      // 获取节点信息
      changeVersion,
      tabFlag,
      optionalFlag,
      outputTabSelected,
      noParamImage,
      nodeNames,
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
