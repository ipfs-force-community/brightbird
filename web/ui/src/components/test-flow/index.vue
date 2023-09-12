<template>
  <div>
    <div class="left">
      <button class="jm-icon-button-left" @click="goBack"></button>
    </div>
  </div>
</template>
<script>
export default {

  setup(props) {
    console.log('=======');
    const goBack = ()=>{
      console.log('============');
    };
    return {
      goBack,
    };
  },
};
</script>
<style lang="less"></style>
