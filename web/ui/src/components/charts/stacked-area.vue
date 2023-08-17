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
    testData: {
      type: Object as () => Map<string, number[]>,
      default: () => new Map(), // 默认为空 Map
      required: false,           // 将 required 属性设为 false
    },
    dateArray: {
      type: Array as () => string[],
      default: () => [],         // 默认为空数组
      required: false,           // 将 required 属性设为 false
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
      const seriesData = Object.entries(this.testData).map(([name, data]) => ({
        name,
        type: 'line',
        stack: 'Total',
        areaStyle: {
          color: this.getRandomColor(),
        },
        emphasis: {
          focus: 'series',
        },
        smooth: true,
        lineStyle: {
          color: this.getRandomColor(),
        },
        data: data,
        markLine: { 
          silent: true,
          data: [
            { yAxis: 0 },
          ],
        },
      }));

      const option = {
        title: {
          text: '近2周 / 测试数据',
        },
        tooltip: {
          trigger: 'axis',
          axisPointer: {
            type: 'cross',
            label: {
              backgroundColor: '#6a7985',
            },
          },
        },
        legend: {
          show: false,
        },
        grid: {
          left: '3%',
          right: '4%',
          bottom: '3%',
          containLabel: true,
        },
        xAxis: [
          {
            type: 'category',
            boundaryGap: false,
            data: this.dateArray.length > 0 ? this.dateArray : [''],
          },
        ],
        yAxis: [
          {
            type: 'value',
          },
        ],
        series: seriesData,
      };
      this.myChart.setOption(option);

    },
    getRandomColor() {
      const colors = ['rgb(240, 235, 246)', 'rgb(253, 247, 236)', 'rgb(226, 248, 252)', 'rgb(236, 243, 255)'];
      const randomIndex = Math.floor(Math.random() * colors.length);
      return colors[randomIndex];
    },
  },
});
</script>
