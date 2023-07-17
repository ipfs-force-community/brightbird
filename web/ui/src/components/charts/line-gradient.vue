<template>
  <div ref="chartContainer" style="width: 100%; height: 100%;"></div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import * as echarts from 'echarts';

export default defineComponent({
  props: {
    dateArray: {
      type: Array,
      required: true
    },
    passRateArray: {
      type: Array,
      required: true
    }
  },
  mounted() {
    this.renderChart();
  },
  methods: {
    renderChart() {
      const chartContainer = this.$refs.chartContainer as HTMLElement;
      const myChart = echarts.init(chartContainer);
      const option = {
        title: {
            text: '近 30 天通过率走势'
          },
        xAxis: {
          type: 'category',
          data: this.dateArray
        },
        yAxis: {
          type: 'value'
        },
        series: [
          {
            data: this.passRateArray,
            type: 'line',
            smooth: true
          }
        ]
      };
      myChart.setOption(option);
    },
  },
});
</script>
