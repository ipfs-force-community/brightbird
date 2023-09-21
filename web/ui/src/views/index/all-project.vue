<template>
  <!-- 所有项目 -->
  <div class="all-project" v-loading="allProjectLoading">
    <div class="project-operator">
      <div class="project-list">
        <div class="text">项目列表</div>
      </div>
      <div class="search">
        <jm-input
          placeholder="请输入测试流名称"
          v-model="projectName"
          clearable
        />
      </div>
    </div>
    <div class="divider-line"></div>
    <div class="project">
      <template v-if="initialized && groupListRefresh">
        <template v-if="testflowGroups.length > 0">
          <project-group
            v-for="testflowGroup in testflowGroups"
            :key="testflowGroup.id"
            :sortType="sortType"
            :testflow-group="testflowGroup"
            :pageable="false"
            :project-name="projectName"
          />
        </template>
        <div class="project-empty" v-else>
          <jm-empty description="暂无项目" :image-size="98" />
        </div>
      </template>
    </div>
    <ElDialog
      v-model="copyActive"
      class="copy-dialog"
      width="500"
      title="复制测试流"
      :on-close="onCopyClose"
      :center="true"
      v-loading="copyLoading"
    >
      <ElForm :model="copyForm" :rules="rules" ref="ruleFormRef">
        <ElFormItem prop="group">
          <el-select
            v-model="copyForm.group"
            placeholder="选择你要复制到的组"
            size="large"
          >
            <el-option
              v-for="item in testflowGroups"
              :key="JSON.stringify({ name: item.name, id: item.id })"
              :label="item.name"
              :value="JSON.stringify({ name: item.name, id: item.id })"
            />
          </el-select>
        </ElFormItem>

        <ElFormItem prop="name">
          <ElInput
            size="large"
            v-model="copyForm.name"
            placeholder="请输入测试流名称"
          ></ElInput>
        </ElFormItem>
        <ElFormItem>
          <el-button @click="onCopyConfirm" type="primary">确认</el-button>
        </ElFormItem>
      </ElForm>
    </ElDialog>
  </div>
</template>

