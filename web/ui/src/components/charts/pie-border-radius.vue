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
    map: {
      type: Object as () => Map<string, number>,
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
      const data = Object.entries(this.map).map(([name, value]) => ({ name, value }));
      const option = {
        title: {
          text: '任务失败占比',
          top: '30px',
          textStyle: {
            fontSize: 18,
            fontWeight: 'bold',
          },
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
            radius: ['50%', '70%'],
            avoidLabelOverlap: false,
            itemStyle: {
              borderRadius: 10,
              borderColor: '#fff',
              borderWidth: 2,
            },
            label: {
              show: true, 
              formatter: '{b}：{d}%',
            },
            emphasis: {
              itemStyle: {
                shadowBlur: 10,
                shadowOffsetX: 0,
                shadowColor: 'rgba(0, 0, 0, 0.5)',
              },
            },
            data,
          },
        ],
      };
      this.myChart.setOption(option);
    },
  },
});
</script>
  