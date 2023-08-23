<template>
  <jm-dialog v-model="dialogVisible" width="776px" :destroy-on-close="true">
    <template #title>
      <div class="creator-title">新建Job</div>
    </template>
    <jm-form label-width="auto" :model="createForm" :rules="editorRule" ref="createFormRef" @submit.prevent>

      <jm-form-item label="Job名称" label-position="top" prop="name">
        <jm-input v-model="createForm.name" @blur="checkJobName" v-bind:class="isDupName ? 'invadateName' : ''" clearable
          placeholder="请输入Job名称" />
      </jm-form-item>

      <jm-form-item label="测试流" prop="testFlowId">
        <jm-select 测试流组 v-loading="groupLoading" :disabled="groupLoading" v-model="selectGroupId" @change="changeGroup"
          placeholder="请选择测试流组">
          <jm-option v-for="item in groups" :key="item.id" :label="item.name" :value="item.id" />
        </jm-select>

        <jm-select 测试流 v-loading="testflowsLoading" :disabled="testflowsLoading" v-model="createForm.testFlowId"
          @change="onSelectTf" placeholder="请选择测试流">
          <jm-option v-for="item in  testflows" :key="item.id" :label="item.name" :value="item.id" />
        </jm-select>
      </jm-form-item>


      <jm-form-item label="描述" prop="description">
        <jm-input type="textarea" v-model="createForm.description" clearable maxlength="256" show-word-limit
          placeholder="请输入描述" :autosize="{ minRows: 6, maxRows: 10 }" />
        <div class="tips">描述信息不超过 256个字符</div>
      </jm-form-item>


      <jm-form-item label="类型" prop="jobType">
        <jm-select v-loading="jobTypesLoading" :disabled="jobTypesLoading" v-model="createForm.jobType"
          @change="onSelectJobtype" placeholder="请选择Job类型">
          <jm-option v-for="item in jobTypesRef" :key="item" :label="item" :value="item" />
        </jm-select>
      </jm-form-item>


      <jm-form-item label="cron表达式" v-show="createForm.jobType === JobEnum.CronJob" prop="cronExpression">
        <jm-input v-model="createForm.cronExpression" clearable placeholder="请输入Cron表达式" />
      </jm-form-item>
      <div>
        <div v-show="createForm.jobType == JobEnum.CronJob">
          <el-text>版本设置:</el-text>
          <div class="form-inter version" v-for="(version, component) in createForm.versions" :key="component">
            <el-row>
              <el-col :span="6">
                {{ component }}
              </el-col>
              <el-col :span="14">
                <jm-input :content=version :placeholder="`填写组件${component}的版本`"
                  v-model="createForm.versions[component]" />
              </el-col>
            </el-row>
          </div>
        </div>

        <div v-show="createForm.jobType == JobEnum.TagCreated">
          <el-text>tag匹配:</el-text>
          <div class="form-inter version" v-for="match,index in createForm.tagCreateEventMatches" :key="index">
            <el-row>
              <el-col :span="6">
                {{ getRepoName(match.repo) }}
              </el-col>
              <el-col :span="14">
                <jm-input :content=match.tagPattern :placeholder="`匹配模式 例 tag/.`" v-model="match.tagPattern" />
              </el-col>
            </el-row>
          </div>
        </div>

        <div v-show="createForm.jobType == JobEnum.PRMerged">
          <el-text>分支匹配:</el-text>
          <div class="form-inter version" v-for="match,index in createForm.prMergedEventMatches" :key="index">
            <el-row>
              <el-col :span="7">
                {{ getRepoName(match.repo) }}
              </el-col>
              <el-col :span="7">
                <jm-input :content=match.basePattern v-model="match.basePattern" />
              </el-col>
              <el-col :span="7">
                <jm-input :content=match.sourcePattern v-model="match.sourcePattern" />
              </el-col>
            </el-row>
          </div>
        </div>
      </div>
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
import { countJob, createJob, getJobTypes } from '@/api/job';
import { fetchDeployPlugins } from '@/api/plugin';
import { fetchTestFlowDetail, listTestflowGroup, queryTestFlow } from '@/api/view-no-auth';
import { START_PAGE_NUM } from '@/utils/constants';
import { IJobCreateVo } from '@/api/dto/job';
import { Mutable } from '@/utils/lib';
import { JobEnum, PluginTypeEnum } from '@/api/dto/enumeration';
import { ITestflowGroupVo } from '@/api/dto/testflow-group';
import { ITestFlowDetail, Plugin } from '@/api/dto/testflow';
import { ElCol, ElRow, FormRules } from 'element-plus';
import yaml from 'yaml';

