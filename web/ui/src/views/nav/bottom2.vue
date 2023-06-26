<template>
  <div class="bottom-nav">
    <a href="https://github.com/ipfs-force-community" target="_blank" class="mulan"
    >IPFS FORCE</a>
    <a href="https://github.com/ipfs-force-community/brightbird" target="_blank">关于brightbird</a>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
  setup() {
    return {
      currentYear: new Date().getFullYear(),
    };
  },
});
</script>

<style scoped lang="less">
.bottom-nav {
  width: 100%;
  height: 64px;
  font-size: 14px;
  font-weight: 400;
  color: #595959;
  text-align: center;
  border-top: 1px solid rgba(185, 207, 230, 0.59);
  box-sizing: border-box;
  display: flex;
  align-items: center;
  justify-content: center;

  > :nth-child(n + 2) {
    margin-left: 20px;
  }

  .line {
    width: 1px;
    height: 14px;
    background: #B9CFE6;
  }

  .mulan {
    line-height: 24px;
    padding-left: 30px;
    background-image: url('@/assets/svgs/logo/mulan.svg');
    background-repeat: no-repeat;
    background-size: 24px 24px;
  }
}
</style>
