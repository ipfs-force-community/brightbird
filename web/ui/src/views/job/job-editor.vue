<template>
  <jm-dialog v-model="dialogVisible" width="776px" :destroy-on-close="true">
    <template #title>
      <div class="editor-title">编辑项目分组</div>
    </template>
    <jm-form :model="editorForm" :rules="editorRule" ref="editorFormRef" @submit.prevent>
      <jm-form-item label="Job名称" label-position="top" prop="name">
        <jm-input v-model="editorForm.name" clearable placeholder="请输入Job名称" />
      </jm-form-item>

      <jm-form-item label="测试流" prop="testFlowId">
        <jm-select 测试流组 v-loading="groupLoading" :disabled="groupLoading" v-model="selectGroupId" @change="changeGroup"
          placeholder="请选择测试流组">
          <jm-option v-for="item in groups" :key="item.id" :label="item.name" :value="item.id" />
        </jm-select>

        <jm-select 测试流 v-loading="testflowsLoading" :disabled="testflowsLoading" v-model="editorForm.testFlowId" @change="onSelectTf"
          placeholder="请选择测试流">
          <jm-option v-for="item in  testflows?.list" :key="item.id" :label="item.name" :value="item.id" />
        </jm-select>
      </jm-form-item>

      <jm-form-item label="版本设置" prop="version">
        <div v-for="(version, component) in editorForm.versions">
          <jm-input :content=version :placeholder="`填写组件${component}的版本`" v-model="editorForm.versions[component]">
            <template #prepend>{{component}}:</template>
          </jm-input>
        </div>  
      </jm-form-item>

      <jm-form-item label="描述" prop="description">
        <jm-input type="textarea" v-model="editorForm.description" clearable maxlength="256" show-word-limit
          placeholder="请输入描述" :autosize="{ minRows: 6, maxRows: 10 }" />
        <div class="tips">描述信息不超过 256个字符</div>
      </jm-form-item>

    </jm-form>
    <template #footer>
      <span>
        <jm-button size="small" @click="dialogVisible = false">取消</jm-button>
        <jm-button size="small" type="primary" @click="save" :loading="loading">保存</jm-button>
      </span>
    </template>
  </jm-dialog>
</template>

<script lang="ts">
import {
  defineComponent,
  getCurrentInstance,
  ref,
  SetupContext,
  onMounted,
} from 'vue';

import {fetchTestFlowDetail, listProjectGroup, queryTestFlow} from '@/api/view-no-auth';
import { IProjectGroupVo } from '@/api/dto/project-group';
import { IJobUpdateVo } from '@/api/dto/job';
import { getJob, updateJob } from '@/api/job'
import { ITestFlowDetail } from '@/api/dto/project';
import { IPageVo } from '@/api/dto/common';
import { Mutable } from '@/utils/lib';

export default defineComponent({
  emits: ['completed'],
  props: {
    id: { type: String, required: true },
  },
  setup(props, { emit }: SetupContext) {
    const { proxy } = getCurrentInstance() as any;
    const dialogVisible = ref<boolean>(true);
    const editorFormRef = ref<any>(null);
    const editorForm = ref<Mutable<IJobUpdateVo>>({
      testFlowId: "",
      name: "",
      description: "",
      versions:{"a":"b"},
      cronExpression: "",
    });
    const editorRule = ref<object>({
      name: [{ required: true, message: '分组名称不能为空', trigger: 'blur' }],
    });

    const loading = ref<boolean>(false);
    const fetchJob = async () => {
      loading.value = true;
      try {
        editorForm.value = await getJob(props.id);
      } catch (err) {
        proxy.$throw(err, proxy);
      } finally {
        loading.value = false;
      }
    };

    //testflow select
    const selectGroupId = ref<string>();
    const groupLoading = ref<boolean>(false);
    const testflowsLoading = ref<boolean>(false);
    const groups = ref<IProjectGroupVo[]>([]);
    const testflows = ref<IPageVo<ITestFlowDetail>>({
      total: -1,
      pages: 0,
      list: [],
      pageNum: 0,
    });
    const fetchGroupList = async () => {
      groupLoading.value = true;
      try {
        groups.value = await listProjectGroup();
        const testflow = await fetchTestFlowDetail({id:"editorForm.value.testFlowId", name:""});
        selectGroupId.value = testflow.groupId;
      } catch (err) {
        proxy.$throw(err, proxy);
      } finally {
        groupLoading.value = false;
      }
    }

    const changeGroup = async () => {
      testflowsLoading.value = true;
      editorForm.value.testFlowId = ""
      try {
        testflows.value = await queryTestFlow({
          groupId: selectGroupId.value??"",
          pageNum: 0,
          pageSize: 0,
        })
      } catch (err) {
        proxy.$throw(err, proxy);
      } finally {
        testflowsLoading.value = false;
      }
    }

    const onSelectTf = async () =>{
        let versions: any= {};
        const selTf = testflows.value?.list?.find(a=>a.id == editorForm.value.testFlowId)
        selTf?.nodes?.forEach(f=>{
          versions[f.name] = "";
        })
        editorForm.value.versions = versions;
    }

    const save = async () => {
      editorFormRef.value.validate(async (valid: boolean) => {
        if (!valid) {
          return;
        }
        const { name, description, testFlowId, versions, cronExpression } = editorForm.value;
        try {
          loading.value = true;
          await updateJob(props.id, {
            name:name,
            testFlowId:testFlowId,
            description: description,
            versions: versions,
            cronExpression:cronExpression,
          });
          proxy.$success('项目分组修改成功');
          emit('completed');
          dialogVisible.value = false;
        } catch (err) {
          proxy.$throw(err, proxy);
        } finally {
          loading.value = false;
        }
      });
    };
    onMounted(async () => {
      await fetchJob()
      await fetchGroupList()
      testflows.value = await queryTestFlow({
        groupId: selectGroupId.value??"",
        pageNum: 0,
        pageSize: 0,
      })
    });
    return {
      dialogVisible,
      editorFormRef,
      editorForm,
      editorRule,
      loading,

      selectGroupId,
      groupLoading,
      testflowsLoading,
      groups,
      testflows,
      changeGroup,
      onSelectTf,
      save,
    };
  },
});
</script>

<style scoped lang="less">
.el-form-item {
  &.is-show {
    margin-bottom: 0px;
    margin-top: -10px;
  }
}

.editor-title {
  padding-left: 36px;
  background-image: url('@/assets/svgs/btn/edit.svg');
  background-repeat: no-repeat;
  background-position: left center;
}

.tips {
  color: #6b7b8d;
  margin-left: 15px;
}
</style>
