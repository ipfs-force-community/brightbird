<template>
  <div class="param-value">
    <div class="value" :style="{maxWidth:maxWidth}">
      <jm-text-viewer @loaded="({contentMaxWidth})=>getMaxWidth(contentMaxWidth)" :value="paramValue"
                      :tip-append-to-body="tipAppendToBody" :tip-placement="tipPlacement" :threshold="0"/>
    </div>
  </div>
</template>

<script lang='ts'>
import { computed, defineComponent, ref } from 'vue';

export default defineComponent({
  name: 'jm-param-value',
  props: {
    value: {
      type: [String, Number, Boolean],
      default: '',
    },
    type: {
      type: String,
    },
    // tooltip 显示方向
    tipPlacement: {
      type: String,
      default: 'bottom-end',
    },
    // 控制tooltip是否被放置到body元素上
    tipAppendToBody: {
      type: Boolean,
      default: true,
    },
  },
  setup(props) {
    const paramValue = computed<string>(() => String(props.value));
    const maxWidth = ref<string>('');
    const getMaxWidth = (width: number) => {
      maxWidth.value += width + 'px';
    };
    const regExp = /((http|https):\/\/([\w\-]+\.)+[\w\-]+(\/[\w\u4e00-\u9fa5\-\.\/?\@\%\!\&=\+\~\:\#\;\,]*)?)/ig;
    const isLink = computed<boolean>(() => {
      if (paramValue.value.startsWith('http://') || paramValue.value.startsWith('https://')) {
        return !!paramValue.value.match(regExp);
      } else {
        return false;
      }
    });
    return {
      paramValue,
      maxWidth,
      getMaxWidth,
      isLink,
    };
  },
});
</script>

<style scoped lang='less'>
.param-value {

  .value {
    display: inline-block;
    width: calc(100% - 26px);
    height: 1.5em;

    a {
      display: inline-block;
      width: 100%;
      color: #096DD9;
    }
  }

  &:hover .copy-btn {
    visibility: visible;
  }
}
</style>
