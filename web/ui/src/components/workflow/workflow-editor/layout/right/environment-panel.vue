<template >
    <ElDrawer @closed="onClosed" v-model="visible" class="drawer" title="全局变量" destroy-on-close :size="800">
    <div class="drawerWrap">
      <div>
      <div class="title">
        <div></div>
        <div>字段名</div>
        <div>类型</div>
        <div>值</div>
      </div>
        <div class="form">
          <template  v-for="(item,index) in globalProperties" :key="index">
            <img @click="onDelete(index)" src="~@/assets/svgs/icon_delete.svg" alt="">
            <ElInput v-model="item.name"></ElInput>
            <ElInput v-model="item.type"></ElInput>
            <ElInput v-model="item.value"></ElInput>
        </template>    
        </div>
      <div class="add-btn" @click="onAdd">添加参数</div>
     </div>
     <div class="confirm"> <ElButton @click="onConfirm"  type="primary">确认</ElButton></div>
    </div>
    </ElDrawer>
</template>
<script lang="ts">

import { GlobalProperty } from '@/api/dto/testflow';
import { ElButton, ElDrawer, ElInput } from 'element-plus';
import { PropType, getCurrentInstance, ref, toRefs } from 'vue';
import { IWorkflow } from '../../model/data/common';
import { cloneDeep } from 'lodash';
export default {
  components: { ElInput, ElButton, ElDrawer },
  name: 'environment-panel',
  emits: ['update:workflow-data', 'update:modelValue'],
  props: {
    workflowData: {
      type: Object as PropType<IWorkflow>,
      required: true,
    },
    modelValue: {
      type: Boolean,
      required: true,
    },
  },
  setup(props, { emit }) {
    const proxy = getCurrentInstance()?.proxy as any;
    const globalProperties = ref<GlobalProperty[]>([]);
    const visible = ref<boolean>(false);
    visible.value = toRefs(props).modelValue.value;
    globalProperties.value = cloneDeep(props.workflowData.globalProperties) ?? [];
    return {
      globalProperties,
      visible,
      onConfirm: () => {
        for (const iterator of globalProperties.value) {
          if (iterator.name === '' || iterator.type === '' || iterator.value === '') {
            proxy?.$success('输入框内容不能为空');
            return;
          }
        }
        const wd = cloneDeep(props.workflowData);
        wd.globalProperties = globalProperties.value;
        emit('update:workflow-data', wd);
        visible.value = false;
      },
      onDelete: (index:number )=> {
        globalProperties.value.splice(index, 1);
      },
      onAdd: () =>  {
        globalProperties.value?.push({
          name: '',
          type: '',
          value:'',
        });
      },
      onClosed: ()=>{
        emit('update:modelValue', false);
      },
    };
  },
};
</script>
<style lang="less" scoped>
    .drawerWrap{
      display: flex;
      flex-direction: column;
      justify-content: space-between;
      height: 100%;
    }
    .title {
      font-size: 16px;
      margin-bottom: 20px;
    }
    
    .title ,.form {
      display: grid;
      grid-template-columns: 20px repeat(3,1fr);
      grid-column-gap: 30px;
      grid-row-gap: 20px;
      align-items: center;
    }

    .form {
      img {
        width: 20px;
        height: 20px;
      }
    }
    .add-btn {
      display: flex;
      justify-content: center;
      align-items: center;
      width: 100%;
      border: 1px dashed grey;
      height: 40px;
      cursor: pointer;
      margin-top: 30px;
    }

    .confirm {
      display: flex;
      justify-content: flex-end;
      align-items: center;
      height: 50px;
      margin-top: 30px;

      &>::v-deep.el-button {
        height: 50px;
      }
    }
</style>