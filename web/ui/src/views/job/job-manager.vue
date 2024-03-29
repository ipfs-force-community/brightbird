<template>
  <jm-scrollbar>
    <router-view v-if="childRoute"></router-view>
    <div class="group-manager" v-else>
      <div class="right-top-btn">
        <router-link :to="{ name: 'index' }">
          <jm-button type="primary" class="jm-icon-button-cancel" size="small"
          >关闭
          </jm-button
          >
        </router-link>
      </div>
      <div class="menu-bar">
        <button class="add" @click="add">
          <div class="label">新建Job</div>
        </button>
      </div>
      <div class="title">
        <div>
          <span>Job</span>
          <span class="desc">（共有 {{ jobList.length }} 个组）</span>
        </div>
      </div>
      <div class="content" v-loading="loading" ref="contentRef">
        <jm-empty v-if="jobList.length === 0"/>
        <div  v-else class="item" v-for="i in jobList" :key="i.id">
          <div class="wrapper">
            <div class="top">
              <router-link :to="{ path: `/job/detail/${i.id}` }">
                <div class="name">
                  <jm-text-viewer :value="i.name" :threshold="10"/>
                </div>
              </router-link>
              <div class="operation">
                <div
                  class="run op-item"
                  @click="
                    run(i.id)
                  "
                ></div>
                <div
                  class="edit op-item"
                  @click="
                    toEdit(i.id)
                  "
                ></div>
                <div
                  class="delete op-item"
                  @click="toDelete(i.name, i.id)"
                ></div>
              </div>
            </div>
            <div class="description">
              <jm-text-viewer class="text-viewer" :value="(i.description || '无')"/>
            </div>
            <div class="update-time">
              <span>最后修改时间：</span
              ><span>{{ datetimeFormatter(i.modifiedTime) }}</span>
            </div>
          </div>
        </div>
      </div>
      <job-creator
        v-if="creationActivated"
        @closed="creationActivated = false"
        @completed="addCompleted"
      />
      <job-editor
        :id="jobId || ''"
        v-if="editionActivated"
        @closed="editionActivated = false"
        @completed="editCompleted"
      />
    </div>
  </jm-scrollbar>
</template>

<script lang="ts">
import {
  defineComponent,
  ref,
  getCurrentInstance,
  onMounted,
  Ref,
} from 'vue';
import JobCreator from './job-creator.vue';
import JobEditor from './job-editor.vue';
import { datetimeFormatter } from '@/utils/formatter';
import { IJobVo } from '@/api/dto/job';
import { getTask } from '@/api/tasks';

import {
  listJobs,
  deleteJob,
  execImmediately,
} from '@/api/job';
import { Mutable } from '@/utils/lib';
import {
  onBeforeRouteUpdate,
  RouteLocationNormalized,
  RouteLocationNormalizedLoaded,
  useRoute,
  useRouter,
} from 'vue-router';


export default defineComponent({
  components: {
    JobCreator,
    JobEditor,
  },
  setup() {
    const { proxy } = getCurrentInstance() as any;
    const loading = ref<boolean>();
    const contentRef = ref<HTMLElement>();
    const spacing = ref<string>('');
    const creationActivated = ref<boolean>(false);
    const editionActivated = ref<boolean>(false);
    const jobList = ref<Mutable<IJobVo>[]>([]);
    const currentItem = ref<string>('-1');
    const currentSelected = ref<boolean>(false);
    const router = useRouter();

    const jobId = ref<string>();

    function changeView(
      childRoute: Ref<boolean>,
      route: RouteLocationNormalizedLoaded | RouteLocationNormalized,
    ) {
      childRoute.value = route.matched.length > 2;
    }

    const fetchJobList = async () => {
      loading.value = true;
      try {
        jobList.value = await listJobs()??[];
      } catch (err) {
        proxy.$throw(err, proxy);
      } finally {
        loading.value = false;
      }
    };
    onMounted(async () => {
      await fetchJobList();
    });
    const addCompleted = async () => {
      await fetchJobList();
    };
    const editCompleted = async () => {
      await fetchJobList();
    };
    const add = () => {
      creationActivated.value = true;
    };

    const run = async (id:string) => {
      try {
        const taskId = await execImmediately(id);
        const task = await getTask({ ID:taskId });
        proxy.$alert(`create new task ${task.name}`);
        router.push({ name: 'job-detail', params: { id } });
      } catch (err) {
        proxy.$throw(err, proxy);
      } finally {
        loading.value = false;
      }
    };
    const toEdit = (
      id: string,
    ) => {
      jobId.value = id;
      editionActivated.value = true;
    };
    const toDelete = async (name: string, jobId: string) => {
      let msg = '<div>确定要删除Job吗?</div>';
      msg += `<div style="margin-top: 5px; font-size: 12px; line-height: normal;">名称：${name}</div>`;

      proxy
        .$confirm(msg, '删除Job', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning',
          dangerouslyUseHTMLString: true,
        })
        .then(async () => {
          try {
            await deleteJob(jobId);
            proxy.$success('Job删除成功');
            await fetchJobList();
          } catch (err) {
            proxy.$throw(err, proxy);
          }
        })
        .catch(() => {
          // eslint-disable-next-line @typescript-eslint/no-empty-function
        });
    };
    const childRoute = ref<boolean>(false);
    changeView(childRoute, useRoute());
    onBeforeRouteUpdate(to => changeView(childRoute, to));
    return {
      spacing,
      contentRef,
      childRoute,
      jobId,
      start(e: any) {
        currentSelected.value = true;
        currentItem.value = e.item.getAttribute('_id');
      },
      currentSelected,
      addCompleted,
      editCompleted,
      datetimeFormatter,
      jobList,
      loading,
      creationActivated,
      editionActivated,
      add,
      run,
      toEdit,
      toDelete,
    };
  },
});
</script>

