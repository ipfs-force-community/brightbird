<template>
  <div class="jm-workflow-editor-async-task-panel">
    <jm-form :model="form" label-position="top" ref="formRef" @submit.prevent>
      <div class="set-padding">
        <jm-form-item label="节点名称" prop="name" class="name-item" :rules="nodeData.getFormRules().name">
          <jm-input v-model="form.name" show-word-limit :maxlength="36" />
        </jm-form-item>
        <jm-form-item label="节点版本" prop="version" :rules="nodeData.getFormRules().version" class="node-item">
          <jm-select
            v-loading="versionLoading"
            :disabled="versionLoading"
            v-model="form.version"
            placeholder="请选择节点版本"
            @change="changeVersion"
          >
            <jm-option v-for="item in versionList.versions" :key="item" :label="item" :value="item" />
          </jm-select>
          <div v-if="form.versionDescription ? !versionLoading : false" class="version-description">
            {{ form.versionDescription }}
          </div>
        </jm-form-item>
      </div>
      <div class="separate"></div>
      <div v-if="form.version">
      </div>
    </jm-form>
  </div>
</template>

<script lang="ts">
import { defineComponent, getCurrentInstance, inject, onMounted, PropType, ref } from 'vue';
import { AsyncTask } from '../../model/data/node/async-task';
import { NodeGroupEnum, ParamTypeEnum } from '../../model/data/enumeration';
import { INodeDefVersionListVo } from '@/api/dto/node-definitions';
import ExpressionEditor from './form/expression-editor.vue';
import CacheSelector from './form/cache-selector.vue';
// eslint-disable-next-line no-redeclare
import { Node } from '@antv/x6';
import noParamImage from '../../svgs/no-param.svg';
import { pushParams } from '../../model/workflow-node';
import { v4 as uuidv4 } from 'uuid';

export default defineComponent({
  components: { ExpressionEditor, CacheSelector },
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

    return {};
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

    .input-tab,
    .output-tab {
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

    .input-tab {
      margin-right: 40px;
    }

    .selected-tab {
      color: #096dd9;
    }
  }

  .cache-item {
    .cache-label {
      line-height: 20px;
      margin-bottom: 16px;
      padding-top: 10px;
      color: #3f536e;
      font-size: 14px;
    }
  }

  .outputs-container {
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
