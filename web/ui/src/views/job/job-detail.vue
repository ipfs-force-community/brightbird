<template>
  <div class="project-group-detail">
    <div class="right-top-btn">
      <jm-button type="primary" class="jm-icon-button-cancel" size="small" @click="close">关闭</jm-button>
    </div>
    <div class="top-card" v-loading="loadingTop">
      <div class="top-title">
        <div class="name">{{ jobDetail?.name }}</div>
      </div>
      <div class="count">
        测试流: {{ jobDetail?.testFlowName }}
      </div>
      <div class="description">
        描述:<span v-html="(jobDetail?.description || '无').replace(/\n/g, '<br/>')" />
      </div>
    </div>
    <div class="content">
      <div class="title">
        <div>
          <span>任务列表</span>
        </div>
      </div>
      <div class="task-wrapper">
        <jm-empty description="暂无任务" :image-size="98" v-if="pageData.total === 0" />
        <task-item class="task" v-else v-for="task of pageData.tasks" :key="task.id" :task="task" />
      </div>
      <!-- 显示更多 -->
      <div class="load-more">
        <jm-load-more :state="loadState" :load-more="btnDown">LoadMore</jm-load-more>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { getJobDetail } from '@/api/job'
import { getTaskInJob } from '@/api/tasks'
import { defineComponent, getCurrentInstance, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useStore } from 'vuex';
import { IRootState } from '@/model';
import { IJobDetailVo } from '@/api/dto/job';

import { ITaskVo } from '@/api/dto/tasks';

import { StateEnum } from '@/components/load-more/enumeration';
import TaskItem from "@/views/common/task-item.vue"
import { TaskStateEnum } from '@/api/dto/enumeration';
import { getTask } from '@/api/tasks';

export default defineComponent({
  components: { TaskItem },
  props: {
    id: {
      type: String,
      required: true,
    },
  },
  setup(props) {
    const { proxy } = getCurrentInstance() as any;
    const router = useRouter();
    const store = useStore();
    const rootState = store.state as IRootState;
    const loadState = ref<StateEnum>(StateEnum.NONE);
    const initialized = ref<boolean>(false);
    const loadingTop = ref<boolean>(false);
    const jobDetail = ref<IJobDetailVo>();
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
    })

    const fetchJobDetail = async () => {
      try {
        loadingTop.value = true;
        jobDetail.value = await getJobDetail(props.id)
        const queryTask = await getTaskInJob({ jobId: props.id, pageNum: 1, pageSize: 10 });
        pageData.value.total = queryTask.total;
        pageData.value.pages = queryTask.pages;
        pageData.value.tasks.push(...queryTask.list);
        initialized.value = true;
        if (queryTask.pages == 1) {
          loadState.value = StateEnum.NO_MORE
        } else {
          loadState.value = StateEnum.MORE
        }
      } catch (err) {
        proxy.$throw(err, proxy);
      } finally {
        loadingTop.value = false;
      }
    };


    // 加载更多
    const btnDown = async () => {
      try {
        if (pageData.value.pageNum < pageData.value.pages) {
          pageData.value.pageNum++;
          const queryTask = await getTaskInJob({ jobId: props.id, pageNum: pageData.value.pageNum, pageSize: 10 });
          pageData.value.tasks.push(...queryTask.list);
          pageData.value.total = queryTask.total;
          pageData.value.pages = queryTask.pages;
          if (pageData.value.pageNum == pageData.value.pages) {
            loadState.value = StateEnum.NO_MORE;
          }
        }
      } catch (err) {
        proxy.$throw(err, proxy);
      } finally {
        // bottomLoading.value = false;
      }
    };

    // 更新状态 
    setInterval(async () => {
      try {
        for (var i = 0; i < pageData.value.tasks.length; i++) {
          if (pageData.value.tasks[i].state != TaskStateEnum.Error && pageData.value.tasks[i].state != TaskStateEnum.Successful) {
            pageData.value.tasks[i] = await getTask(pageData.value.tasks[i].id)
          }
        }
      } catch (err) {
        proxy.$throw(err, proxy);
      } 
    },5000)


    onMounted(async () => {
      await fetchJobDetail();
    });
    return {
      initialized,
      pageData,
      loadState,
      btnDown,
      loadingTop,
      close: () => {
        if (!['/', '/job'].includes(rootState.fromRoute.path)) {
          router.push({ name: 'index' });
          return;
        }
        router.push(rootState.fromRoute.fullPath);
      },
      jobDetail,
    };
  },
});
</script>

<style scoped lang="less">
.project-group-detail {
  margin-bottom: 20px;

  .right-top-btn {
    position: fixed;
    right: 20px;
    top: 78px;

    .jm-icon-button-cancel::before {
      font-weight: bold;
    }
  }

  .top-card {
    min-height: 58px;
    font-size: 14px;
    padding: 24px;
    background-color: #ffffff;

    .top-title {
      display: flex;
      align-items: center;
      color: #082340;

      .name {
        font-size: 40px;
        font-weight: 500;
      }
    }

    .description {
      margin-top: 10px;
      color: #6b7b8d;
    }
  }

  .content {
    margin-top: 20px;
    padding: 15px 15px 0px;
    background-color: #ffffff;

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
          // margin: 0.5%;
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
      font-size: 18px;
      font-weight: bold;
      color: #082340;
      position: relative;
      margin: 30px 0px 20px;
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

    .task-wrapper {
      display: flex;
      flex-wrap: wrap;

      .task {
        margin-top: -10px;
      }
    }

    .load-more {
      text-align: center;
    }
  }
}
</style>
