<template>
    <div class="labels">
        <div v-for="p,index in props?.labels" :key="index">
            <el-text class="label">{{ p }}
                <el-icon color="red" @click="onDelete(p)">
                    <Delete />
                </el-icon>
            </el-text>
        </div>
        <div>
            <el-text @click="enableAddNewLabel = true" class="add">{{enableAddNewLabel?"":"新增"}}
                <el-icon v-show="!enableAddNewLabel">
                    <Plus />
                </el-icon>
                <!-- <el-input class="newlabel" v-model="newLabel" @keyup.enter="addLabel" v-show="enableAddNewLabel" /> -->
                <ElSelect @blur="addLabel" allow-create filterable v-model="selectLabels" class="tag-select newlabel" size="small" v-show="enableAddNewLabel">
                  <ElOption 
                  v-for="(item,index) in _labels"
                  :key="index" 
                  :label="item" 
                  :value="item"
                  />
                </ElSelect>
            </el-text>
        </div>

    </div>
</template>
  
<script lang="ts">
import { addPluginLabel, delPluginLabel } from '@/api/plugin';
import { defineComponent, ref, getCurrentInstance, PropType } from 'vue';
import { mapActions, mapState, useStore } from 'vuex';

export default defineComponent({
  emits: [],
  props: {
    name: {
      type: String,
      required: true,
    },
    labels: {
      type: Array as PropType<string[]>,
      require: true,
    },
  },
  computed:{
    ...mapState('worker-editor', {
      _labels:state=>{
        return state.labels;
      },
    }),
  },
  methods:{
    ...mapActions('worker-editor', [
      'getLabels',
    ]),
  },
  setup(props, { emit }) {
    const { proxy } = getCurrentInstance() as any;
    const newLabel = ref<string>();
    const enableAddNewLabel = ref<boolean>(false);
    const selectLabels = ref<string>('');
    const store = useStore();

    const onDelete = async (label: string) => {
      if (!label || label === '') {
        return;
      }

      enableAddNewLabel.value = true;
      try {
        await delPluginLabel({ name: props.name, label: label });
        // eslint-disable-next-line vue/no-mutating-props
        props.labels?.splice(props.labels?.indexOf(label), 1);
      } catch (err) {
        proxy.$throw(err, proxy);
      } finally {
        enableAddNewLabel.value = false;
      }
    };
    const addLabel = async () => {
      // if (!newLabel.value || newLabel.value === '') {
      //   return;
      // }
      if (selectLabels.value.length === 0) {
        return;
      }

      enableAddNewLabel.value = true;
      try {
        await addPluginLabel({ name: props.name, label: selectLabels.value });
        // eslint-disable-next-line vue/no-mutating-props
        props.labels?.push(selectLabels.value);
        // newLabel.value = '';
        selectLabels.value = '';
        // ts-ignore
        store.dispatch('worker-editor/getLabels');
      } catch (err) {
        proxy.$throw(err, proxy);
      } finally {
        enableAddNewLabel.value = false;
      }
    };
    return {
      selectLabels,
      enableAddNewLabel,
      newLabel,
      addLabel,
      onDelete,
      props,
    };
  },
});
</script>
  
<style scoped lang="less">
.labels {
    display: flex;

    .label {
        margin-right: 10px;
        border: 1px solid small;
        border-radius: 5px;
        background-color: azure;
    }

    .add {
        cursor: pointer;
        margin-right: 10px;
        border: 1px solid small;
        border-radius: 5px;
        background-color: azure;
        font-weight: bold;
    }
    .newlabel {
        width: 50%;
    }
}
</style>
  