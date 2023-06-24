<template>
    <div ref="chartContainer" style="width: 100%; height: 100%"></div>
  </template>
  
  <script lang="ts">
  import { defineComponent } from 'vue';
  import * as echarts from 'echarts';
  
  export default defineComponent({
    mounted() {
      this.renderChart();
    },
    methods: {
      renderChart() {
        const chartContainer = this.$refs.chartContainer as HTMLElement;
        const myChart = echarts.init(chartContainer);
        const option = {
          title: {
            text: '组件统计'
          },
          tooltip: {
            trigger: 'item'
          },
          legend: {
            show: false
          },
          series: [
            {
              name: 'Access From',
              type: 'pie',
              radius: '50%',
              data: [
                { value: 15, name: '部署组件' },
                { value: 30, name: '测试组件' },
              ],
              emphasis: {
                itemStyle: {
                  shadowBlur: 10,
                  shadowOffsetX: 0,
                  shadowColor: 'rgba(0, 0, 0, 0.5)'
                }
              },
              label: {
                show: true,
                formatter: '{b}：{d}%' // 显示名称和占比
              }
            }
          ]
        };
        myChart.setOption(option);
      },
    },
  });
  </script>
  