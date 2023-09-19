<template>
  <div class="index">
    <div class="main">
      <div class="menu-bar">
        <div class="left-area">
          <div>BrightBird</div>
        </div>
        <div class="right-area">
          <ElButton plain>
            <router-link
              :to="{ path: 'test-flow', query: {}, hash: '#job' }"
            >
              测试流管理
            </router-link>
          </ElButton>
          <ElButton plain>
            <router-link :to="{ name: 'create-pipeline' }">
              开启工作
            </router-link>
          </ElButton>
        </div>
      </div>
      <div class="charts">
        <DashBoard />
      </div>
    </div>
    <bottom-nav />
  </div>
</template>

<script lang="ts">
import { computed, defineComponent } from 'vue';
import BottomNav from '@/views/nav/bottom2.vue';
import AllProject from '@/views/index/all-project.vue';
import SearchProject from '@/views/index/search-project.vue';
import DashBoard from '@/components/charts/dashboard.vue';
import { ElButton } from 'element-plus';
export default defineComponent({
  // eslint-disable-next-line vue/no-unused-components
  components: { AllProject, SearchProject, BottomNav, DashBoard, ElButton },
  props: {
    searchName: {
      type: String,
    },
    groupId: {
      type: String,
    },
  },
  setup(props) {
    // 切换到搜索结果页
    return {
      searchResultFlag: computed<boolean>(() => props.searchName === undefined),
    };
  },
});
</script>

<style scoped lang="less">
.index {
  .main {
    min-height: calc(100vh - 130px);

    .menu-bar {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 10px 20px;
      background-color: #ffffff;

      a {
        margin-right: 60px;

        &:last-child {
          margin-right: 0;
        }
      }

      .btn-item {
        display: flex;
        align-items: center;
        flex-direction: column;

        &:hover {
          .text {
            color: #096dd9;
          }
        }

        button {
          width: 56px;
          height: 56px;
          border: none;
          background-color: #ffffff;
          background-position: center center;
          background-repeat: no-repeat;
          cursor: pointer;

          &.graph {
            background-image: url("@/assets/svgs/index/graph-project-btn.svg");
          }

          &.code {
            background-image: url("@/assets/svgs/index/code-project-btn.svg");
          }

          &.git {
            background-image: url("@/assets/svgs/index/git-btn.svg");
          }

          &.plugin-library {
            background-image: url("@/assets/svgs/index/node-library-btn.svg");
          }

          &.job {
            background-image: url("@/assets/svgs/index/code-project-btn.svg");
          }

          &.group {
            background-image: url("@/assets/svgs/index/group-btn.svg");
          }

          &.workflow-list {
            background-image: url("@/assets/svgs/process-template/process-template.svg");
          }
        }

        .text {
          margin-top: 6px;
          color: #082340;
          font-weight: 400;
          font-size: 14px;
        }
      }

      .left-area,
      .right-area {
        display: flex;
      }
    }

    .charts {
      margin-top: 19px;
      padding: 10px;
      background-color: #ffffff;
    }
  }
}
</style>
