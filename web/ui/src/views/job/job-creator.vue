<template>
  <jm-dialog v-model="dialogVisible" width="776px" :destroy-on-close="true">
    <template #title>
      <div class="creator-title">新建Job</div>
    </template>
    <jm-form :model="createForm" :rules="editorRule" ref="createFormRef" @submit.prevent>

      <jm-form-item label="Job名称" label-position="top" prop="name">
        <jm-input v-model="createForm.name" clearable placeholder="请输入Job名称" />
      </jm-form-item>

      <jm-form-item label="测试流" prop="testFlowId">
        <jm-select 测试流组 v-loading="groupLoading" :disabled="groupLoading" v-model="selectGroupId" @change="changeGroup"
          placeholder="请选择测试流组">
          <jm-option v-for="item in groups" :key="item.id" :label="item.name" :value="item.id" />
        </jm-select>

        <jm-select 测试流 v-loading="testflowsLoading" :disabled="testflowsLoading" v-model="createForm.testFlowId" @change="onSelectTf"
          placeholder="请选择测试流">
          <jm-option v-for="item in  testflows?.list" :key="item.id" :label="item.name" :value="item.id" />
        </jm-select>
      </jm-form-item>

      <jm-form-item label="类型" prop="jobType">
        <jm-select v-loading="jobTypesLoading" :disabled="jobTypesLoading" v-model="createForm.jobType"
          placeholder="请选择Job类型">
          <jm-option v-for="item in jobTypesRef" :key="item" :label="item" :value="item" />
        </jm-select>
      </jm-form-item>

      <jm-form-item label="cron表达式"  v-show="createForm.jobType === JobEnum.CronJob" prop="cronExpression">
        <jm-input v-model="createForm.cronExpression" clearable placeholder="请输入Cron表达式" />
      </jm-form-item>


      <jm-form-item label="版本设置" prop="version">
        <div v-for="[component, version] in createForm.versions" >
          <jm-input :content=version :placeholder="`填写组件${component}的版本`" >
            <template #prepend>{{component}}:</template>
          </jm-input>
        </div>  
      </jm-form-item>

      <jm-form-item label="描述" prop="description">
        <jm-input type="textarea" v-model="createForm.description" clearable maxlength="256" show-word-limit
          placeholder="请输入描述" :autosize="{ minRows: 6, maxRows: 10 }" />
        <div class="tips">描述信息不超过 256个字符</div>
      </jm-form-item>
    </jm-form>

    <template #footer>
      <span>
        <jm-button size="small" @click="dialogVisible = false">取消</jm-button>
        <jm-button size="small" type="primary" @click="create" :loading="loading">确定</jm-button>
      </span>
    </template>
  </jm-dialog>
</template>

<script lang="ts">
import { defineComponent, getCurrentInstance, ref, SetupContext } from 'vue';
import { createJob, getJobTypes } from '@/api/job';
import { listProjectGroup, queryTestFlow } from '@/api/view-no-auth';

import { IJobCreateVo } from '@/api/dto/job';
import { Mutable } from '@/utils/lib';
import { JobEnum } from '@/api/dto/enumeration';
import { IProjectGroupVo } from '@/api/dto/project-group';
import { ITestFlowDetail } from '@/api/dto/project';
import { IPageVo } from '@/api/dto/common';

export default defineComponent({
  emits: ['completed'],
  setup(_, { emit }: SetupContext) {
    const { proxy } = getCurrentInstance() as any;
    const dialogVisible = ref<boolean>(true);
    const createFormRef = ref<any>(null);
    const jobTypesLoading = ref<boolean>(false);


    const groupLoading = ref<boolean>(false);
    const testflowsLoading = ref<boolean>(false);
    const jobTypesRef = ref<JobEnum[]>([]);
    const selectGroupId = ref<string>();
    const groups = ref<IProjectGroupVo[]>([]);
    const testflows = ref<IPageVo<ITestFlowDetail>>({
      total: -1,
      pages: 0,
      list: [],
      pageNum: 0,
    });

    const createForm = ref<Mutable<IJobCreateVo>>({
      name: '',
      testFlowId: '',
      jobType: JobEnum.CronJob,
      description: "",
      versions:new Map<string, string>(),
      cronExpression: "",
    });

    const editorRule = ref<object>({
      name: [{ required: true, message: 'job名称不能为空', trigger: 'blur' }],
      testFlowId: [{ required: true, message: '需要选择测试流', trigger: 'blur' }],
      jobType: [{ required: true, message: '选择job类型', trigger: 'blur' }],
    });
    const loading = ref<boolean>(false);
    const create = async () => {
      createFormRef.value.validate(async (valid: boolean) => {
        if (!valid) {
          return;
        }

        loading.value = true;
        try {
          await createJob(createForm.value);
          proxy.$success('Job创建成功');
          emit('completed');
          dialogVisible.value = false;
        } catch (err) {
          proxy.$throw(err, proxy);
        } finally {
          loading.value = false;
        }
      });
    };

    const initJobTypes = async () => {
      jobTypesLoading.value = true;
      try {
        jobTypesRef.value = await getJobTypes()
      } catch (err) {
        proxy.$throw(err, proxy);
      } finally {
        jobTypesLoading.value = false;
      }
    }
    initJobTypes()


    const fetchGroupList = async () => {
      groupLoading.value = true;
      try {
        groups.value = await listProjectGroup()
      } catch (err) {
        proxy.$throw(err, proxy);
      } finally {
        groupLoading.value = false;
      }
    }
    fetchGroupList()


    const changeGroup = async () => {
      testflowsLoading.value = true;
      createForm.value.testFlowId = ""
      try {
        testflows.value = await queryTestFlow({
          groupId: selectGroupId.value,
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
      const versions = new Map<string, string>();
        testflows.value?.list?.forEach(f=>{
          versions.set(f.name, "");
        })
        createForm.value.versions = versions;
    }

    return {
      JobEnum,
      dialogVisible,
      createFormRef,
      createForm,
      editorRule,
      loading,
      jobTypesRef,
      jobTypesLoading,
      groupLoading,
      selectGroupId,
      groups,
      changeGroup,
      onSelectTf,
      testflowsLoading,
      testflows,
      create,
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

.creator-title {
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
