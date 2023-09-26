<template>
  <div>
    <div class="content" v-loading="loading">
      <folding :status="toggle">
        <template #prefix>
          <span class="prefix-wrapper">
            <i
              :class="[
                'jm-icon-button-right',
                'prefix',
                toggle ? 'rotate' : '',
              ]"
              :disabled="pageData.total === 0"
              @click="saveFoldStatus(toggle, jobVo?.id)"
            />
          </span>
        </template>
        <template #title>
          <div class="title">
            <div class="left">
              <div>{{ jobVo?.name }}</div>
              <div>
                {{
                  `(${jobVo?.description === "" ? "无" : jobVo?.description})`
                }}
              </div>
            </div>
            <div class="right">
              <span>最后修改时间:{{ datetimeFormatter(jobVo?.modifiedTime) }}</span>
              <span>类型:{{ jobVo?.jobType }}</span>
              <div class="operation">
                <ElButton @click="run(jobVo?.id)" type="primary"><div class="run op-item" /></ElButton>
               <el-button @click="toEdit(jobVo?.id)" type="primary" > <div class="edit op-item"/></el-button>
               <el-button @click="toDelete(jobVo?.name, jobVo?.id)" type="primary"> <div class="delete op-item"/></el-button>
              </div>
            </div>
          </div>
        </template>
        <template #default>
          <div>
            <div class="job-wrapper" v-show="toggle && pageData.total > 0">
            <jm-empty
              description="暂无任务"
              :image-size="98"
              v-if="pageData.total === 0"
            />
            <task-item
              class="task"
              v-else
              v-for="task of pageData.tasks"
              :key="task.id"
              :task="task"
            />
          </div>
          <!-- 显示更多 -->
          <div class="load-more" v-show="toggle">
            <el-pagination :total="pageData.total" :page-size="pageSize" @current-change="btnDown">
            </el-pagination>
          </div>
          </div>
        </template>
      </folding>
    </div>
  </div>