export default defineComponent({
  emits: ['completed'],
  components: { ElRow, ElCol },
  setup(_, { emit }: SetupContext) {
    const { proxy } = getCurrentInstance() as any;
    const dialogVisible = ref<boolean>(true);
    const isDupName = ref<boolean>(true);
    const createFormRef = ref<any>(null);
    const jobTypesLoading = ref<boolean>(false);
    const groupLoading = ref<boolean>(false);
    const testflowsLoading = ref<boolean>(false);
    const jobTypesRef = ref<JobEnum[]>([]);
    const selectGroupId = ref<string>();
    const groups = ref<ITestflowGroupVo[]>([]);
    const testflows = ref<ITestFlowDetail[]>([]);
    const createForm = ref<Mutable<IJobCreateVo>>({
      name: '',
      testFlowId: '',
      jobType: JobEnum.CronJob,
      description: '',
      versions: {},
      cronExpression: '',
      prMergedEventMatches: [],
      tagCreateEventMatches: [],
    });
    const editorRule = ref<FormRules<IJobCreateVo>>({
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
        if (isDupName.value) {
          proxy.$error('Job名称不合法，空或者重复');
          return;
        }

        loading.value = true;
        try {
          await createJob(createForm.value);
          proxy.$success('Job创建成功');
          emit('completed');
          dialogVisible.value = false;
        }
        catch (err) {
          proxy.$throw(err, proxy);
        }
        finally {
          loading.value = false;
        }
      });
    };

    const initJobTypes = async () => {
      jobTypesLoading.value = true;
      try {
        jobTypesRef.value = await getJobTypes();
      }
      catch (err) {
        proxy.$throw(err, proxy);
      }
      finally {
        jobTypesLoading.value = false;
      }
    };

    initJobTypes();

    const fetchGroupList = async () => {
      groupLoading.value = true;
      try {
        groups.value = await listTestflowGroup();
      }
      catch (err) {
        proxy.$throw(err, proxy);
      }
      finally {
        groupLoading.value = false;
      }
    };
    fetchGroupList();

    const refreshSelect = async (testflow: ITestFlowDetail) => {
      createForm.value.testFlowId = testflow.id ?? '';
      let versions: any = {};
      const { pipeline } = yaml.parse(testflow.graph);
      Object.values(pipeline).forEach((f: any) => {
        if (f.pluginType === PluginTypeEnum.Exec) {
          return;
        }
        versions[f.name] = '';
      });
      // use for cron
      createForm.value.versions = versions;
    };
    const changeGroup = async () => {
      testflowsLoading.value = true;
      createForm.value.testFlowId = '';
      try {
        testflows.value = (await queryTestFlow({
          groupId: selectGroupId.value ?? '',
          pageNum: START_PAGE_NUM,
          pageSize: Number.MAX_SAFE_INTEGER,
        })).list;
        const firstflow = testflows.value[0];
        if (firstflow) {
          refreshSelect(firstflow);
        }
      }
      catch (err) {
        proxy.$throw(err, proxy);
      }
      finally {
        testflowsLoading.value = false;
      }
    };
    const onSelectJobtype = async () => {
      try {
        // fetch testflow
        const nodeInUse = new Set<string>();
        const testflow = await fetchTestFlowDetail({ id: createForm.value.testFlowId });
        const { pipeline } = yaml.parse(testflow.graph);
        Object.values(pipeline).forEach((a:any) => nodeInUse.add(a.name + a.version));
        // fetch plugins
        const pluginMap = new Map<string, Plugin>();
        (await fetchDeployPlugins()).map(a => {
          a.pluginDefs?.map(p => {
            if (nodeInUse.has(p.name + p.version)) {
              pluginMap.set(p.name, p);
            }
          });

        });

        const repos = new Set<string>();
        Object.entries(createForm.value.versions).map(([k, v]) => {
          const repoName = pluginMap.get(k)?.repo ?? '';
          if (!repos.has(repoName)) {
            repos.add(repoName);
          }
        });

        if (createForm.value.jobType === JobEnum.TagCreated) {
          createForm.value.tagCreateEventMatches = [];
          [...repos].map(repoName => createForm.value.tagCreateEventMatches.push({
            repo: repoName,
            tagPattern: 'tag/.+',
          }));
        } else if (createForm.value.jobType === JobEnum.PRMerged) {
          createForm.value.prMergedEventMatches = [];
          [...repos].map(repoName => createForm.value.prMergedEventMatches.push({
            repo: repoName,
            sourcePattern: 'feat\/.+|fix\/.+',
            basePattern: 'master|main',
          }));
        }
      }
      catch (err) {
        proxy.$throw(err, proxy);
      }
    };
    const checkJobName = async () => {
      try {
        const count = await countJob({
          name: createForm.value.name,
        });
        isDupName.value = count > 0;
      } catch (err) {
        proxy.$throw(err, proxy);
      } finally {
        loading.value = false;
      }
    };
    const onSelectTf = async () => {
      const selTf = testflows.value?.find(a => a.id === createForm.value.testFlowId);
      if (selTf) {
        refreshSelect(selTf);
      }
    };
    const getRepoName = (gitURL: string): string => {
      const url = new URL(gitURL);
      return url.pathname.replace('.git', '').substring(1).split('/')[1];
    };
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
      onSelectJobtype,
      testflowsLoading,
      testflows,
      create,
      // utils
      getRepoName,
      checkJobName,
      isDupName,
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

.form-inter {
  display: inline-block;
  padding-top: 18px;
  font-size: 14px;
  width: 100%;
}

.jobtype {
  padding-top: 20px;
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

.version {
  margin-left: 24px;
}

.invadateName ::v-deep input {
  border-color: #f56c6c;
}</style>
