<template>
    <el-input v-if="props.propName == 'script'" type="textarea" v-model="refValue"
        :placeholder="property.description ? property.description : '请输入' + property.name" @focus="inputFocus"
        @input="showTree = false" @blur="setPropValue" />

    <el-input v-else v-model="refValue"
        :placeholder="property.description ? property.description : '请输入' + property.name" @focus="inputFocus"
        @input="showTree = false" @blur="setPropValue" />

    <el-tree v-show="showTree" :data="treeData" :props="defaultProps" :load="loadNode" lazy>
        <template #default="{ node, data }">
            <span class="custom-tree-node">
                <span @click="handleNodeClick(data, node)">{{node.label }}</span>
                <el-input-number v-show="data.name == 'index'" class="arrayIndex" v-model="data.index" size="small" />
            </span>
        </template>
    </el-tree>
</template>
  
<script lang="ts">
import { defineComponent, ref, PropType } from 'vue';
import { TreeProp } from '@/components/workflow/workflow-editor/model/data/common';
import type Node from 'element-plus/es/components/tree/src/model/node';
import { JSONSchema } from 'json-schema-to-typescript';
import { JSONSchema4Object, JSONSchema4Array } from 'json-schema';

export default defineComponent({
  emits: [],
  props: {
    instanceName: {
      type: String,
      require: true,
      nullable: false,
    },
    propName: {
      type: String,
      require: true,
      nullable: false,
    },
    property: {
      type: Object as PropType<JSONSchema>,
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
    if (!props.propName) {
      return;
    }

    const defaultProps = {
      children: 'children',
      label: 'name',
      isLeaf: (data, node) => {
        return data.isLeaf;
      },
    };

    const refValue = ref<string>('');
    const showTree = ref<boolean>(false);

    if (props.input[props.propName]) {
      // eslint-disable-next-line vue/no-setup-props-destructure
      refValue.value = props.input[props.propName];
    } else {
      if (props.property.default) {
        // todo check object and arrary default value
        refValue.value = props.property.default as string;
        // eslint-disable-next-line vue/no-mutating-props
        props.input[props.propName] = refValue.value;
      }
    }

    const handleNodeClick = function (data: TreeProp, obj: any) {
      var pathSeq:string[] = [];
      var parent = obj;
      while (parent.parent && parent.level > 0) {
        let onePath = parent.data.name;
        if (parent.data.name === 'index') {
          onePath = parent.data.index;
        }
        pathSeq.push(onePath);
        parent = parent.parent;
      }
      const expressValue = '{{' + pathSeq.reverse().join('.') + '}}';
      refValue.value = expressValue;
      showTree.value = false;
      // eslint-disable-next-line no-use-before-define
      setPropValue();
    };

    const inputFocus = () => {
      showTree.value = true;
    };

    const setPropValue = () => {
      if (!props.propName) {
        return;
      }
      // TODO: 临时不强制必选项
      // if (!refValue.value) {
      //   return;
      // }
      // eslint-disable-next-line vue/no-mutating-props
      props.input[props.propName] = refValue.value;
    };


    const loadNode = (node: Node, resolve: (data: TreeProp[]) => void) => {
      if (node.level === 0) {
        resolve(props.treeData ?? []);
        return;
      }

      const defs = node.data.defs;
      let schema = node.data.schema as JSONSchema;

      const resolveSchema = (schemaWithChild: JSONSchema): JSONSchema => {
        if (schemaWithChild['$ref']) {
          const refType = schemaWithChild['$ref'].replace('#/definitions/', '');
          return defs[refType];  // "#/$defs/TestparamssendEmbedStruct"
        }
        return schemaWithChild;
      };

      const isSimpleType = (type: string): boolean => {
        switch (type) {
          case 'string':
          case 'number':
          case 'integer':
          case 'boolean':
            return true;
          default:
            return false;
        }
      };

      if (node.data.children.length > 0) {
        resolve(node.data.children);
        return;
      }

      let treeData: TreeProp[] = [];
      if (schema.type === 'object') {
        for (let [key, prop] of Object.entries(schema.properties)) {
          let treeProp: TreeProp = {
            name: key,
            index: 0,
            defs: defs,
            isLeaf: false,
            type: '',
            schema: null,
            children: [],
          };          

          prop = resolveSchema(prop);

          if (isSimpleType(prop.type as string)) {
            treeProp.type = prop.type as string;
            treeProp.schema = resolveSchema(prop);
            treeProp.isLeaf = true;
            treeData.push(treeProp);
            continue;
          }

          if (prop.type instanceof Array) {
            treeProp.type = 'array';
            treeProp.index = 0;
            treeProp.schema = resolveSchema(prop);
            treeProp.children = [{
              name: 'index',
              index: 0,
              defs: defs,
              isLeaf: false,
              type: treeProp.schema.type as string,
              schema: treeProp.schema,
              children: [],
            }];
            if (isSimpleType(treeProp.schema.type as string)) {
              treeProp.children[0].isLeaf = true;
            }
            treeData.push(treeProp);
            continue;
          }

          // object
          treeProp.type = 'object';
          treeProp.schema = prop;
          treeData.push(treeProp);
          continue;
        }
      } else {
        throw new TypeError('unexpect json type' + schema.type);
      }
      resolve(treeData);
    };
    return {
      props,
      defaultProps,
      refValue,
      showTree,
      loadNode,
      handleNodeClick,
      inputFocus,
      setPropValue,
      treeData: props.treeData,
    };
  },
});
</script>


<style lang="less" scoped>
.arrayIndex {
    height: 20px;
    width: 80px;
    margin-left: 20px;
}
</style>