</template>
<script lang="ts">
import { getCurrentInstance, onMounted, ref, PropType, computed, onUnmounted } from 'vue';
import { ITaskVo } from '@/api/dto/tasks';
import {
  execImmediately,
  getJobDetail,
  listJobs,
  nextNTime,
} from '@/api/job';
import { IJobDetailVo } from '@/api/dto/job';
import { getTask, getTaskInJob } from '@/api/tasks';
import { StateEnum } from '@/components/load-more/enumeration';
import { IJobVo } from '@/api/dto/job';
import TaskItem from '@/views/common/task-item.vue';
import Folding from '@/views/common/folding.vue';
import { createNamespacedHelpers, useStore } from 'vuex';
import { namespace } from '@/store/modules/test-flow';
import { useRouter } from 'vue-router';
import { Mutable } from '@/utils/lib';
import { TaskStateEnum } from '@/api/dto/enumeration';
import { datetimeFormatter } from '@/utils/formatter';
import { ElPagination, ElButton } from 'element-plus';
export default {
  components: { TaskItem, Folding, ElPagination, ElButton },
  emits:['toEdit', 'toDelete'],
  props: {
    jobVo: {
      type: Object as PropType<IJobVo>,
      required: true,
    },
  },
  setup(props: any, { emit }) {
    const pageSize = 15;
    const store = useStore();
    const loading = ref<boolean>();
    const creationActivated = ref<boolean>(false);
    const jobId = ref<string>();
    const projectGroupFoldingMapping = store.state[namespace];
    const { mapMutations } = createNamespacedHelpers(namespace);
    const { proxy } = getCurrentInstance() as any;
    const loadingTop = ref<boolean>(false);
    const jobDetail = ref<IJobDetailVo>();
    const initialized = ref<boolean>(false);
    const next = ref<Date[]>([]);
    const loadState = ref<StateEnum>(StateEnum.NONE);
    const jobList = ref<Mutable<IJobVo>[]>([]);
    const pageData = ref<{
      pageNum: number;
      pageSize: number;
      total: number;
      pages: number;
      tasks: ITaskVo[];
    }>({
      pageNum: 1,
      pageSize: 10,
      pages: 0,
      total: 0,
      tasks: [],
    });

    const toggle = computed<boolean>(() => {
      if (projectGroupFoldingMapping[props.jobVo?.id] === undefined) {
        return false;
      }
      return projectGroupFoldingMapping[props.jobVo?.id];
    });
    const fetchJobDetail = async () => {
      try {
        loadingTop.value = true;
        jobDetail.value = await getJobDetail(props.jobVo?.id);
        const queryTask = await getTaskInJob({
          jobId: props.jobVo?.id,
          pageNum: 1,
          pageSize: pageSize,
        });
        pageData.value.total = queryTask.total;
        pageData.value.pages = queryTask.pages;
        pageData.value.tasks = queryTask.list;
        initialized.value = true;        
        if (queryTask.pages === 1) {
          loadState.value = StateEnum.NO_MORE;
        } else {
          loadState.value = StateEnum.MORE;
        }
        const nextN = await nextNTime({ id: props.jobVo?.id, n: 3 });
        next.value = nextN.map(a => new Date(a * 1000));
      } catch (err) {
        proxy.$throw(err, proxy);
      } finally {
        loadingTop.value = false;
      }
    };

    const fetchJobList = async () => {
      loading.value = true;
      try {
        jobList.value = (await listJobs()) ?? [];
      } catch (err) {
        proxy.$throw(err, proxy);
      } finally {
        loading.value = false;
      }
    };

    // 更新状态 
    const timer =  setInterval(async () => {
      try {
        for (var i = 0; i < pageData.value.tasks.length; i++) {
          if (pageData.value.tasks[i].state !== TaskStateEnum.Error && pageData.value.tasks[i].state !== TaskStateEnum.Successful) {
            pageData.value.tasks[i] = await getTask({ ID:pageData.value.tasks[i].id });
          }
        }
      } catch (err) {
        proxy.$throw(err, proxy);
      }
    }, 10000);

    const btnDown = async value => {      
      try {
        const queryTask = await getTaskInJob({
          jobId: props.jobVo?.id,
          pageNum: value,
          pageSize:pageSize,
        });
        pageData.value.tasks = queryTask.list;
        pageData.value.total = queryTask.total;
        pageData.value.pages = queryTask.pages;
        if (pageData.value.pageNum === pageData.value.pages) {
          loadState.value = StateEnum.NO_MORE;
        }
      } catch (err) {
        proxy.$throw(err, proxy);
      } finally {
        // bottomLoading.value = false;
      }
    };

    const saveFoldStatus = (status: boolean, id?: string) => {
      // 改变状态
      const toggle = !status;
      // 调用vuex的mutations更改对应测试流组的状态
      proxy.mutate({
        id,
        status: toggle,
      });
    };

    const run = async (id: string) => {
      try {
        const taskId = await execImmediately(id);
        await getTask({ ID: taskId });
        await fetchJobDetail();
      } catch (err) {
        proxy.$throw(err, proxy);
      } finally {
        loading.value = false;
      }
    };
    const toEdit = (id: string) => {
      emit('toEdit', { id });
    };
    const toDelete = async (name: string, jobId: string) => {
      emit('toDelete', { name, jobId });
    };
    onMounted(async () => {
      await fetchJobDetail();
    });
    onUnmounted(() => {
      clearInterval(timer);
    });
    return {
      pageSize,
      creationActivated,
      jobId,
      loading,
      loadState,
      pageData,
      toggle,
      projectGroupFoldingMapping,
      ...mapMutations({
        mutate: 'mutate',
      }),
      btnDown,
      saveFoldStatus,
      run,
      toEdit,
      toDelete,
      fetchJobList,
      datetimeFormatter,
    };
  },
};
</script>
<style scoped lang="less">
.prefix-wrapper {
  display: flex;
  align-items: center;
  cursor: not-allowed;
}

.prefix {
  color: #6b7b8d;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.1s linear;

  &[disabled="true"] {
    color: #a7b0bb;
    pointer-events: none;
  }

  &.rotate {
    transition: all 0.1s linear;
    transform: rotate(90deg);
  }
}
.content {
  padding: 0px 15px 0px;
  background-color: #ffffff;

  .title {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0px 20px;
    color: #082340;
    font-weight: bold;
    font-size: 18px;

    .left {
      & > :last-child {
        padding-left: 12px;
        color: #082340;
        font-weight: normal;
        font-size: 14px;
        opacity: 0.46;
      }
    }
    .right {
      display: flex;
      align-items: center;
      color: #082340;
      font-weight: normal;
      font-size: 14px;
      
      column-gap: 10px;

      span {
        opacity: 0.46;
      }
      .operation {
        display: flex;
        margin-left: 40px;

        .op-item {
          width: 22px;
          height: 22px;
          background-size: contain;
          cursor: pointer;

          &:active {
            border-radius: 4px;
            background-color: #eff7ff;
          }

          &.run {
            background-image: url("@/assets/svgs/btn/rocketstart.svg");
          }
          &.edit {
            background-image: url("@/assets/svgs/btn/edit2.svg");
          }

          &.delete {
            background-image: url("@/assets/svgs/btn/del3.svg");
          }
        }
      }
    }
  }

  .job-wrapper {
    display: flex;
    flex-wrap: wrap;

    .task {
      margin-top: -10px;
    }
  }

  .load-more {
    display: flex;
    justify-content: center;
    text-align: center;
    padding-bottom: 20px;
  }
}
</style>
