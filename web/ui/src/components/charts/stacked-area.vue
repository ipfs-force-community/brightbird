<template>
  <div ref="chartContainer" style="width: 100%; height: 100%;"></div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import * as echarts from 'echarts';

export default defineComponent({
  props: {
    testData: {
      type: Object as () => Map<string, number[]>,
      required: true
    },
    dateArray: {
      type: Array as () => string[],
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

      const seriesData = Object.entries(this.testData).map(([name, data]) => ({
        name,
        type: 'line',
        stack: 'Total',
        areaStyle: {
          color: this.getRandomColor()
        },
        emphasis: {
          focus: 'series'
        },
        smooth: true,
        lineStyle: {
          color: this.getRandomColor()
        },
        data: data,
      }));

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
            data: this.dateArray,
          }
        ],
        yAxis: [
          {
            type: 'value'
          }
        ],
        series: seriesData
      };
      myChart.setOption(option);
    },
    getRandomColor() {
      const colors = ['rgb(240, 235, 246)', 'rgb(253, 247, 236)', 'rgb(226, 248, 252)', 'rgb(236, 243, 255)'];
      const randomIndex = Math.floor(Math.random() * colors.length);
      return colors[randomIndex];
    }
  },
});
</script>
