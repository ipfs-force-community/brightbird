<template>
  <div class="wrap">
    <template v-for="item in jobList" :key="item?.id">
      <job-item :jobVo="item" @toEdit="toEdit" @toDelete="toDelete"></job-item>
    </template>
    <job-editor
      :id="jobId || ''"
      v-if="editionActivated"
      @closed="editionActivated = false"
      @completed="editCompleted"
    />
  </div>
</template>
<script lang="ts">
import { ref, getCurrentInstance, onMounted, onBeforeMount } from 'vue';
import { listJobs, deleteJob } from '@/api/job';
import { IJobVo } from '@/api/dto/job';
import JobItem from '@/views/job/job-item.vue';
import JobEditor from './job-editor.vue';
import { Mutable } from '@/utils/lib';
import { eventBus } from '@/main';

export default {
  components: { JobItem, JobEditor },
  setup(props: any) {
    const { proxy } = getCurrentInstance() as any;
    const loading = ref<boolean>();
    const jobList = ref<Mutable<IJobVo>[]>([]);
    const jobId = ref<string>();
    const editionActivated = ref<boolean>(false);

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
    const toEdit = val => {
      jobId.value = val.id;
      editionActivated.value = true;
    };

    const toDelete = async (value: { name: string; jobId: string }) => {
      let msg = '<div>确定要删除Job吗?</div>';
      msg += `<div style="margin-top: 5px; font-size: 12px; line-height: normal;">名称：${value.name}</div>`;
      proxy
        .$confirm(msg, '删除Job', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning',
          dangerouslyUseHTMLString: true,
        })
        .then(async () => {
          try {
            await deleteJob(value.jobId);
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

    onBeforeMount(async () => {
      await fetchJobList();
    }),
    onMounted(async () => {
      eventBus.on('newJob', () => {
        fetchJobList();
      });
    });
    const addCompleted = async () => {
      await fetchJobList();
    };
    const editCompleted = async () => {
      await fetchJobList();
    };

    return {
      editionActivated,
      jobList,
      jobId,
      toEdit,
      toDelete,
      editCompleted,
      addCompleted,
    };
  },
};
</script>
<style scoped lang="less">
.wrap {
  margin-top: 10px;
}
</style>
