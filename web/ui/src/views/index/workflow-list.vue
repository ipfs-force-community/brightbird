<template>
    <div class="index">
      <div class="main">
        <!-- 全部项目 -->
        <all-project v-if="searchResultFlag" />
      </div>
    </div>
  </template>
  
<script lang="ts">
import { computed, defineComponent } from 'vue';
import AllProject from '@/views/index/all-project.vue';
  
export default defineComponent({
  components: { AllProject },
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
          margin-top: 20px;
          padding: 40px 20px;
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
      }
    }
  </style>
  