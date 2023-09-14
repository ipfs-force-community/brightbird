<template>
  <div class="project-group" v-loading="loading">
    <folding :status="toggle">
      <template #prefix>
        <span class="prefix-wrapper">
          <i
            :class="['jm-icon-button-right', 'prefix', toggle ? 'rotate' : '']"
            :disabled="projectPage.total === 0"
            @click="saveFoldStatus(toggle, testflowGroup?.id)"
          />
        </span>
      </template>
      <template #title>
        <div class="title">
          <div class="left">
            <div class="group-name">
              <router-link
                :to="{ path: `/project-group/detail/${testflowGroup?.id}` }"
                >{{ testflowGroup?.name }}
              </router-link>
              <div class="description">
                {{ ('('+ (testflowGroup?.description || '无') + ')' )  }}
              </div>
            </div>
          </div>
          <div class="right">
            <span class="desc"
              >（共有
              {{ projectPage.total >= 0 ? projectPage.total : 0 }}
              个测试流）</span
            >
            <div class="update-time">
              <span>最后修改时间：</span
              ><span>{{ datetimeFormatter(testflowGroup?.modifiedTime) }}</span>
            </div>
            <div class="operation">
                <div
                  class="edit op-item"
                  @click="
                    toEdit(
                      testflowGroup?.id,
                      testflowGroup?.name,
                      testflowGroup?.isShow,
                      testflowGroup?.description
                    )
                  "
                ></div>
                <div
                  class="delete op-item"
                  @click="toDelete(testflowGroup?.name, testflowGroup?.id)"
                ></div>
              </div>
          </div>
        </div>
      </template>
      <template #default>
        <div>
          <div class="testflows" v-show="toggle && testflows.length > 0">
            <jm-empty
              description="暂无测试流"
              :image-size="98"
              v-if="testflows.length === 0"
            />
            <project-item
              v-else
              v-for="project of testflows"
              :key="project.id"
              :project="project"
              @triggered="handleProjectTriggered"
              @synchronized="handleProjectSynchronized"
              @deleted="handleProjectDeleted"
              @terminated="handleProjectTerminated"
            />
          </div>
        </div>
      </template>
    </folding>
  </div>
  <group-editor
        :name="groupName || ''"
        :description="groupDescription"
        :is-show="showInHomePage"
        :id="groupId || ''"
        v-if="editionActivated"
        @closed="editionActivated = false"
        @completed="editCompleted"
      />
</template>

<script lang="ts">
import {
  computed,
  defineComponent,
  getCurrentInstance,
  nextTick,
  onBeforeMount,
  onUpdated,
  PropType,
  ref,
} from 'vue';
import { ITestFlowDetail } from '@/api/dto/testflow';
import { ITestflowGroupVo } from '@/api/dto/testflow-group';
import { queryTestFlow } from '@/api/view-no-auth';
import { IQueryForm } from '@/model/modules/project';
import ProjectItem from '@/views/common/project-item.vue';
import { IPageVo } from '@/api/dto/common';
import { Mutable } from '@/utils/lib';
import { START_PAGE_NUM } from '@/utils/constants';
import Folding from '@/views/common/folding.vue';
import { createNamespacedHelpers, useStore } from 'vuex';
import { namespace } from '@/store/modules/project-group';
import noDataImg from '@/assets/svgs/index/no-data.svg';
import sleep from '@/utils/sleep';
import { datetimeFormatter } from '@/utils/formatter';
import GroupEditor from '@/views/project-group/project-group-editor.vue';
import { eventBus } from '@/main';
import { deleteProjectGroup } from '@/api/testflow-group';

