<template>
  <jm-dialog v-model="dialogVisible" width="1200px" :destroy-on-close="true">
    <template #title>
      <div class="creator-title">
        <div class="edit-icon">
          <img src="~@/assets/svgs/btn/edit.svg" alt="" />
        </div>
        <span>添加测试流</span>
      </div>
    </template>
    <jm-form label-width="auto" :model="createForm" :rules="editorRule" ref="createFormRef" @submit.prevent>
      <jm-form-item size="large" label="选择测试流组">
        <jm-select @change="selectChange" v-model="createForm.groupId" placeholder="请选择测试流组">
          <jm-option v-for="item in projectGroupList" :disabled="id === item.id" :key="item.id" :label="item.name"
            :value="item.id">
          </jm-option>
        </jm-select>
      </jm-form-item>

      <div class="selected-list">
        <div class="title">已选测试流</div>
        <div class="selected-list-wrapper">
          <div class="selected-item" v-for="(i, index) in compSelectedList" :key="i.id">
            <!--            <span class="item-name">{{ i.name }}</span>-->
            <jm-text-viewer :value="i.name" class="item-name" />
            <span class="close" @click="removeItem(index)"></span>
          </div>
        </div>
      </div>

      <jm-input placeholder="请输入测试流名称或描述" type="text" size="large" class="search-input" v-model.trim="keyword"
        @keyup.enter="search">
        <template #prepend>
          <el-button :icon="Search" />
        </template>
      </jm-input>

      <div class="card-wrapper">
        <jm-empty v-if="projectList?.list.length === 0" />
        <div v-else :class="['card-item', compSelectedList.some(item => i.id === item.id) ? 'active' : '']"
          v-for="i in projectList?.list" :key="i.id" @click="selectProject(i)">
          <!--          <div class="project-name">{{ i.name }}</div>-->
          <div class="project-name">
            <jm-text-viewer :value="i.name" />
          </div>
          <div class="project-desc">
            {{ i.description || '无' }}
          </div>
          <div class="selected"></div>
        </div>
      </div>
      <div class="page" v-if="projectList && projectList?.list.length !== 0">
        <jm-pagination small layout="prev, pager, next" :page-count="projectList?.pages" :pager-count="5"
          @current-change="pageChange">
        </jm-pagination>
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
import { IPageVo } from '@/api/dto/common';
import { ITestFlowDetail } from '@/api/dto/testflow';
import { ITestflowGroupVo } from '@/api/dto/testflow-group';

import { listTestflowGroup, queryTestFlow, changeTestflowGroup } from '@/api/view-no-auth';
import { IProjectGroupAddingForm } from '@/model/modules/project-group';
import { Mutable } from '@/utils/lib';
import { defineComponent, ref, onMounted, getCurrentInstance, computed } from 'vue';
import { START_PAGE_NUM } from '@/utils/constants';
import { Search } from '@element-plus/icons-vue';

export default defineComponent({
  emits: ['completed'],
  props: {
    id: {
      type: String,
      required: true,
    },
  },
  setup(props, { emit }) {
    const { proxy } = getCurrentInstance() as any;
    const dialogVisible = ref<boolean>(true);
    const createFormRef = ref<any>(null);
    const createForm = ref<IProjectGroupAddingForm>({
      groupId: '',
      testflowIds: [],
    });
    const keyword = ref<string>();
    // 被选中的测试流
    const selectedList = ref<Mutable<ITestFlowDetail[]>>([]);
    const compSelectedList = computed(() => selectedList.value);
    const editorRule = ref<Record<string, any>>({
      groupId: [{ required: true, message: '请选择测试流组', trigger: 'change' }],
    });
    const projectGroupList = ref<ITestflowGroupVo[]>([]);
    // 已选择的测试流数组
    const selectProject = (item: ITestFlowDetail) => {
      selectedList.value.some(i => i.id === item.id)
        ? (selectedList.value = selectedList.value.filter(i => i.id !== item.id))
        : selectedList.value.push(item);
    };
    const projectList = ref<IPageVo<ITestFlowDetail>>();
    const selectChange = async (groupId: string) => {
      projectList.value = await queryTestFlow({
        groupId,
        pageNum: START_PAGE_NUM,
        pageSize: 8,
      });
      selectedList.value = [];
    };
    const pageChange = async (currentPage: number) => {
      projectList.value = await queryTestFlow({
        groupId: createForm.value.groupId,
        pageNum: currentPage,
        pageSize: 8,
        name: keyword.value,
      });
    };
    const search = async () => {
      if (!createForm.value.groupId) {
        proxy.$error('请选择测试流组');
        return;
      }
      projectList.value = await queryTestFlow({
        groupId: createForm.value.groupId,
        pageNum: START_PAGE_NUM,
        pageSize: 8,
        name: keyword.value,
      });
    };
    const removeItem = (index: number) => {
      selectedList.value.splice(index, 1);
    };
    onMounted(async () => {
      projectGroupList.value = await listTestflowGroup();
    });
    const loading = ref<boolean>(false);
    const create = () => {
      createFormRef.value.validate(async (valid: boolean) => {
        if (!valid) {
          return;
        }
        const testflowIds: string[] = [];
        // 获取选中数组中的id
        compSelectedList.value.forEach(item => {
          if (item.id) {
            testflowIds.push(item.id);
          }
        });
        try {
          if (testflowIds.length === 0) {
            proxy.$error('请添加测试流');
            return;
          }
          loading.value = true;
          await changeTestflowGroup({
            groupId: props.id,
            testflowIds,
          });
          proxy.$success('测试流添加成功');
          emit('completed');
          dialogVisible.value = false;
        } catch (err) {
          proxy.$throw(err, proxy);
        } finally {
          loading.value = false;
        }
      });
    };
    return {
      Search,
      createFormRef,
      compSelectedList,
      projectList,
      projectGroupList,
      keyword,
      dialogVisible,
      createForm,
      editorRule,
      loading,
      create,
      selectProject,
      selectChange,
      pageChange,
      search,
      removeItem,
    };
  },
});
</script>

