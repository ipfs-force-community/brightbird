<template>
  <div class="group" v-if="isShow">
    <div class="fold-switch" @click="toggle">
      <i :class="['jm-icon-button-right',collapsed?'':'rotate']"></i>
      <span class="group-name">{{ groupName }}</span>
      <div class="loading" v-loading="loading"></div>
    </div>
    <div class="nodes" v-show="!collapsed">
      <!-- 网络异常 -->
      <div class="network-anomaly" v-if="networkAnomaly">
        <div class="anomaly-img"></div>
        <span class="anomaly-tip">网络开小差啦</span>
        <div class="reload-btn" @click="loadNodes(keyword,true)">重新加载</div>
      </div>
      <template v-else-if="nodesCache.length>0">
        <div class="nodes-wrapper">
          <x6-vue-shape
            v-for="item in nodesCache"
            :key="item.getName()"
            :node-data="item"
            @mouseup="(e)=> onMouseUpEvent(item,e)"
            @mousedown="(e) => onMouseDownEvent(item,e)"
            @mousemove="(e)=>onMouseMoveEvent(item,e)"
            />
        </div>
        <!-- 显示更多 -->
        <div class="load-more">
          <jm-load-more :state="loadState" :load-more="btnDown"></jm-load-more>
        </div>
      </template>
    </div>
  </div>
</template>
<script lang="ts">
import { computed, defineComponent, inject, onMounted, onUpdated, PropType, ref } from 'vue';
import X6VueShape from '../../shape/x6-vue-shape.vue';
import { IWorkflowNode } from '../../model/data/common';
import { WorkflowNode } from '../../model/workflow-node';
import { TimeoutError } from '@/utils/rest/error';
import { StateEnum } from '@/components/load-more/enumeration';
import { WorkflowDnd } from '../../model/workflow-dnd';
import { PluginTypeEnum } from '@/api/dto/enumeration';

export default defineComponent({
  emits: ['getNodeCount', 'onNodeClick'],
  props: {
    type: {
      type: String as PropType<PluginTypeEnum>,
      required: true,
    },
    keyword: {
      type: String,
      required: true,
    },

    tags:{
      type:Array<string>,
      required: true,
    },
  },
  components: {
    X6VueShape,
  },

  watch:{
    nodes:{
      handler(val) {
        this.nodesCache = this.nodes;
      },
      deep:true,
    },
    tags:{
      handler(val:string[]) {
        if (val.length > 0) {
          this.nodesCache = this.nodes.filter(value=>{
            for (const item of val) {
              if ( value.getLabels().includes(item)) {
                return true;
              }
            }
            return false;
          });
          return;
        } 
        this.nodesCache = this.nodes;
      },

    },
  },
  setup(props, { emit }) {
    const groupName = computed<string>(() => {
      if (props.type === PluginTypeEnum.Deploy) {
        return '部署节点';
      } else if (props.type === PluginTypeEnum.Exec) {
        return '测试节点';
      } else {
        return '';
      }
    });
    // 组件初始加载的节点个数
    const initialPageSize = 12;
    const searchPageSize = 10000;
    // 显示更多组件状态
    const loadState = ref<StateEnum>(StateEnum.NONE);
    // 控制整个分组的显隐
    const isShow = ref<boolean>(true);
    const loading = ref<boolean>(false);
    const collapsed = ref<boolean>(true);
    const nodes = ref<IWorkflowNode[]>([]);
    const nodesCache = ref<IWorkflowNode[]>([]);
    const keyWord = ref<string>(props.keyword);
    const mousePosition = ref();
    mousePosition.value = {
      'clientX': 0,
      'clientY':0,
    };

    const getWorkflowDnd = inject('getWorkflowDnd') as () => WorkflowDnd;
    const workflowNode = new WorkflowNode();
    // 当前为第几页，默认第一页
    const currentPage = ref<number>(1);
    // 网络请求超时
    const networkAnomaly = ref<boolean>(false);
    /**
     * @param keyword 搜索关键字
     * @param reload 网络异常点击重试
     * @param currentPage 节点数据分页，当前第几页
     * @param pageSize 一页加载几条数据
     */
    const loadNodes = async (keyword: string, reload: boolean) => {
      // 点击重试，将请求超时的状态还原到初始状态
      if (reload) {
        networkAnomaly.value = false;
      }
      try {
        loading.value = true;
        switch (props.type) {
          // 官方节点
          case PluginTypeEnum.Deploy: {
            const { content } = await workflowNode.loadDeployPlugins(keyword);
            nodes.value = content;
            break;
          }
          // 社区节点
          case PluginTypeEnum.Exec: {
            const { content } = await workflowNode.loadExecPlugins(keyword);
            // 显示更多
            // totalPages <= currentPage ? (loadState.value = StateEnum.NONE) : (loadState.value = StateEnum.MORE);
            // currentPage > 1 ? nodes.value = [...nodes.value, ...content] : nodes.value = content;
            nodes.value = content;
            break;
          }
        }
        // 传递当前分组已加载的节点数量
        emit('getNodeCount', nodes.value.length);
        if (nodes.value.length <= 0) {
          isShow.value = false;
          return;
        }
        // 加载的节点数量，大于0分组展开
        collapsed.value = false;
        isShow.value = true;
      } catch (err) {
        // 网络超时
        if (err instanceof TimeoutError) {
          networkAnomaly.value = true;
          // 网络超时折叠分组
          collapsed.value = true;
        }
      } finally {
        loading.value = false;
      }
    };
    // 切换分组的折叠状态
    const toggle = async () => {
      collapsed.value = !collapsed.value;
    };
    // 显示更多
    // eslint-disable-next-line @typescript-eslint/no-empty-function
    const btnDown = async () => {};
    onMounted(async () => {
      await loadNodes(keyWord.value, false);
    });
    onUpdated(async () => {
      if (keyWord.value === props.keyword) {
        return;
      }
      keyWord.value = props.keyword;
      // 搜索节点
      await loadNodes(keyWord.value, false);
      // 搜索后将currentPage初始化
      currentPage.value = 1;
    });

    const onMouseDownEvent = (item: IWorkflowNode, e: MouseEvent) => {
      mousePosition.value['clientX'] = e.clientX;
      mousePosition.value['clientY'] = e.clientY;
      getWorkflowDnd().callback(() => {
        mousePosition.value['clientX'] = 0;
        mousePosition.value['clientY'] = 0;
      });
    };
    const onMouseUpEvent = (item: IWorkflowNode, e: MouseEvent) => {
      console.log(item, e);
      if (mousePosition.value['clientX'] === e.clientX && mousePosition.value['clientY'] === e.clientY) {
        emit('onNodeClick', item);
      }
      mousePosition.value['clientX'] = 0;
      mousePosition.value['clientY'] = 0;
    };

    const onMouseMoveEvent = (item: IWorkflowNode, e: MouseEvent) => { 
      if (mousePosition.value['clientX'] !== 0 && mousePosition.value['clientY'] !== 0) {
        mousePosition.value['clientX'] = 0;
        mousePosition.value['clientX'] = 0;
        getWorkflowDnd().drag(item, e);
      }
    };

    return {
      onMouseMoveEvent,
      onMouseUpEvent,
      onMouseDownEvent,
      mousePosition,
      initialPageSize,
      isShow,
      groupName,
      loadState,
      networkAnomaly,
      loading,
      nodes,
      nodesCache,
      collapsed,
      toggle,
      btnDown,
      drag: (data: IWorkflowNode, event: MouseEvent) => {
        getWorkflowDnd().drag(data, event);
      },
      loadNodes,
    };
  },
});
</script>
<style lang="less" scoped>
@import '../../vars';

