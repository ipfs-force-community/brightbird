<template>
    <div ref="chartContainer" style="width: 100%; height: 100%"></div>
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
    deployPlugin: {
      type: Number,
      required: true,
    },
    execPlugin: {
      type: Number,
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
          text: '组件统计',
        },
        tooltip: {
          trigger: 'item',
        },
        legend: {
          show: false,
        },
        series: [
          {
            name: 'Access From',
            type: 'pie',
            radius: '50%',
            data: [
              { value: this.deployPlugin, name: '部署组件' },
              { value: this.execPlugin, name: '测试组件' },
            ],
            emphasis: {
              itemStyle: {
                shadowBlur: 10,
                shadowOffsetX: 0,
                shadowColor: 'rgba(0, 0, 0, 0.5)',
              },
            },
            label: {
              show: true,
              formatter: '{b}：{d}%', // 显示名称和占比
            },
          },
        ],
      };
      this.myChart.setOption(option);
    },
  },
});
</script>
  