<style scoped lang="less">
.el-dialog {
  .creator-title {
    display: flex;
    align-items: center;

    .edit-icon {
      width: 26px;
      height: 26px;
      margin-right: 10px;
    }
  }

  .el-dialog__body {
    .el-form {
      .selected-list {
        .title {
          font-size: var(--el-form-label-font-size);
          color: var(--el-text-color-regular);
        }

        .selected-list-wrapper {
          margin-top: 10px;
          display: flex;
          flex-wrap: wrap;
          min-height: 65px;
          width: 100%;
          border: 1px solid #b9cfe6;
          padding-top: 10px;

          .selected-item {
            margin: 0px 0px 10px 15px;
            width: 169px;
            height: 22px;
            font-size: 12px;
            border: 1px solid #b9cfe6;
            background-color: #f5f5f5;
            box-sizing: border-box;
            display: flex;
            padding: 0 5px;
            align-items: center;
            justify-content: space-between;

            .item-name {
              width: 100%;
            }

            .close {
              width: 10px;
              height: 10px;
              background-image: url('@/assets/svgs/sort/close.svg');
              background-repeat: no-repeat;
              background-size: contain;
              cursor: pointer;
            }
          }
        }
      }

      ::v-deep(.search-input) {
        margin: 24px 0px 20px;

        .el-input__prefix {
          display: flex;
          align-items: center;
          margin-left: 18px;
        }
      }

      ::v-deep(.project-group) {
        display: flex;
        flex-direction: column;

        .el-form-item__label {
          text-align: left;
        }

        .el-form-item__content {
          .el-select {
            width: 50%;

            .el-input__suffix {
              .el-icon-arrow-up {
                &::before {
                  content: '';
                  display: inline-block;
                  width: 14px;
                  height: 14px;
                  background-image: url('@/assets/svgs/node-library/drop-down.svg');
                  background-repeat: repeat;
                  background-size: contain;
                  transform: rotate(180deg);
                }
              }
            }
          }
        }
      }

      .card-wrapper {
        display: flex;
        flex-wrap: wrap;

        .card-item {
          position: relative;
          cursor: pointer;

          &.active {
            border-color: #096dd9;

            .selected {
              display: block;
              width: 20px;
              background-image: url('@/assets/svgs/sort/selected.svg');
              background-repeat: repeat;
              background-size: contain;
              height: 20px;
              position: absolute;
              right: 0;
              top: 0;
            }
          }

          box-shadow: 0px 0px 12px 4px #edf1f8;
          box-sizing: border-box;
          border: 1px solid transparent;
          padding: 20px;
          width: 24%;
          margin: 0.5% 0.5% 20px;
          min-width: 270px;
          height: 170px;

          &:hover {
            border-color: #096dd9;
          }

          .project-name {
            max-width: 230px;
            font-size: 20px;
            font-weight: 500;
            color: #082340;
          }

          .project-desc {
            max-width: 230px;
            font-size: 14px;
            font-weight: 400;
            color: #6b7b8d;
            margin-top: 10px;
          }

          .selected {
            display: none;
          }
        }
      }

      .page {
        display: flex;
        justify-content: flex-end;

        // margin-bottom: 40px;
        ::v-deep(.el-pagination) {
          margin-right: -13px;

          .btn-prev,
          .btn-next {
            height: 24px;
            min-width: 24px;
            line-height: 24px;
            box-sizing: border-box;
            border: none;
          }

          .el-pager {
            height: 24px;

            .number {
              border: none;
              min-width: 24px;
              line-height: 24px;
              height: 24px;

              &.active {
                border: 1px solid #096dd9;
              }
            }
          }
        }
      }
    }
  }
}
</style>