<script lang="ts">
import { ITestflowGroupVo } from '@/api/dto/testflow-group';
import { listTestflowGroup } from '@/api/view-no-auth';
import ProjectGroup from '@/views/common/project-group.vue';
import { cloneDeep, get } from 'lodash';
import {
  computed,
  defineComponent,
  getCurrentInstance,
  inject,
  nextTick,
  onBeforeMount,
  onMounted,
  reactive,
  ref,
} from 'vue';
import { onBeforeRouteLeave, useRouter } from 'vue-router';
import { namespace } from '@/store/modules/project';
import { createNamespacedHelpers, useStore } from 'vuex';
import { SortTypeEnum } from '@/api/dto/enumeration';
import { eventBus } from '@/main';
import {
  ElDialog,
  ElButton,
  ElInput,
  ElForm,
  ElFormItem,
  FormInstance,
} from 'element-plus';
import { testFlowCopy } from '@/api/job';
const { mapMutations } = createNamespacedHelpers(namespace);
export default defineComponent({
  components: { ProjectGroup, ElDialog, ElButton, ElInput, ElForm, ElFormItem },
  setup() {
    const { proxy } = getCurrentInstance() as any;
    const router = useRouter();
    const store = useStore();
    const testflowGroups = ref<ITestflowGroupVo[]>([]);
    // 已初始化
    const initialized = ref<boolean>(false);
    // 项目名称
    const projectName = ref<string>('');
    // 首页loading
    const allProjectLoading = ref<boolean>(false);

    // 改变项目组排序后强制数据及时刷新
    const groupListRefresh = ref<boolean>(true);

    const copyActive = ref<boolean>(false);
    const copyId = ref<string>('');
    const copyValue = ref<string>('');
    const copyLoading = ref<boolean>(false);
    const ruleFormRef = ref<FormInstance>();
    const copyForm = reactive({
      name: '',
      group: '',
    });

    const rules = reactive({
      name: [{ required: true, message: '请输入测试流名称' }],
      group: [{ required: true, message: '选择你要复制到的组' }],
    });

    // 项目组排序类型
    const sortTypeList = ref<Array<{ label: string; value: SortTypeEnum }>>([
      { label: '默认排序', value: SortTypeEnum.DEFAULT_SORT },
      { label: '最近触发', value: SortTypeEnum.LAST_EXECUTION_TIME },
      { label: '最近修改', value: SortTypeEnum.LAST_MODIFIED_TIME },
    ]);
    // 所有项目组在vuex中保存的排序类型
    const sortType = computed<SortTypeEnum>(
      () => store.state[namespace].sortType,
    );
    // 改变项目排序规则
    const sortChange = async (e: number) => {
      // 更改vuex中的项目组排序状态
      proxy.changeSortType(e);
      // 刷新项目组页面
      groupListRefresh.value = false;
      await nextTick();
      groupListRefresh.value = true;
    };

    const load = async () => {
      try {
        allProjectLoading.value = true;
        const testflowGroupList = await listTestflowGroup();
        initialized.value = true;
        testflowGroups.value = cloneDeep(
          testflowGroupList.filter(item => item.isShow),
        );
      } catch (err) {
        proxy.$throw(err, proxy);
      } finally {
        await nextTick(() => {
          allProjectLoading.value = false;
        });
      }
    };

    const setScrollbarOffset = inject('setScrollbarOffset') as () => void;
    const updateScrollbarOffset = inject('updateScrollbarOffset') as () => void;
    onMounted(() => {
      if (setScrollbarOffset) {
        setScrollbarOffset();
      }

      nextTick(() => {
        eventBus.on('newGroup', () => {
          load();
        });
      });
    });

    onBeforeMount(() => {
      load();
    });
    onBeforeRouteLeave((to, from, next) => {
      if (updateScrollbarOffset) {
        updateScrollbarOffset();
      }
      next();
    });

    eventBus.on('test-flow-copy', (id: any) => {
      copyActive.value = true;
      copyId.value = id;
    });

    const onCopyClose = () => {
      copyLoading.value = false;
      copyActive.value = false;
    };

    const onCopyConfirm = async () => {
      if (!ruleFormRef.value) {
        return;
      }

      ruleFormRef.value.validate(async (valid, fields) => {
        if (valid) {
          try {
            copyLoading.value = true;
            await  testFlowCopy({
              groupId:JSON.parse(copyForm.group).id,
              name:copyForm.name,
              id:copyId.value,
            });
            copyActive.value = false;
            initialized.value = false;
            load();
          } catch (error:any) {
            proxy.$error(get(error, 'response.data.message'));
          }finally {
            copyLoading.value = false;
          }
        } 
      });

      return;
  
    };

    return {
      onCopyConfirm,
      onCopyClose,
      proxy,
      testflowGroups,
      projectName,
      initialized,
      allProjectLoading,
      sortChange,
      ...mapMutations({ changeSortType: 'mutate' }),
      sortType,
      sortTypeList,
      groupListRefresh,
      copyActive,
      copyId,
      copyValue,
      copyLoading,
      ruleFormRef,
      copyForm,
      rules,
    };
  },
});
</script>

<style scoped lang="less">
// 所有项目
.all-project {
  margin-bottom: 20px;
  min-height: calc(100vh - 267px);
  background: #fff;

  .project-operator {
    display: flex;
    overflow: hidden;
    align-items: center;
    justify-content: space-between;
    margin-top: 10px;
    padding: 0 20px;
    ::v-deep(.el-input) {
      border-radius: 4px;

      .el-input__inner {
        height: 36px;
        color: #6b7b8d;
        line-height: 36px;
      }

      .el-input__inner:focus {
        border: 1px solid #096dd9;
      }
    }

    .project-list {
      display: flex;
      align-items: center;
      margin-top: 10px;
      color: #6b7b8d;
      font-size: 20px;

      .text {
        margin-right: 30px;
      }

      ::v-deep(.el-input) {
        width: 106px;
      }
    }

    .search {
      position: relative;
      display: flex;
      align-items: center;
      box-sizing: border-box;
      margin-top: 15px;

      ::v-deep(.el-input) {
        width: 488px;

        .el-input__inner {
          text-indent: 1.5em;

          &::placeholder {
            text-indent: 1.5em;
          }
        }
      }

      .jm-icon-button-search::before {
        position: absolute;
        top: 12px;
        left: 10px;
        z-index: 100;
        color: #7f8c9b;
        content: "\e80b";
      }
    }
  }

  .divider-line {
    margin: 20px auto 0;
    width: calc(100% - 40px);
    height: 1px;
    background-color: #e6ebf2;
  }
  .project {
    padding: 0 20px 30px;

    .project-empty {
      .el-empty {
        padding-top: 120px;
      }
    }
  }
}

::v-deep(.copy-dialog) {
  display: flex;
  flex-direction: column;

  .el-dialog__body {
    display: flex;
    flex-direction: column;
    .el-select {
      width: 100%;
    }
  }
}
</style>