.group {
  display: flex;
  flex-direction: column;
  margin: 0 20px 24px;
  position: relative;

  &::after {
    position: absolute;
    bottom: -24px;
    content: '';
    display: inline-block;
    height: 1px;
    width: 100%;
    background-color: #EBEEFB;
  }

  &.trigger, &.inner {
    .nodes {
      .nodes-wrapper {
        margin-bottom: -20px;

        .jm-workflow-x6-vue-shape {
          margin-bottom: 0;
        }
      }
    }
  }

  .fold-switch {
    margin-top: 24px;
    cursor: pointer;
    display: flex;
    align-items: center;
    font-size: 14px;

    .jm-icon-button-right {
      width: 16px;
      height: 16px;
      line-height: 16px;
      transition: all .1s linear;

      &::before {
        font-size: 12px;
        color: #6B7B8D;
        transform: scale(0.8);
      }

      &.rotate {
        transform: rotate(90deg);
      }
    }

    .group-name {
      font-weight: 550;
      margin-left: 5px;
    }

    .loading {
      ::v-deep(.el-loading-spinner) {
        top: 0;
        left: 5px;
        margin-top: -3px;
      }

      width: 14px;
      height: 14px;
    }

    font-weight: normal;
    color: @title-color;
  }

  .nodes {
    margin: 24px 0 0 0;

    .network-anomaly {
      margin-top: 10px;
      width: 100%;
      display: flex;
      flex-direction: column;
      align-items: center;
      font-weight: 400;

      .anomaly-img {
        width: 56px;
        height: 56px;
        background-image: url("../../svgs/network.svg");
      }

      .anomaly-tip {
        margin: 20px 0 24px;
        font-size: 14px;
        color: #8095A9;
      }

      .reload-btn {
        cursor: pointer;
        font-size: 16px;
        color: @primary-color;
        margin-bottom: 10px;
      }
    }

    .nodes-wrapper {
      margin-left: 15px;

      .jm-workflow-x6-vue-shape:nth-child(3n) {
        margin-right: 0;
      }

      // 单独控制最后一行三个节点间距太大的问题
      .jm-workflow-x6-vue-shape:nth-last-child(1) {
        margin-bottom: 0;
      }

      .jm-workflow-x6-vue-shape:nth-last-child(2) {
        margin-bottom: 0;
      }

      .jm-workflow-x6-vue-shape:nth-last-child(3) {
        margin-bottom: 0;
      }

      display: flex;
      flex-wrap: wrap;
    }

    .load-more {
      display: flex;
      justify-content: center;

      ::v-deep(.jm-load-more) {
        margin-top: 16px;

        button {
          padding: 0;
          min-height: 20px;

          .icon {
            top: 7px;
            right: -15px;
            border-width: 5px;
          }
        }
      }
    }

    ::v-deep(.jm-workflow-x6-vue-shape) {
      margin: 0 38px 16px 0;
      width: 64px;

      .icon {
        width: 64px;
        height: 64px;
      }
    }
  }
}
</style>
