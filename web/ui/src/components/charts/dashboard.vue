<template>
    <div>
        <el-row justify="center" style="margin-bottom: 20px;">
            <el-col :span="6">
                <el-row style="margin-bottom: 10px;">
                    <el-col>
                        <div style="color: rgba(16, 16, 16, 1); font-size: 20px; text-align: left;">总任务数</div>
                    </el-col>
                </el-row>
                <el-row style="margin-bottom: 10px;"> 
                    <el-col :span="3">
                        <el-icon :size="25" style="width: 37px; height: 37px; border-radius: 10%; background-color: rgb(255, 228, 186);">
                            <EditPen style="color: rgb(255, 130, 10)" />
                        </el-icon>
                    </el-col>
                    <el-col :span="3"> 
                        <div style="width: 109px; height: 40px; color: rgba(16, 16, 16, 1); font-size: 26px; text-align: left; font-weight: bold;">1,902</div>
                    </el-col>
                </el-row>
            </el-col>
            <el-col :span="6">
                <el-row style="margin-bottom: 10px;">
                    <el-col>
                        <div style="color: rgba(16, 16, 16, 1); font-size: 20px; text-align: left;">通过任务数</div>
                    </el-col>
                </el-row>
                <el-row style="margin-bottom: 10px;"> 
                    <el-col :span="3">
                        <el-icon :size="25" style="width: 37px; height: 37px; border-radius: 10%; background-color: rgb(232, 255, 251);">
                            <CircleCheck style="color: rgb(16, 198, 194)" />
                        </el-icon>
                    </el-col>
                    <el-col :span="3"> 
                        <div style="width: 109px; height: 40px; color: rgba(16, 16, 16, 1); font-size: 26px; text-align: left; font-weight: bold;">1,601</div>
                    </el-col>
                </el-row>
            </el-col>
            <el-col :span="6">
                <el-row style="margin-bottom: 10px;">
                    <el-col>
                        <div style="color: rgba(16, 16, 16, 1); font-size: 20px; text-align: left;">失败任务数</div>
                    </el-col>
                </el-row>
                <el-row style="margin-bottom: 10px;"> 
                    <el-col :span="3">
                        <el-icon :size="25" style="width: 37px; height: 37px; border-radius: 10%; background-color: rgb(232, 243, 255);">
                            <CircleClose style="color: rgb(34, 101, 255)" />
                        </el-icon>
                    </el-col>
                    <el-col :span="3"> 
                        <div style="width: 109px; height: 40px; color: rgba(16, 16, 16, 1); font-size: 26px; text-align: left; font-weight: bold;">207</div>
                    </el-col>
                </el-row>
            </el-col>
            <el-col :span="6">
                <el-row style="margin-bottom: 10px;">
                    <el-col>
                        <div style="color: rgba(16, 16, 16, 1); font-size: 20px; text-align: left;">通过率</div>
                    </el-col>
                </el-row>
                <el-row style="margin-bottom: 10px;"> 
                    <el-col :span="3">
                        <el-icon :size="25" style="width: 37px; height: 37px; border-radius: 10%; background-color: rgb(245, 232, 255);">
                            <PieChart style="color: rgb(129, 67, 214)" />
                        </el-icon>
                    </el-col>
                    <el-col :span="3"> 
                        <div style="width: 109px; height: 40px; color: rgba(16, 16, 16, 1); font-size: 26px; text-align: left; font-weight: bold;">75%</div>
                    </el-col>
                </el-row>
            </el-col>
        </el-row>
        <el-row >
            <el-col :span="12" style="height: 500px;"><StackedAreaChart/></el-col>
            <el-col :span="6"><BarChart/></el-col>
            <el-col :span="6"><PieBorderRadius/></el-col>
        </el-row>
        <el-row>
            <el-col :span="8" style="height: 400px;"><LineGradient/></el-col>
            <el-col :span="8" style="height: 400px;"><PieSimple/></el-col>
            <el-col :span="8" style="height: 400px;"><LineStack/></el-col>
        </el-row>
    </div>
  </template>
  
  <script lang="ts">
  import { defineComponent } from 'vue';
  import StackedAreaChart from '@/components/charts/stacked-area.vue';
  import LineGradient from '@/components/charts/line-gradient.vue';
  import PieSimple from '@/components/charts/pie-simple.vue';
  import BarChart from '@/components/charts/bar-chart.vue';
  import LineStack from '@/components/charts/line-stack.vue';
  import PieBorderRadius from '@/components/charts/pie-border-radius.vue';
  
  export default defineComponent({
    components: { StackedAreaChart, LineGradient, PieSimple, BarChart, LineStack, PieBorderRadius },
    name: 'DashBoard',
  });
  </script>
  
  <style>
    .el-row {
        margin-bottom: 100px;
    }
    .el-row:last-child {
        margin-bottom: 0;
    }
    .el-col {
        border-radius: 4px;
    }
    
    .grid-content {
        border-radius: 4px;
        min-height: px;
    }
  </style>
  