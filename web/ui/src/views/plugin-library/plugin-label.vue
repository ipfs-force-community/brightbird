<template>
    <div class="labels">
        <div v-for="p in props?.labels">
            <el-text class="label">{{ p }}
                <el-icon color="red" @click="onDelete(p)">
                    <Delete />
                </el-icon>
            </el-text>
        </div>
        <div>
            <el-text class="add">{{enableAddNewLabel?"":"新增"}}
                <el-icon v-show="!enableAddNewLabel" @click="enableAddNewLabel = true">
                    <Plus />
                </el-icon>
                <el-input class="newlabel" v-model="newLabel" @keyup.enter="addLabel" v-show="enableAddNewLabel" />
            </el-text>
        </div>

    </div>
</template>
  
<script lang="ts">
import { addPluginLabel, delPluginLabel } from '@/api/plugin';
import { defineComponent, ref, getCurrentInstance, PropType } from 'vue';

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

  setup(props, { emit }) {
    const { proxy } = getCurrentInstance() as any;
    const newLabel = ref<string>();
    const enableAddNewLabel = ref<boolean>(false);
    const onDelete = async (label: string) => {
      if (!label || label == '') {
        return;
      }

      enableAddNewLabel.value = true;
      try {
        await delPluginLabel({ name: props.name, label: label });
        props.labels?.splice(props.labels?.indexOf(label), 1);
      } catch (err) {
        proxy.$throw(err, proxy);
      } finally {
        enableAddNewLabel.value = false;
      }
    };
    const addLabel = async () => {
      if (!newLabel.value || newLabel.value == '') {
        return;
      }

      enableAddNewLabel.value = true;
      try {
        await addPluginLabel({ name: props.name, label: newLabel.value });
        props.labels?.push(newLabel.value);
        newLabel.value = '';
      } catch (err) {
        proxy.$throw(err, proxy);
      } finally {
        enableAddNewLabel.value = false;
      }
    };
    return {
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
  