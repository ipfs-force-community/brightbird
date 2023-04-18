<template>
  <div class="pipeline" v-loading="loading">
    <jm-workflow-editor v-model="workflow" @back="close" @save="save" v-if="!loaded" />
  </div>
</template>

<script lang="ts">
import { defineComponent, getCurrentInstance, inject, nextTick, onMounted, ref, provide } from 'vue';
import { IWorkflow } from '@/components/workflow/workflow-editor/model/data/common';
import { useRoute, useRouter } from 'vue-router';
import { saveTestFlow, fetchTestFlowDetail } from '@/api/view-no-auth';
import { createNamespacedHelpers, useStore } from 'vuex';
import { Case, Node} from "@/api/dto/testflow.js";
import yaml from 'yaml';

export default defineComponent({
  props: {
    id: {
      type: String,
    },
  },
  setup(props) {
    const { proxy, appContext } = getCurrentInstance() as any;
    const router = useRouter();
    const route = useRoute();
    const store = useStore();
    const { payload } = route.params;
    const loading = ref<boolean>(false);
    // workflow数据是否加载完成
    const loaded = ref<boolean>(false);
    const reloadMain = inject('reloadMain') as () => void;
    const editMode = !!props.id;
    const flowCreateTime = ref<string>('');
    const projectPanelVisible = ref<boolean>(false);
    provide('projectPanelVisible', projectPanelVisible);
    const workflow = ref<IWorkflow>({
      name: '未命名项目',
      groupId: '1',
      createTime: '',
      modifiedTime: '',
      data: '',
    });
    onMounted(async () => {
      // if (payload && editMode) {
      //   // 初始化走这里获取到的cache为空
      //   workflow.value = JSON.parse(payload as string);
      //   loaded.value = true;
      //   await nextTick();
      //   loaded.value = false;
      //   return;
      // }
      if (editMode) {
        try {
          loading.value = true;
          loaded.value = true;
          const fetchedData = await fetchTestFlowDetail({id:props.id as string, name:""});
          const rawData = yaml.parse(fetchedData.graph)['raw-data'];
          flowCreateTime.value = fetchedData.createTime
          workflow.value = {
            name: fetchedData.name,
            groupId: fetchedData.groupId,
            createTime: fetchedData.createTime,
            modifiedTime: fetchedData.modifiedTime,
            cases: fetchedData.cases,
            nodes: fetchedData.nodes,
            graph: fetchedData.graph,
            data: rawData,
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
      await router.push({ name: 'index' });
    };
    return {
      loaded,
      loading,
      workflow,
      projectPanelVisible,
      close,
      save: async (back: boolean, caseList: Node[], nodeList: Node[], graph: string) => {
        try {
          if (workflow.value.name === '未命名项目' || workflow.value.groupId === '1') {
            projectPanelVisible.value = true;
            return;
          }

          const  id = await saveTestFlow({
            groupId: workflow.value.groupId,
            name: workflow.value.name,
            createTime: (editMode ? BigInt(flowCreateTime.value) : BigInt(Date.now()) * BigInt(1000000)).toString(),
            modifiedTime: (Date.now() * 1000000).toString(),
            cases: caseList,
            nodes: nodeList,
            graph: graph,
            id: editMode ? props.id : '',
            description: workflow.value.description || '',
          });
          proxy.$success(editMode ? '保存成功' : '新增成功');
          if (!back) {
            // 新增项目，再次点击保存进入项目编辑模式
            if (!editMode) {
              await router.push({ name: 'update-pipeline', params: { id } });
              reloadMain();
              return;
            }
            return;
          }
          await close();
        } catch (err) {
          proxy.$throw(err, proxy);
        }
      },
    };
  },
});
</script>

<style scoped lang="less">
.pipeline {
  height: 100vh;
  position: relative;
}
</style>
