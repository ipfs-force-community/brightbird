<template>
  <div class="project-group-detail">
    <div class="right-top-btn">
      <jm-button type="primary" class="jm-icon-button-cancel" size="small" @click="close">关闭</jm-button>
    </div>
    <div class="top-card" v-loading="loadingTop">
      <div class="top-title">
        <el-text class="mx-1 name">{{ pluginDetail?.name }}</el-text>
      </div>

      <el-row>
        <el-col :span="2">
          <el-text class="mx-1">插件类型:</el-text>
        </el-col>
        <el-col :span="22">
          <el-text class="mx-1 name">{{ pluginDetail?.pluginType == PluginTypeEnum.Deploy ? "部署" : "测试" }}</el-text>
        </el-col>
      </el-row>

      <el-row>
        <el-col :span="2">
          <el-text class="mx-1">标签:</el-text>
        </el-col>
        <el-col :span="22">
          <plugin-label :name="pluginDetail?.name??''" :labels="pluginDetail?.labels"/>
        </el-col>
      </el-row>

      <el-row>
        <el-col :span="2">
          <el-text class="mx-1">描述:</el-text>
        </el-col>
        <el-col :span="22">
          <el-text v-html="(pluginDetail?.description || '无').replace(/\n/g, '<br/>')" />
        </el-col>
      </el-row>

    </div>
    <div class="content">

    </div>

  </div>
</template>

<script lang="ts">
import { PluginDetail } from '@/api/dto/testflow';
import { getPluginByName } from '@/api/plugin';
import { defineComponent, getCurrentInstance, inject, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useStore } from 'vuex';
import { IRootState } from '@/model'
import { PluginTypeEnum } from '@/api/dto/enumeration';
import PluginLabel from './plugin-label.vue';

export default defineComponent({
  props: {
    name: {
      type: String,
      required: true,
    },
  },
  components: {
    PluginLabel,
  },
  setup(props) {
    const { proxy } = getCurrentInstance() as any;
    const router = useRouter();
    const store = useStore();
    const rootState = store.state as IRootState;
    const loadingTop = ref<boolean>(false);

    const pluginDetail = ref<PluginDetail>();
    const fetchProjectGroupDetail = async () => {
      try {
        loadingTop.value = true;
        pluginDetail.value = await getPluginByName(props.name)
      } catch (err) {
        proxy.$throw(err, proxy);
      } finally {
        loadingTop.value = false;
      }
    };
    onMounted(async () => {
      await fetchProjectGroupDetail();
    });
    return {
      loadingTop,
      close: () => {
        if (!['/', '/plugin'].includes(rootState.fromRoute.path)) {
          router.push({ name: 'index' });
          return;
        }
        router.push(rootState.fromRoute.fullPath);
      },

      pluginDetail,
      PluginTypeEnum,
    };
  },
});
</script>

<style scoped lang="less">
.project-group-detail {
  margin-bottom: 20px;

  .right-top-btn {
    position: fixed;
    right: 20px;
    top: 78px;

    .jm-icon-button-cancel::before {
      font-weight: bold;
    }
  }

  .top-card {
    min-height: 58px;
    font-size: 24px;
    padding: 24px;
    background-color: #ffffff;

    .top-title {
      display: flex;
      align-items: center;
      color: #082340;

      .name {
        font-size: 32px;
        font-weight: 500;
      }
    }

    .description {
      margin-top: 10px;
      color: #6b7b8d;
    }
  }

  .content {
    margin-top: 20px;
    padding: 15px 15px 0px;
    background-color: #ffffff;

    .menu-bar {
      button {
        position: relative;

        .label {
          position: absolute;
          left: 0;
          bottom: 40px;
          width: 100%;
          text-align: center;
          font-size: 18px;
          color: #b5bdc6;
        }

        &.add {
          // margin: 0.5%;
          width: 19%;
          min-width: 260px;
          height: 170px;
          background-color: #ffffff;
          border: 1px dashed #b5bdc6;
          background-image: url('@/assets/svgs/btn/add.svg');
          background-position: center 45px;
          background-repeat: no-repeat;
          cursor: pointer;
        }
      }
    }

    .title {
      font-size: 18px;
      font-weight: bold;
      color: #082340;
      position: relative;
      margin: 30px 0px 20px;
      display: flex;
      justify-content: space-between;
      align-items: center;

      .move {
        cursor: pointer;
        width: 24px;
        height: 24px;
        background-image: url('@/assets/svgs/sort/move.svg');
        background-size: contain;

        &.active {
          background-image: url('@/assets/svgs/sort/move-active.svg');
        }
      }

      .desc {
        font-weight: normal;
        margin-left: 12px;
        font-size: 14px;
        color: #082340;
        opacity: 0.46;
      }
    }

    .group-list-wrapper {
      display: flex;
      flex-direction: column;

      .load-more {
        align-self: center;
      }

      .project-group {
        margin-top: -10px;
      }
    }
  }
}
</style>
