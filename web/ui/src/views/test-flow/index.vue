<template>
  <div class="wrap">
    <Toolbar @newGroup="onNewGroup" @newJob="onNewJob" />
    <div class="content">
      <WorkFlowList v-if="!isJob"></WorkFlowList>
      <!-- <JobManager v-if="isJob"></JobManager> -->
      <JobManagerNew v-if="isJob"></JobManagerNew>
    </div>

    <group-creator
      v-if="creationActivated"
      @closed="creationActivated = false"
      @completed="addCompleted"
    />
    <job-creator
        v-if="creationJobActivated"
        @closed="creationJobActivated = false"
        @completed="addJobCompleted"
      />
  </div>
</template>
<script lang="ts">
import { useRoute } from 'vue-router';
import { onBeforeUpdate, onMounted, ref } from 'vue';
import Toolbar from '@/components/test-flow/layout/top/toolbar.vue';
import WorkFlowList from '@/views/index/workflow-list.vue';
import JobManager from '@/views/job/job-manager.vue';
import JobManagerNew from '@/views/job/job-manager-new.vue';
import GroupCreator from '@/views/project-group/project-group-creator.vue';
import JobCreator from '@/views/job/job-creator.vue';

import { eventBus } from '@/main';
export default {
  components: {
    Toolbar,
    WorkFlowList,
    JobManager,
    GroupCreator,
    JobManagerNew,
    JobCreator,
  },

  setup(props: any) {
    let router = useRoute();
    const isJob = ref<boolean>(false);
    const groupName = ref<string>();
    const creationActivated = ref<boolean>(false);
    const creationJobActivated = ref<boolean>(false);
    const groupId = ref<string>();
    const groupDescription = ref<string>();
    const showInHomePage = ref<boolean>(false);

    const addCompleted = () => {
      eventBus.emit('newGroup');
    };

    const addJobCompleted = () => {
      eventBus.emit('newJob');
    };
    const onNewGroup = () => {
      creationActivated.value = true;
    };
    const onNewJob = () => {
      creationJobActivated.value = true;
    };

    onMounted(() => {
      isJob.value = router.hash.includes('#job');
    });
    onBeforeUpdate(() => {
      isJob.value = router.hash.includes('#job');
    });
    return {
      onNewJob,
      onNewGroup,
      addCompleted,
      addJobCompleted,
      isJob,
      groupName,
      creationActivated,
      creationJobActivated,
      groupDescription,
      showInHomePage,
      groupId,
    };
  },
};
</script>
<style scoped lang="less">
.wrap {
  .content {
    overflow: auto;
    padding-top: 64px;
    max-height: 100vh;
  }
}
</style>
