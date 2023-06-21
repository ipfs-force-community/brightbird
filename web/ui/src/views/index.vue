<template>
  <div class="index">
    <div class="main">
      <div class="menu-bar">
        <div class="left-area">
          <router-link :to="{ name: 'create-pipeline' }">
            <div class="btn-item">
              <button class="graph"></button>
              <span class="text">图形项目</span>
            </div>
          </router-link>
        </div>
        <div class="right-area">
          <router-link :to="{ name: 'workflow-list' }">
            <div class="btn-item">
              <button class="workflow-list"></button>
              <span class="text">测试流列表</span>
            </div>
          </router-link>
          <router-link :to="{ name: 'plugin-library' }">
            <div class="btn-item">
              <button class="plugin-library"></button>
              <span class="text">本地节点</span>
            </div>
          </router-link>
          <router-link :to="{ name: 'job' }">
            <div class="btn-item">
              <button class="job"></button>
              <span class="text">Job管理</span>
            </div>
          </router-link>
          <router-link :to="{ name: 'project-group' }">
            <div class="btn-item">
              <button class="group"></button>
              <span class="text">分组管理</span>
            </div>
          </router-link>
        </div>
      </div>
      <div class="charts" style="display: flex;">
        <div style="display: flex; flex-direction: column;">
          <div class="icon-container">
            <div style="position: relative;">
              <el-icon :size="25" style="margin-left: 23px; margin-top: 66px; width: 37px; height: 37px; border-radius: 10%; background-color: rgb(255, 228, 186);">
                <EditPen style="color: rgb(255, 130, 10)" />
              </el-icon>
              <div style="position: absolute; left: 23px; top: 30px; width: 109px; height: 19px; color: rgba(16, 16, 16, 1); font-size: 15px; text-align: left;">
                总任务数
              </div>
              <div style="position: absolute; left: 69px; top: 66px; width: 109px; height: 40px; color: rgba(16, 16, 16, 1); font-size: 26px; text-align: left; font-weight: bold;">
                1,902
              </div>
            </div>
            <div style="position: relative;">
              <el-icon :size="25" style="margin-left: 190px; margin-top: 66px; width: 37px; height: 37px; border-radius: 10%; background-color: rgb(232, 255, 251);">
                <CircleCheck style="color: rgb(16, 198, 194)" />
              </el-icon>
              <div style="position: absolute; left: 193px; top: 30px; width: 109px; height: 19px; color: rgba(16, 16, 16, 1); font-size: 15px; text-align: left;">
                通过任务数
              </div>
              <div style="position: absolute; left: 239px; top: 66px; width: 109px; height: 40px; color: rgba(16, 16, 16, 1); font-size: 26px; text-align: left; font-weight: bold;">
                1,601
              </div>
            </div>
            <div style="position: relative;">
              <el-icon :size="25" style="margin-left: 190px; margin-top: 66px; width: 37px; height: 37px; border-radius: 10%; background-color: rgb(232, 243, 255);">
                <CircleClose style="color: rgb(34, 101, 255)" />
              </el-icon>
              <div style="position: absolute; left: 193px; top: 30px; width: 109px; height: 19px; color: rgba(16, 16, 16, 1); font-size: 15px; text-align: left;">
                失败任务数
              </div>
              <div style="position: absolute; left: 239px; top: 66px; width: 109px; height: 40px; color: rgba(16, 16, 16, 1); font-size: 26px; text-align: left; font-weight: bold;">
                207
              </div>
            </div>
            <div style="position: relative;">
            <el-icon :size="25" style="margin-left: 190px; margin-top: 66px; width: 37px; height: 37px; border-radius: 10%; background-color: rgb(245, 232, 255);">
              <PieChart style="color: rgb(129, 67, 214)" />
            </el-icon>
              <div style="position: absolute; left: 193px; top: 30px; width: 109px; height: 19px; color: rgba(16, 16, 16, 1); font-size: 15px; text-align: left;">
                通过率
              </div>
              <div style="position: absolute; left: 239px; top: 66px; width: 109px; height: 40px; color: rgba(16, 16, 16, 1); font-size: 26px; text-align: left; font-weight: bold;">
                75%
              </div>
            </div>
          </div>
          <StackedAreaChart/>
          <div style="display: flex;">
            <LineGradient/>
            <PieSimple/>
          </div>
        </div>
        <div style="display: flex; flex-direction: column;">
          <BarChart/>
          <PieBorderRadius/>
          <LineStack/>
        </div>
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
import StackedAreaChart from '@/components/charts/stacked-area.vue';
import BarChart from '@/components/charts/bar-chart.vue';
import PieBorderRadius from '@/components/charts/pie-border-radius.vue';
import LineGradient from '@/components/charts/line-gradient.vue';
import PieSimple from '@/components/charts/pie-simple.vue';
import LineStack from '@/components/charts/line-stack.vue';
import { Edit, EditPen, CircleCheck, CircleClose, PieChart } from '@element-plus/icons-vue';

export default defineComponent({
  components: { AllProject, SearchProject, BottomNav, StackedAreaChart, BarChart, PieBorderRadius, 
    LineGradient, PieSimple, LineStack, Edit, EditPen, CircleCheck, CircleClose, PieChart },
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
      margin-top: 20px;
      background-color: #ffffff;
      padding: 10px 20px;
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-left: -40px; 
      margin-right: -40px; 

      a {
        margin-right: 60px;

        &:last-child {
          margin-right: 0;
        }
      }

      .btn-item {
        display: flex;
        flex-direction: column;
        align-items: center;

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
          cursor: pointer;
          background-position: center center;
          background-repeat: no-repeat;

          &.graph {
            background-image: url('@/assets/svgs/index/graph-project-btn.svg');
          }

          &.code {
            background-image: url('@/assets/svgs/index/code-project-btn.svg');
          }

          &.git {
            background-image: url('@/assets/svgs/index/git-btn.svg');
          }

          &.plugin-library {
            background-image: url('@/assets/svgs/index/node-library-btn.svg');
          }


          &.job {
            background-image: url('@/assets/svgs/index/code-project-btn.svg');
          }

          &.group {
            background-image: url('@/assets/svgs/index/group-btn.svg');
          }

          &.workflow-list {
            background-image: url('@/assets/svgs/process-template/process-template.svg');
          }
        }

        .text {
          font-size: 14px;
          margin-top: 6px;
          font-weight: 400;
          color: #082340;
        }
      }

      .left-area,
      .right-area {
        display: flex;
      }
    }

    .charts {
      margin-top: 19px;
      background-color: #ffffff;
      padding: 10px;
      display: flex;
      justify-content: flex-start;
      align-items: flex-start;

      .icon-container {
        height: 150px;
        width: 875px;
        display: flex;
      }
    }
  }
}
</style>