export default defineComponent({
  components: { ProjectItem, Folding, GroupEditor },
  props: {
    // 测试流组
    testflowGroup: {
      type: Object as PropType<ITestflowGroupVo>,
    },
    // 查询关键字
    name: {
      type: String,
    },
    // 是否开启移动模式
    move: {
      type: Boolean,
      default: false,
    },
  },
  setup(props: any) {
    const store = useStore();
    const { mapMutations } = createNamespacedHelpers(namespace);
    const groupName = ref<string>();
    const editionActivated = ref<boolean>(false);
    const groupId = ref<string>();
    const groupDescription = ref<string>();
    const showInHomePage = ref<boolean>(false);
    const projectGroupFoldingMapping = store.state[namespace];
    // 根据测试流组在vuex中保存的状态，进行展开、折叠间的切换
    const toggle = computed<boolean>(() => {
      // 只有全等于为undefined说明该测试流组一开始根本没有做折叠操作
      if (projectGroupFoldingMapping[props.testflowGroup?.id] === undefined) {
        return true;
      }
      return projectGroupFoldingMapping[props.testflowGroup.id];
    });

    const { proxy } = getCurrentInstance() as any;
    const loading = ref<boolean>(false);
    const projectPage = ref<Mutable<IPageVo<ITestFlowDetail>>>({
      total: -1,
      pages: 0,
      list: [],
      pageNum: START_PAGE_NUM,
    });
    const testflows = computed<ITestFlowDetail[]>(() => projectPage.value.list);

    const queryForm = ref<IQueryForm>({
      pageNum: START_PAGE_NUM,
      pageSize: 40,
      groupId: props.testflowGroup?.id,
      name: props.name,
    });
    // 保存单个测试流组的展开折叠状态
    const saveFoldStatus = (status: boolean, id?: string) => {
      // 改变状态
      const toggle = !status;
      // 调用vuex的mutations更改对应测试流组的状态
      proxy.mutate({
        id,
        status: toggle,
      });
    };
    // 重新加载当前已经加载过的测试流
    const reloadCurrentProjectList = async () => {
      try {
        const { pageSize, pageNum } = queryForm.value;
        // 获得当前已经加载了的总数
        const currentCount = pageSize * pageNum;
        projectPage.value = await queryTestFlow({
          ...queryForm.value,
          pageNum: START_PAGE_NUM,
          pageSize: currentCount,
        });
        console.log(projectPage.value);
      } catch (err) {
        proxy.$throw(err, proxy);
      }
    };

    const loadProject = async () => {
      projectPage.value = await queryTestFlow({ ...queryForm.value });
      // 测试流组中测试流为空，将其自动折叠
      if (projectPage.value.total === 0) {
        saveFoldStatus(true, props.testflowGroup?.id);
      }
      return;
    };
    const toEdit = (
      id?: string,
      name?: string,
      isShow?: boolean,
      description?: string,
    ) => {
      groupName.value = name;
      groupDescription.value = description;
      groupId.value = id;
      showInHomePage.value = isShow ?? false;
      editionActivated.value = true;      
    };
    const toDelete = async (name?: string, groupId?: string) => {
      let msg = '<div>确定要删除分组吗?</div>';
      msg += `<div style="margin-top: 5px; font-size: 12px; line-height: normal;">名称：${name}</div>`;

      proxy
        .$confirm(msg, '删除分组', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning',
          dangerouslyUseHTMLString: true,
        })
        .then(async () => {
          if (!groupId) { return; }
          try {
            await deleteProjectGroup(groupId);
            proxy.$success('测试流分组删除成功');
            eventBus.emit('newGroup');
          } catch (err) {
            proxy.$throw(err, proxy);
          }
        })
        // eslint-disable-next-line @typescript-eslint/no-empty-function
        .catch(() => {
        });
    };
    const editCompleted = () => {
      eventBus.emit('newGroup');
    };
    // 初始化测试流列表
    onBeforeMount(async () => {
      await nextTick(() => {
        queryForm.value.name = props.name;
      });
      await loadProject();
    });
    onUpdated(async () => {
      if (
        queryForm.value.name === props.name &&
        queryForm.value.groupId === props.testflowGroup?.id
      ) {
        return;
      }
      queryForm.value.name = props.name;
      queryForm.value.groupId = props.testflowGroup?.id;
    });
    const currentItem = ref<string>('');

    reloadCurrentProjectList(); // init
    return {
      editCompleted,
      toEdit,
      toDelete,
      datetimeFormatter,
      noDataImg,
      ...mapMutations({
        mutate: 'mutate',
      }),
      toggle,
      loading,
      projectPage,
      testflows,
      queryForm,
      handleProjectSynchronized: async () => {
        // 刷新测试流列表，保留查询状态
        await reloadCurrentProjectList();
      },
      handleProjectDeleted: (id: string) => {
        const index = testflows.value.findIndex(item => item.id === id);
        testflows.value.splice(index, 1);
      },
      handleProjectTriggered: async (id: string) => {
        await sleep(400);
        // 刷新测试流列表，保留查询状态
        await reloadCurrentProjectList();
      },
      handleProjectTerminated: async (id: string) => {
        // 刷新测试流列表，保留查询状态
        await reloadCurrentProjectList();
      },
      saveFoldStatus,
      projectGroupFoldingMapping,
      groupName,
      editionActivated,
      groupId,
      groupDescription,
      showInHomePage,
    };
  },
});
</script>

