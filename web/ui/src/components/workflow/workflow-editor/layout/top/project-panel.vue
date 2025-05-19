<template>
  <jm-dialog width="667px" destroy-on-close @close="cancel">
    <template #title>
      <div class="title-container"><i class="jm-icon-workflow-edit"></i>编辑测试流信息</div>
    </template>
    <div class="jm-workflow-editor-project-panel">
      <jm-form label-width="auto" ref="editProjectInfoRef" :model="projectInfoForm" :rules="rules" @submit.prevent>
        <jm-form-item label="名称" prop="name" >
          <jm-input
            v-model="projectInfoForm.name"
            :maxlength="45"
            placeholder="请输入测试流名称"
            :show-word-limit="true"
            :class="{ 'has-error': nameError }"
          />
          <div v-if="nameError" class="error-msg">该测试流名称已存在，请修改</div>
        </jm-form-item>

        <jm-form-item label="分组" class="group-item" prop="groupId">
          <jm-select v-model="projectInfoForm.groupId" placeholder="请选择测试流分组" v-loading="loading">
            <jm-option v-for="item in testGroupList" :key="item.id" :label="item.name" :value="item.id" />
          </jm-select>
        </jm-form-item>
        <jm-form-item label="描述" class="description-item">
          <jm-input
            type="textarea"
            v-model="projectInfoForm.description"
            :maxlength="255"
            placeholder="请输入测试流描述"
            :show-word-limit="true"
          />
        </jm-form-item>
        <div class="btn-container">
          <jm-button @click="cancel" class="cancel">取消</jm-button>
          <jm-button type="primary" @click="save()">确定</jm-button>
        </div>
      </jm-form>
    </div>
  </jm-dialog>
</template>

<script lang="ts">
import { defineComponent, onMounted, getCurrentInstance, PropType, ref } from 'vue';
import { IWorkflow } from '../../model/data/common';
import { ITestflowGroupVo } from '@/api/dto/testflow-group';
import { listTestflowGroup, countTestFlow } from '@/api/view-no-auth';

export interface IProjectInfo {
  name: string;
  description: string;
  groupId: string;
}

export default defineComponent({
  props: {
    workflowData: {
      type: Object as PropType<IWorkflow>,
      required: true,
    },
  },
  emits: ['save', 'update:model-value'],
  setup(props, { emit }) {
    const { proxy } = getCurrentInstance() as any;
    const loading = ref<boolean>(true);
    const workflowForm = ref<IWorkflow>(props.workflowData);
    const projectInfoForm = ref<IProjectInfo>({
      name: props.workflowData.name,
      description: props.workflowData.description || '',
      groupId: props.workflowData.groupId,
    });

    const editProjectInfoRef = ref<HTMLFormElement>();
    // 分组列表
    const testGroupList = ref<ITestflowGroupVo[]>([]);

    const nameError = ref<boolean>(false);

    onMounted(async () => {
      testGroupList.value = await listTestflowGroup();
      loading.value = false;
    });

    return {
      editProjectInfoRef,
      testGroupList,
      projectInfoForm,
      loading,
      nameError,
      save: async () => {
        editProjectInfoRef.value?.validate(async (valid: boolean) => {
          if (!valid) {
            return false;
          }
          if (projectInfoForm.value.name && projectInfoForm.value.groupId) {
            const count = await countTestFlow({
              name: projectInfoForm.value.name,
            });
            if (count !== 0) {
              nameError.value = true;
              return;
            }
          }
          // 测试流组中测试流为空，将其自动折叠
          workflowForm.value.name = projectInfoForm.value.name;
          workflowForm.value.groupId = projectInfoForm.value.groupId;
          emit('update:model-value', false);
          emit('save');
        });
      },
      cancel: () => {
        emit('update:model-value', false);
      },
      rules: {
        name: [{ required: true, message: '请输入项目名称', trigger: 'blur' }],
        groupId: [{ required: true, message: '请选择项目分组', trigger: 'change' }],
      },
    };
  },
});
</script>

<style scoped lang="less">
.el-dialog {
  .title-container {
    font-size: 16px;

    .jm-icon-workflow-edit {
      margin-right: 10px;
    }
  }

  .jm-workflow-editor-project-panel {
    .group-item,
    .description-item {
      padding-top: 10px;
    }

    .description-item {
      padding-bottom: 30px;
    }

    .btn-container {
      display: flex;
      justify-content: flex-end;

      .cancel {
        color: #082340;
        background: #f5f5f5;
        border-radius: 2px;
        border: none;
        box-shadow: none;

        &:hover {
          background: #d9d9d9;
        }
      }
    }
  }
  .error-msg {
    font-size: 12px;
    color: #f56c6c;
    margin-top: 5px;
  }
  .has-error ::v-deep input{
    border-color: #f56c6c;
  }
}
</style>
