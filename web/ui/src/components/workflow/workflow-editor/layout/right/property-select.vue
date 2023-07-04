<template>
    <el-input v-model="inputValue" :placeholder="property.description ? property.description : '请输入' + property.name"
        show-word-limit :maxlength="50" @focus="inputFocus" @input="showTree = false" @blur="setPropValue" />
    <el-tree v-show="showTree" :data="treeData" :props="defaultProps" expand-on-click-node accordion
        @node-click="handleNodeClick" />
</template>
  
<script lang="ts">
import { defineComponent, ref, PropType } from 'vue';
import { TreeProp } from '@/components/workflow/workflow-editor/model/data/common';
import { Property } from '@/api/dto/testflow';

export default defineComponent({
    emits: [],
    props: {
        name: {
            type: String,
            require: true,
        },
        property: {
            type: Object as PropType<Property>,
            required: true,
        },
        treeData: {
            type: Array as PropType<TreeProp[]>,
            require: true,
        },
        input: {
            type: Object,
            required: true,
        },
    },

    setup(props) {
        const defaultProps = {
            children: 'children',
            label: 'name',
        }

        const inputValue = ref<string>("");
        const showTree = ref<boolean>(false);

        if (props.input[props.property.name]) {
            inputValue.value = props.input[props.property.name];
        }

        const handleNodeClick = function (data: TreeProp, obj: any) {
            var pathSeq = [data.name];
            var parent = obj;
            while (parent.parent && parent.level > 1) {
                parent = parent.parent;
                pathSeq.push(parent.data.name)
            }
            const expressValue = "{{" + pathSeq.reverse().join(".") + "}}";
            inputValue.value = expressValue;
            showTree.value = false;
            setPropValue()
        }

        const inputFocus = () => {
            if (props.treeData?.length && props.treeData?.length > 0) {
                showTree.value = true;
            }
        }

        const setPropValue = () => {
            props.input[props.property.name] = inputValue.value;
        }

        return {
            defaultProps,
            inputValue,
            showTree,
            handleNodeClick,
            inputFocus,
            setPropValue,
        };
    },
});
</script>