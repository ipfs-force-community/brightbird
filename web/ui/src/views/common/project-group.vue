<template>
  <div class="project-group" v-loading="loading">
    <folding :status="toggle">
      <template #prefix>
        <span class="prefix-wrapper">
          <i
              :class="['jm-icon-button-right', 'prefix', toggle ? 'rotate' : '']"
              :disabled="projectPage.total === 0"
              @click="saveFoldStatus(toggle, projectGroup.id)"
          />
        </span>
      </template>
      <template #title>
        <div class="name">
          <div class="group-name">
            <router-link :to="{ path: `/project-group/detail/${projectGroup?.id}` }"
            >{{ projectGroup?.name }}
            </router-link>
            <span class="desc">（共有 {{ projectPage.total >= 0 ? projectPage.total : 0 }} 个测试流）</span>
          </div>
        </div>
      </template>
      <template #default>
        <div>
          <div class="projects" v-show="toggle&&projects.length > 0">
            <jm-empty description="暂无测试流" :image-size="98" v-if="projects.length === 0" />
            <project-item
              v-else
              v-for="project of projects"
              :concurrent="project.concurrent"
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
</template>

<script lang="ts">
import {
  computed,
  defineComponent,
  getCurrentInstance,
  inject,
  nextTick,
  onBeforeMount,
  onBeforeUnmount,
  onUpdated,
  PropType,
  ref,
  watch,
} from 'vue';
import { IProjectVo } from '@/api/dto/project';
import { IProjectGroupVo } from '@/api/dto/project-group';
import { queryProject } from '@/api/view-no-auth';
import { IQueryForm } from '@/model/modules/project';
import ProjectItem from '@/views/common/project-item.vue';
import { IPageVo } from '@/api/dto/common';
import { Mutable } from '@/utils/lib';
import { START_PAGE_NUM } from '@/utils/constants';
import Folding from '@/views/common/folding.vue';
import { createNamespacedHelpers, useStore } from 'vuex';
import { namespace } from '@/store/modules/project-group';
import JmSorter from '@/components/sorter/index.vue';
import noDataImg from '@/assets/svgs/index/no-data.svg';
import sleep from '@/utils/sleep';

const MAX_AUTO_REFRESHING_OF_NO_RUNNING_COUNT = 5;

export default defineComponent({
  components: { JmSorter, ProjectItem, Folding },
  props: {
    // 测试流组
    projectGroup: {
      type: Object as PropType<IProjectGroupVo>,
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
    const projectGroupFoldingMapping = store.state[namespace];
    // 根据测试流组在vuex中保存的状态，进行展开、折叠间的切换
    const toggle = computed<boolean>(() => {
      // 只有全等于为undefined说明该测试流组一开始根本没有做折叠操作
      if (projectGroupFoldingMapping[props.projectGroup?.id] === undefined) {
        return true;
      }
      return projectGroupFoldingMapping[props.projectGroup.id];
    });

    const { proxy } = getCurrentInstance() as any;
    const loading = ref<boolean>(false);
    const projectPage = ref<Mutable<IPageVo<IProjectVo>>>({
      total: -1,
      pages: 0,
      list: [],
      pageNum: START_PAGE_NUM,
    });
    const projects = computed<IProjectVo[]>(() => projectPage.value.list);

    const queryForm = ref<IQueryForm>({
      pageNum: START_PAGE_NUM,
      pageSize:  40 ,
      groupId: props.projectGroup?.id,
      name: props.name,
    });
    // 保存单个测试流组的展开折叠状态
    const saveFoldStatus = (status: boolean, id: string) => {
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
        projectPage.value =  await queryProject({ ...queryForm.value, pageNum: 1, pageSize: currentCount });
        console.log(projectPage.value)
      } catch (err) {
        proxy.$throw(err, proxy);
      }
    };

    const loadProject = async () => {
      projectPage.value = await queryProject({ ...queryForm.value });
      // 测试流组中测试流为空，将其自动折叠
      if (projectPage.value.total === 0) {
        saveFoldStatus(true, props.projectGroup?.id);
      }
      return;
    };
    // 初始化测试流列表
    onBeforeMount(async () => {
      await nextTick(() => {
        queryForm.value.name = props.name;
      });
      await loadProject();
    });
    onUpdated(async () => {
      if (queryForm.value.name === props.name && queryForm.value.groupId === props.projectGroup?.id) {
        return;
      }
      queryForm.value.name = props.name;
      queryForm.value.groupId = props.projectGroup?.id;
    });
    const currentItem = ref<string>('');

    reloadCurrentProjectList() //init
    return {
      noDataImg,
      ...mapMutations({
        mutate: 'mutate',
      }),
      toggle,
      loading,
      projectPage,
      projects,
      queryForm,
      handleProjectSynchronized: async () => {
        // 刷新测试流列表，保留查询状态
        await reloadCurrentProjectList();
      },
      handleProjectDeleted: (id: string) => {
        const index = projects.value.findIndex(item => item.id === id);
        projects.value.splice(index, 1);
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

    &[disabled='true'] {
      pointer-events: none;
      color: #a7b0bb;
    }

    &.rotate {
      transform: rotate(90deg);
      transition: all 0.1s linear;
    }
  }

  .name {
    margin-left: 10px;
    font-size: 18px;
    font-weight: bold;
    color: #082340;
    display: flex;
    justify-content: space-between;
    align-items: flex-end;
    padding-right: 0.7%;

    .group-name {
      .desc {
        margin-left: 12px;
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
        background: url('@/assets/svgs/btn/more.svg') no-repeat;
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
          background: url('@/assets/svgs/btn/more-active.svg') no-repeat;
        }
      }
    }
  }

  .projects {
    display: flex;
    flex-wrap: wrap;

    .el-empty {
      padding-top: 102px;
    }

    ::v-deep(.jm-sorter) {
      .drag-target-insertion {
        width: v-bind(spacing);
        border-width: 0;
        background-color: transparent;

        &::after {
          content: '';
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