<style scoped lang="less">
.group-manager {
  padding: 15px;
  background-color: #ffffff;
  margin-bottom: 20px;

  .right-top-btn {
    position: fixed;
    right: 20px;
    top: 78px;

    .jm-icon-button-cancel::before {
      font-weight: bold;
    }
  }

  .menu-bar {
    button {
      position: relative;

      .label {
        position: absolute;
        left: 0;
        bottom: 40px;
        width: 100%;
        text-align: center;
        font-size: 18px;
        color: #b5bdc6;
      }

      &.add {
        margin: 0.5%;
        width: 19%;
        min-width: 260px;
        height: 170px;
        background-color: #ffffff;
        border: 1px dashed #b5bdc6;
        background-image: url('@/assets/svgs/btn/add.svg');
        background-position: center 45px;
        background-repeat: no-repeat;
        cursor: pointer;
      }
    }
  }

  .title {
    font-size: 20px;
    font-weight: bold;
    color: #082340;
    position: relative;
    padding-left: 20px;
    padding-right: 20px;
    margin: 30px -15px 20px;
    display: flex;
    justify-content: space-between;
    align-items: center;

    .move {
      cursor: pointer;
      width: 24px;
      height: 24px;
      background-image: url('@/assets/svgs/sort/move.svg');
      background-size: contain;

      &.active {
        background-image: url('@/assets/svgs/sort/move-active.svg');
      }
    }

    .desc {
      font-weight: normal;
      margin-left: 12px;
      font-size: 14px;
      color: #082340;
      opacity: 0.46;
    }
  }

  .content {
    display: flex;
    flex-wrap: wrap;

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
          border: 2px solid #096DD9;
          background: rgba(9, 109, 217, 0.3);
          position: absolute;
          top: 0;
          left: 20%;
        }
      }
    }

    .list {
      display: flex;
      flex-wrap: wrap;
      width: 100%;

      .item {
        cursor: move;

        &.move {
          .wrapper {
            position: relative;
            border-color: #096dd9;

            .cover {
              position: absolute;
              top: 0;
              left: 0;
              width: 100%;
              height: 170px;
              background-color: rgba(140, 140, 140, 0.3);

              .drag-icon {
                position: absolute;
                right: 0;
                bottom: 0;
                width: 30px;
                height: 30px;
                background-image: url('@/assets/svgs/sort/drag.svg');
                background-size: contain;
              }
            }
          }
        }
      }
    }

    .item {
      position: relative;
      margin: 0.5%;
      width: 19%;
      min-width: 260px;
      height: 170px;
      background-color: #ffffff;
      box-shadow: 0px 0px 8px 4px #eff4f9;

      .wrapper {
        padding: 15px;
        border: 1px solid transparent;
        height: 138px;
        color: #6b7b8d;
        font-size: 13px;
        position: relative;

        &:hover {
          border-color: #096dd9;
          box-shadow: 0px 6px 16px 4px #e6eef6;

          .top {
            a {
              max-width: calc(100% - 59px);
            }

            .operation {
              display: flex;
            }
          }
        }

        .top {
          display: flex;
          justify-content: space-between;
          align-items: center;
          white-space: nowrap;

          a {
            flex: 1;
          }

          .name {
            color: #082340;
            font-size: 20px;
            font-weight: 500;

            &:hover {
              color: #096dd9;
            }
          }

          .operation {
            margin-left: 5px;
            display: none;

            .op-item {
              width: 22px;
              height: 22px;
              background-size: contain;
              cursor: pointer;

              &:active {
                background-color: #eff7ff;
                border-radius: 4px;
              }

              &.run {
                background-image: url('@/assets/svgs/btn/rocketstart.svg');
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

        .description {
          margin-top: 5px;

          .text-viewer {
            height: 54px;
          }

          line-height: 20px;
        }

        .update-time {
          position: absolute;
          bottom: 38px;
        }

        .switch {
          position: absolute;
          bottom: 10px;
          left: 15px;
          display: flex;
          align-items: center;

          span {
            margin-left: 5px;
            opacity: 0.5;
          }
        }

        .total {
          position: absolute;
          bottom: 12px;
          right: 25px;
          font-weight: 400;
          text-align: end;
          margin-top: 10px;

          .count {
            color: #096dd9;
          }
        }
      }
    }
  }
}
</style>
