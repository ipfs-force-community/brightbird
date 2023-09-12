<template>
  <div class="wrap">
    <div class="left">
      <button class="jm-icon-button-left" @click="goBack"></button>
      <div class="tabs">
        <div @click="testClick" :class="{ tab: true, active: active }">
          测试流管理
        </div>
        <div @click="jobClick" :class="{ tab: true, active: !active }">
          Job管理
        </div>
      </div>
    </div>
    <div class="right">
      <ElButton @click="onNewGroup" v-if="active" type="primary">添加分组</ElButton>
      <ElButton @click="onNewJob" v-if="!active" type="primary">新建Job</ElButton>
    </div>
  </div>
</template>
<script lang="ts">
import { ElButton } from 'element-plus';
import { ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';

export default {
  components: { ElButton },
  emits:['newGroup', 'newJob'],
  setup(props, { emit }) {
    const router = useRouter();
    const route = useRoute();
    const active = ref<boolean>(true);
    if (route.hash.includes('job')) {
      active.value = false;
    }
    const goBack = () => {
      router.push('/');
    };
    const testClick = () => {
      active.value = true;
      router.push('#manager');
    };
    const jobClick = () => {
      active.value = false;
      router.push('#job');
    };

    const onNewGroup = ()=>{
      emit('newGroup', '');
    };
    const onNewJob = ()=>{
      emit('newJob', '');
    };

    return {
      testClick,
      jobClick,
      goBack,
      onNewGroup,
      onNewJob,
      active,
    };
  },
};
</script>
<style scoped lang="less">
@primary-color: #096dd9;

.wrap {
  position: fixed;
  z-index: 101;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 30px;
  width: 100%;
  height: 64px;
  background: white;
  color: #042749;
  button[class^="jm-icon-"] {
    width: 24px;
    height: 24px;
    border-width: 0;
    border-radius: 2px;
    background-color: transparent;
    color: #6b7b8d;
    text-align: center;
    font-size: 18px;
    cursor: pointer;

    &::before {
      font-weight: 500;
    }

    &:hover {
      background-color: #eff7ff;
      color: @primary-color;
    }
  }

  .left {
    display: flex;
    .tabs {
      display: flex;
      margin-left: 20px;
      column-gap: 10px;

      .tab {
        color: rgb(83, 83, 83);
        font-size: 16px;
        cursor: pointer;
      }
      .tab.active {
        color: #333333;
        font-weight: 600;
      }
    }
  }

  .right {
    
  }
}
</style>
