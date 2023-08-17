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
    testData: {
      type: Object as () => Map<string, number[]>,
      required: true,
    },
    dateArray: {
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
      const seriesData = Object.entries(this.testData).map(([name, data]) => ({
        name,
        type: 'line',
        stack: 'Total',
        smooth: true,
        data,
      }));

      const option = {
        title: {
          text: '近 30 天Job成功数量走势',
        },
        tooltip: {
          trigger: 'axis',
        },
        legend: {
          show: false,
        },
        grid: {
          left: '3%',
          right: '4%',
          bottom: '3%',
          containLabel: true,
        },
        toolbox: {
          feature: {
            saveAsImage: {},
          },
        },
        xAxis: {
          type: 'category',
          boundaryGap: false,
          data: this.dateArray,
        },
        yAxis: {
          type: 'value',
        },
        series: seriesData,
      };
      this.myChart.setOption(option);
    },
  },
});
</script>
  