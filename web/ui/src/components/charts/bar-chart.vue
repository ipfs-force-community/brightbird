<template>
    <div ref="chartContainer" style="width: 100%; height: 150%;"></div>
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
    jobNames: {
      type: Array,
      default: () => [],
      required: false,
    },
    passRates: {
      type: Array,
      default: () => [],
      required: false,
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
        // 将你提供的图表配置复制到这里
        title: {
          text: '当日通过率排行',
        },
        grid: {
          left: '3%',
          right: '4%',
          bottom: '3%',
          containLabel: true,
        },
        xAxis: {
          type: 'value',
          boundaryGap: [0, 0.01],
          axisLabel: {
            show: true,
            formatter: '{value}%',
          },
        },
        yAxis: {
          type: 'category',
          data: this.jobNames,
          axisLabel: {
            show: true, // 显示y轴标签
          },
        },
        series: [
          {
            name: '',
            type: 'bar',
            data: this.passRates,
            itemStyle: {
              // 设置柱子圆角
              barBorderRadius: [0, 10, 10, 0],
              color: 'rgb(64, 134, 255)',
            },
            label: {
              show: true,
              position: 'outside',
              formatter: '{c}%',
            },
          },
        ],
      };
      this.myChart.setOption(option);
    },
  },
});
</script>
  