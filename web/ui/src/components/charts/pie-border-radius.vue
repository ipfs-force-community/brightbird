<template>
    <div ref="chartContainer" style="width: 100%; height: 100%;"></div>
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
            text: '任务失败占比',
            top: '30px',
            textStyle: {
              fontSize: 18,
              fontWeight: 'bold'
            }
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
              radius: ['50%', '70%'],
              avoidLabelOverlap: false,
              itemStyle: {
                borderRadius: 10,
                borderColor: '#fff',
                borderWidth: 2
              },
              label: {
                show: true, // 显示标签文本
                formatter: '{b}：{d}%' // 标签文本格式，{b}代表name，{d}%代表占比
              },
              emphasis: {
                itemStyle: {
                  shadowBlur: 10,
                  shadowOffsetX: 0,
                  shadowColor: 'rgba(0, 0, 0, 0.5)'
                }
              },
              data: [
                { value: 15, name: 'Job1' },
                { value: 65, name: 'Job2' },
                { value: 2, name: 'Job3' },
                { value: 42, name: 'Job4' }
              ]
            }
          ]
        };

        myChart.setOption(option);
      },
    },
  });
  </script>
  