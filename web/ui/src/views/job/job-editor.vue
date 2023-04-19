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

        <jm-select 测试流 v-loading="testflowsLoading" :disabled="testflowsLoading" v-model="editorForm.testFlowId"
          @change="onSelectTf" placeholder="请选择测试流">
          <jm-option v-for="item in  testflows" :key="item.id" :label="item.name" :value="item.id" />
        </jm-select>
      </jm-form-item>

      <jm-form-item v-show="jobType == JobEnum.CronJob" label="版本设置" prop="version">
        <div v-for="(version, component) in editorForm.versions">
          <jm-input :content=version :placeholder="`填写组件${component}的版本(branch/tag/commit/不填默认主分支)`"
            v-model="editorForm.versions[component]">
            <template #prepend>{{ component }}:</template>
          </jm-input>
        </div>
      </jm-form-item>



      <jm-form-item v-show="jobType == JobEnum.TagCreated" label="" prop="tagCreateEventMatchs">
        <div v-for="match in editorForm.tagCreateEventMatchs">
          <jm-input :content=match.tagPattern :placeholder="`填写代码库${match.repo}的匹配模式 tag/*`"
            v-model="match.tagPattern">
            <template #prepend>{{ match.repo }}:</template>
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

import { fetchTestFlowDetail, listTestflowGroup, queryTestFlow } from '@/api/view-no-auth';
import { ITestflowGroupVo } from '@/api/dto/testflow-group';
import { IJobUpdateVo, IPRMergedEventMatch } from '@/api/dto/job';
import { getJob, updateJob } from '@/api/job'
import { ITestFlowDetail } from '@/api/dto/testflow.js';
import { Mutable } from '@/utils/lib';
import { START_PAGE_NUM } from '@/utils/constants';
import { JobEnum } from '@/api/dto/enumeration';

export default defineComponent({
  emits: ['completed'],
  props: {
    id: { type: String, required: true },
  },
  setup(props, { emit }: SetupContext) {
    const { proxy } = getCurrentInstance() as any;
    const dialogVisible = ref<boolean>(true);
    const editorFormRef = ref<any>(null);
    const jobType = ref<JobEnum>(JobEnum.CronJob);
    const editorForm = ref<Mutable<IJobUpdateVo>>({
      testFlowId: "",
      name: "",
      description: "",
      versions: { "a": "b" },

      //cron
      cronExpression: "",
      prMergedEventMatchs: [],
      tagCreateEventMatchs: [],
    });
    const editorRule = ref<object>({
      name: [{ required: true, message: '分组名称不能为空', trigger: 'blur' }],
    });

    const loading = ref<boolean>(false);
    const fetchJob = async () => {
      loading.value = true;
      try {
        const job = await getJob(props.id);
        editorForm.value = job;
        jobType.value = job.jobType;
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
    const groups = ref<ITestflowGroupVo[]>([]);
    const testflows = ref<ITestFlowDetail[]>([]);

    const fetchGroupList = async () => {
      groupLoading.value = true;
      try {
        groups.value = await listTestflowGroup();
        const testflow = await fetchTestFlowDetail({ id: editorForm.value.testFlowId, name: "" });
        selectGroupId.value = testflow.groupId;
      } catch (err) {
        proxy.$throw(err, proxy);
      } finally {
        groupLoading.value = false;
      }
    }

    const refreshSelect = (testflow: ITestFlowDetail) => {
      editorForm.value.testFlowId = testflow.id ?? "";
      let versions: any = {};
      testflow?.nodes?.forEach(f => {
        versions[f.name] = "";
      })
      editorForm.value.versions = versions;
    }

    const changeGroup = async () => {
      testflowsLoading.value = true;
      editorForm.value.testFlowId = "";
      editorForm.value.versions = {};
      try {
        testflows.value = (await queryTestFlow({
          groupId: selectGroupId.value ?? "",
          pageNum: START_PAGE_NUM,
          pageSize: Number.MAX_SAFE_INTEGER,
        })).list;
        const firstflow = testflows.value[0]
        if (firstflow) {
          refreshSelect(firstflow);
        }
      } catch (err) {
        proxy.$throw(err, proxy);
      } finally {
        testflowsLoading.value = false;
      }
    }

    const onSelectTf = async () => {
      const selTf = testflows.value?.find(a => a.id == editorForm.value.testFlowId)
      if (selTf) {
        refreshSelect(selTf);
      }
    }

    const save = async () => {
      editorFormRef.value.validate(async (valid: boolean) => {
        if (!valid) {
          return;
        }
        const { name, description, testFlowId, versions, cronExpression, prMergedEventMatchs, tagCreateEventMatchs } = editorForm.value;
        try {
          loading.value = true;
          await updateJob(props.id, {
            name: name,
            testFlowId: testFlowId,
            description: description,
            versions: versions,
            cronExpression: cronExpression,
            prMergedEventMatchs: prMergedEventMatchs,
            tagCreateEventMatchs: tagCreateEventMatchs,
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
      testflows.value = (await queryTestFlow({
        groupId: selectGroupId.value ?? "",
        pageNum: START_PAGE_NUM,
        pageSize: Number.MAX_SAFE_INTEGER,
      })).list
    });
    return {
      dialogVisible,
      editorFormRef,
      editorForm,
      editorRule,
      loading,
      jobType,

      selectGroupId,
      groupLoading,
      testflowsLoading,
      groups,
      testflows,
      changeGroup,
      onSelectTf,
      save,
      JobEnum,
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
