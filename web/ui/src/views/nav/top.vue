<template>
  <div class="top-nav">
    <div class="left">
      <router-link to="/">
        <div class="logo" />
      </router-link>
      <div class="version">
        <span class="txt">{{ currentVersion }}</span>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref } from 'vue';
import { useStore } from 'vuex';
import { version as v } from '@/../package.json';
import { IRootState } from '@/model';

export default defineComponent({
  setup() {
    const store = useStore();
    const rootState = store.state as IRootState;
    // 头像图片正常显示
    const loaded = ref<boolean>(true);
    const currentVersion = `v${v}`;

    return {
      loaded,
      currentVersion,
    };
  },
});
</script>

<style scoped lang="less">
.top-nav {
  margin-right: 0.5%;
  position: relative;
  height: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;

  .left {
    display: flex;
    align-items: center;
    position: relative;

    .version {
      position: absolute;
      left: 150px;
      bottom: 2px;
      letter-spacing: normal;
      white-space: nowrap;
      padding: 0 10px 1px;
      height: 16px;
      background-color: #f2f2f2;
      border-radius: 10px;
      display: flex;
      align-items: center;

      .new {
        position: absolute;
        right: -3px;
        top: -3px;
        width: 8px;
        height: 8px;
        background-color: #ff0d0d;
        border-radius: 4px;

        & + .txt {
          cursor: pointer;
        }
      }

      .txt {
        font-size: 12px;
        color: #082340;
      }
    }
  }
}
</style>
