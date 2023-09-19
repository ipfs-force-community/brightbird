<template>
  <div class="pipeline" v-loading="loading">
    <jm-workflow-editor v-model="workflow" @back="close" @save="save" v-if="!loaded" />
  </div>
</template>

<script lang="ts">
import { defineComponent, getCurrentInstance, inject, onMounted, ref, provide } from 'vue';
import { IWorkflow } from '@/components/workflow/workflow-editor/model/data/common';
import { useRoute, useRouter } from 'vue-router';
import { saveTestFlow, fetchTestFlowDetail } from '@/api/view-no-auth';
import yaml from 'yaml';
import JmWorkflowEditor from '@/components/workflow/workflow-editor/index.vue';

export default defineComponent({
  components: { JmWorkflowEditor },
  props: {
    id: {
      type: String,
    },
  },
  setup(props) {
    const { proxy, appContext } = getCurrentInstance() as any;
    const router = useRouter();
    const route = useRoute();
    const loading = ref<boolean>(false);
    // workflow数据是否加载完成
    const loaded = ref<boolean>(false);
    const reloadMain = inject('reloadMain') as () => void;
    const editMode = !!props.id;
    const flowCreateTime = ref<string>('');
    const projectPanelVisible = ref<boolean>(false);
    provide('projectPanelVisible', projectPanelVisible);
    const workflow = ref<IWorkflow>({
      name: '',
      groupId: '',
      createTime: '',
      modifiedTime: '',
      data: '',
    });
    onMounted(async () => {
      if (editMode) {
        try {
          loading.value = true;
          loaded.value = true;
          const fetchedData = await fetchTestFlowDetail({ id: props.id as string });
          const rawData = yaml.parse(fetchedData.graph)['raw-data'];
          flowCreateTime.value = fetchedData.createTime;
          workflow.value = {
            name: fetchedData.name,
            groupId: fetchedData.groupId,
            createTime: fetchedData.createTime,
            modifiedTime: fetchedData.modifiedTime,
            graph: fetchedData.graph,
            data: rawData,
            globalProperties: fetchedData.globalProperties,
          };
        } catch (err) {
          proxy.$throw(err, proxy);
        } finally {
          loading.value = false;
          loaded.value = false;
        }
      }
    });
    const close = async () => {
      if (history.state.back) {
        router.back();
      } else {
        router.push('/');
      }
    };
    const save = async (back: boolean, graph: string) => {
      try {
        if (workflow.value.name === '' || workflow.value.groupId === '') {
          projectPanelVisible.value = true;
          return;
        }

        const id = await saveTestFlow({
          groupId: workflow.value.groupId,
          name: workflow.value.name,
          createTime: (editMode ? BigInt(flowCreateTime.value) : BigInt(Date.now()) * BigInt(1000000)).toString(),
          modifiedTime: (Date.now() * 1000000).toString(),
          graph: graph,
          id: editMode ? props.id : '',
          description: workflow.value.description || '',
          globalProperties: workflow.value.globalProperties,
        });

        if (!back) {
          // 新增项目，再次点击保存进入项目编辑模式
          if (!editMode) {
            await router.push({ name: 'update-pipeline', params: { id } });
            reloadMain();
            return;
          }
          proxy.$success(editMode ? '保存成功' : '新增成功');
          return;
        }
        proxy.$success(editMode ? '保存成功' : '新增成功');
        await close();
      } catch (err) {
        proxy.$throw(err, proxy);
      }
    };

    return {
      loaded,
      loading,
      workflow,
      projectPanelVisible,
      close,
      save,
    };
  },
});
</script>

<style scoped lang="less">
.pipeline {
  position: relative;
  height: 100vh;
}
</style>
