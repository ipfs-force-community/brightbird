<template>
  <div ref="chartContainer" style="width: 100%; height: 100%;"></div>
</template>

<script lang="ts">
import { defineComponent, markRaw } from 'vue';
import * as echarts from 'echarts';

export default defineComponent({
  data() {
    return {
      myChart:{} as echarts.ECharts,
    };
  },
  props: {
    dateArray: {
      type: Array,
      required: true,
    },
    passRateArray: {
      type: Array,
      required: true,
    },
  },
  mounted() {
    this.renderChart();
    window.addEventListener('resize', this.onResize);
  },
  beforeUnmount() {
    window.removeEventListener('resize', this.onResize);
  },
  methods: {
    onResize() {
      this.myChart.resize();
    },
    renderChart() {
      const chartContainer = this.$refs.chartContainer as HTMLElement;
      this.myChart = markRaw(echarts.init(chartContainer));
      const option = {
        title: {
          text: '近 30 天通过率走势',
        },
        xAxis: {
          type: 'category',
          data: this.dateArray,
        },
        yAxis: {
          type: 'value',
        },
        series: [
          {
            data: this.passRateArray,
            type: 'line',
            smooth: true,
          },
        ],
      };
      this.myChart.setOption(option);
    },
  },
});
</script>
