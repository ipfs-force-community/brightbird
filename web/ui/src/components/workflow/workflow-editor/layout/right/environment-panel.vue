<template >
  <ElDrawer @closed="onClosed" v-model="visible" class="drawer" title="全局变量" destroy-on-close :size="800">
    <ElForm
      :model="globalProperties"
      ref="formRef"
    >
      <div class="drawerWrap">
        <div>
        <div class="title">
          <div></div>
          <div>字段名</div>
          <div>值</div>
        </div>
          <div class="form">
            <template  v-for="(item,index) in globalProperties" :key="index">
              <img :class="index == 0 ? 'disabled': ''" @click="onDelete(index)" src="~@/assets/svgs/icon_delete.svg" alt="">
              <ElFormItem
              :prop="[`${index}`,'name']"
              :rules="rules.name"
              >
                <ElInput size="small" v-model="item.name"></ElInput>
              </ElFormItem>
                <ElFormItem
                :prop="[`${index}`,'value']"
                :rules="rules.value"
                >
                  <ElInput v-model="item.value"></ElInput>
              </ElFormItem>
          </template>    
          </div>
        <div class="add-btn" @click="onAdd">添加参数</div>
      </div>
      <div class="confirm"> <ElButton @click="onConfirm"  type="primary">确认</ElButton></div>
      </div>
    </ElForm>
  </ElDrawer>
</template>
<script lang="ts">

import { GlobalProperty } from '@/api/dto/testflow';
import { ElButton, ElDrawer, ElForm, ElFormItem, ElInput, FormInstance, FormRules } from 'element-plus';
import { PropType, getCurrentInstance, ref, toRefs } from 'vue';
import { IWorkflow } from '../../model/data/common';
import { cloneDeep } from 'lodash';
export default {
  components: { ElInput, ElButton, ElDrawer, ElForm, ElFormItem },
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
    const formRef = ref<FormInstance>();
    const globalProperties = ref<GlobalProperty[]>([]);

    const rules = ref<FormRules<GlobalProperty>>({
      name: [
        { required: true, message: '不能为空', trigger: 'blur' },
      ],
      value: [
        { required: true, message: '不能为空', trigger: 'blur' },
        {
          validator(rule, value, callback) {
            const key = Number(rule.fullField?.replace('.value', '').replace('.', ''));
           // const type = (globalProperties.value)[key].type;       
            // callback(new Error('请输入number 类型'));
            //check duplicate
            return true;
          },
          trigger: 'blur',
        },
      ],
    });

    const visible = ref<boolean>(false);
    visible.value = toRefs(props).modelValue.value;
    globalProperties.value = cloneDeep(props.workflowData.globalProperties) ?? [];

    return {
      formRef,
      globalProperties,
      rules,
      visible,
      onConfirm: () => {
        formRef.value?.validate((valid: boolean) => {
          if (valid) {
            const wd = cloneDeep(props.workflowData);
            wd.globalProperties = globalProperties.value;
            emit('update:workflow-data', wd);
            visible.value = false;
          }
        });
      },
      onDelete: (index: number) => {
        if (index === 0) { return; }
        globalProperties.value.splice(index, 1);
      },
      onAdd: () =>  {
        globalProperties.value?.push({
          name: '',
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
      grid-template-columns: 20px repeat(2,1fr);
      grid-column-gap: 30px;
      align-items: flex-start;
    }

    .form {
      img {
        margin-top: 10px;
        width: 20px;
        height: 20px;
        cursor: pointer;
      }

      img.disabled {
        cursor: not-allowed;
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