<style scoped lang="less">
.project-group {
  margin-top: 24px;

  .prefix-wrapper {
    cursor: not-allowed;
    display: flex;
    align-items: center;
  }

  .prefix {
    cursor: pointer;
    font-size: 12px;
    transition: all 0.1s linear;
    color: #6b7b8d;

    &[disabled="true"] {
      pointer-events: none;
      color: #a7b0bb;
    }

    &.rotate {
      transform: rotate(90deg);
      transition: all 0.1s linear;
    }
  }

  .title {
    display: flex;
    justify-content: space-between;
    align-items: center;
    .left {
      margin-left: 10px;
      font-size: 18px;
      font-weight: bold;
      color: #082340;
      display: flex;
      justify-content: space-between;
      align-items: flex-end;
      padding-right: 0.7%;

      .group-name {
        display: flex;
        align-items: center;
        .description {
          padding-left: 12px;
          font-size: 14px;
          font-weight: normal;
          color: #082340;
          opacity: 0.46;
        }
      }
      .more-container {
        width: 86px;
        height: 24px;
        background: #eff7ff;
        border-radius: 15px;
        font-size: 12px;
        font-weight: 400;
        cursor: pointer;
        display: flex;
        justify-content: center;
        align-items: center;

        a {
          color: #6b7b8d;
          line-height: 24px;
        }

        .more-icon {
          display: inline-block;
          width: 12px;
          height: 12px;
          text-align: center;
          line-height: 12px;
          background: url("@/assets/svgs/btn/more.svg") no-repeat;
          position: relative;
          top: 1.4px;
          right: 0px;
        }

        &:hover {
          color: #096dd9;

          a {
            color: #096dd9;
          }

          .more-icon {
            background: url("@/assets/svgs/btn/more-active.svg") no-repeat;
          }
        }
      }
    }

    .right {
      display: flex;
      margin-left: 12px;
      font-size: 14px;
      font-weight: normal;
      color: #082340;
      opacity: 0.46;

      .operation {
            margin-left: 5px;
            display: flex;

            .op-item {
              width: 22px;
              height: 22px;
              background-size: contain;
              cursor: pointer;

              &:active {
                background-color: #eff7ff;
                border-radius: 4px;
              }

              &.edit {
                background-image: url('@/assets/svgs/btn/edit.svg');
              }

              &.delete {
                margin-left: 15px;
                background-image: url('@/assets/svgs/btn/del.svg');
              }
            }
          }
    }
  }

  .testflows {
    display: flex;
    flex-wrap: wrap;

    .el-empty {
      padding-top: 102px;
    }

    ::v-deep(.jm-sorter) {
      .drag-target-insertion {
        // width: v-bind(spacing);
        border-width: 0;
        background-color: transparent;

        &::after {
          content: "";
          width: 60%;
          height: 100%;
          box-sizing: border-box;
          border: 2px solid #096dd9;
          background: rgba(9, 109, 217, 0.3);
          position: absolute;
          top: 0;
          left: 20%;
        }
      }
    }

    .list {
      width: 100%;
      display: flex;
      flex-wrap: wrap;
    }
  }

  .load-more {
    margin: 10px auto 0px;
    display: flex;
    justify-content: center;
  }
}
</style>
