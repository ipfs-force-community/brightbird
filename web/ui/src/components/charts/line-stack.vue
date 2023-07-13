<template>
    <div ref="chartContainer" style="width: 100%; height: 100%;"></div>
  </template>
  
  <script lang="ts">
  import { defineComponent } from 'vue';
  import * as echarts from 'echarts';
  
  export default defineComponent({
    props: {
      chartData: {
        type: Array as () => number[][],
        required: true,
      },
    },
    mounted() {
      this.renderChart();
    },
    methods: {
      renderChart() {
        const chartContainer = this.$refs.chartContainer as HTMLElement;
        const myChart = echarts.init(chartContainer);
        const seriesData = this.chartData.map((data: number[]) => ({
          type: 'line',
          stack: 'Total',
          smooth: true,
          data,
        }));

        const option = {
          title: {
            text: '近 30 天Job成功数量走势'
          },
          tooltip: {
            trigger: 'axis'
          },
          legend: {
            show: false
          },
          grid: {
            left: '3%',
            right: '4%',
            bottom: '3%',
            containLabel: true
          },
          toolbox: {
            feature: {
              saveAsImage: {}
            }
          },
          xAxis: {
            type: 'category',
            boundaryGap: false,
            data: ['Wek1', 'Wek2', 'Wek3', 'Wek4']
          },
          yAxis: {
            type: 'value'
          },
          series: seriesData,
        };
        myChart.setOption(option);
      },
    },
  });
  </script>
  