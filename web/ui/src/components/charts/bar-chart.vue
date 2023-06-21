<template>
    <div ref="chartContainer" style="margin-left: 50px; margin-top: 45px; width: 503px; height: 228px;"></div>
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
          // 将你提供的图表配置复制到这里
          title: {
            text: '当日通过率排行'
          },
          tooltip: {
            trigger: 'axis',
            axisPointer: {
              type: 'shadow'
            }
          },
          legend: {},
          grid: {
            left: '3%',
            right: '4%',
            bottom: '3%',
            containLabel: true
          },
          xAxis: {
            type: 'value',
            boundaryGap: [0, 0.01],
            axisLabel: {
              formatter: '{value}%'
            }
          },
          yAxis: {
            type: 'category',
            data: ['Job1', 'Job2', 'Job3']
          },
          series: [
            {
              name: '',
              type: 'bar',
              data: [50, 60, 80], // 修改为百分比数据
              itemStyle: {
                // 设置柱子圆角
                barBorderRadius: [0, 10, 10, 0],
                color: 'rgb(64, 134, 255)'
              },
              barWidth: 12,
              label: {
                show: true,
                position: 'inside',
                formatter: '{c}%'
              }
            },
          ]
        };
        myChart.setOption(option);
      },
    },
  });
  </script>
  