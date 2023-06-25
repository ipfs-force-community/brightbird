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
          text: '近2周 / 测试数据'
        },
        tooltip: {
          trigger: 'axis',
          axisPointer: {
            type: 'cross',
            label: {
              backgroundColor: '#6a7985'
            }
          }
        },
        legend: {
          show: false
        },
        toolbox: {
          feature: {
            saveAsImage: {}
          }
        },
        grid: {
          left: '3%',
          right: '4%',
          bottom: '3%',
          containLabel: true
        },
        xAxis: [
          {
            type: 'category',
            boundaryGap: false,
            data: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']
          }
        ],
        yAxis: [
          {
            type: 'value'
          }
        ],
        series: [
          {
            name: 'Job1',
            type: 'line',
            stack: 'Total',
            areaStyle: {
              color: 'rgb(240, 235, 246)'
            },
            emphasis: {
              focus: 'series'
            },
            smooth: true,
            lineStyle: {
              color: 'rgb(130, 65, 196)'
            },
            data: [120, 132, 101, 134, 90, 230, 210]
          },
          {
            name: 'Job2',
            type: 'line',
            stack: 'Total',
            areaStyle: {
              color: 'rgb(253, 247, 236)'
            },
            emphasis: {
              focus: 'series'
            },
            smooth: true,
            lineStyle: {
              color: 'rgb(255, 125, 0)'
            },
            data: [220, 182, 191, 234, 290, 330, 310]
          },
          {
            name: 'Job3',
            type: 'line',
            stack: 'Total',
            areaStyle: {
              color: 'rgb(226, 248, 252)'
            },
            emphasis: {
              focus: 'series'
            },
            smooth: true,
            lineStyle: {
              color: 'rgb(59, 226, 226)'
            },
            data: [150, 232, 201, 154, 190, 330, 410]
          },
          {
            name: 'Job4',
            type: 'line',
            stack: 'Total',
            areaStyle: {
              color: 'rgb(236, 243, 255)'
            },
            emphasis: {
              focus: 'series'
            },
            smooth: true,
            lineStyle: {
              color: 'rgb(56, 108, 255)'
            },
            data: [320, 332, 301, 334, 390, 330, 320]
          },
        ]
      };
      myChart.setOption(option);
    },
  },
});
</script>
