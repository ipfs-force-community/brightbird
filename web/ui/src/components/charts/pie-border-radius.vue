<template>
    <div ref="chartContainer" style="width: 100%; height: 100%;"></div>
  </template>
  
  <script lang="ts">
  import { defineComponent } from 'vue';
  import * as echarts from 'echarts';
  
  export default defineComponent({
    props: {
      map: {
        type: Object as () => Map<string, number>,
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

        const data = Object.entries(this.map).map(([name, value]) => ({ name, value }));

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
                show: true, 
                formatter: '{b}：{d}%'
              },
              emphasis: {
                itemStyle: {
                  shadowBlur: 10,
                  shadowOffsetX: 0,
                  shadowColor: 'rgba(0, 0, 0, 0.5)'
                }
              },
              data
            }
          ]
        };

        myChart.setOption(option);
      },
    },
  });
  </script>
  