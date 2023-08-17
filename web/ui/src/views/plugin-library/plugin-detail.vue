<template>
  <div class="plugin-detail">
    <div class="right-top-btn">
      <jm-button type="primary" class="jm-icon-button-cancel" size="small" @click="close">关闭</jm-button>
    </div>
    <div class="content" v-loading="loadingTop">
      <div class="top-title">
        <el-text class="mx-1 name">{{ pluginDetail?.name }}</el-text>
      </div>

      <el-row class="row-t">
        <el-col :span="2">
          <el-text class="mx-1">插件类型:</el-text>
        </el-col>
        <el-col :span="22">
          <el-text class="mx-1 name">{{ pluginDetail?.pluginType == PluginTypeEnum.Deploy ? "部署" : "测试" }}</el-text>
        </el-col>
      </el-row>

      <el-row class="row-t">
        <el-col :span="2">
          <el-text class="mx-1">标签:</el-text>
        </el-col>
        <el-col :span="22">
          <plugin-label :name="pluginDetail?.name ?? ''" :labels="pluginDetail?.labels" />
        </el-col>
      </el-row>

      <el-row class="row-t">
        <el-col :span="2">
          <el-text class="mx-1">版本:</el-text>
        </el-col>
        <el-col :span="22">
          <div class="version">
            <div v-for="p in pluginDetail?.pluginDefs">
              <el-text class="v-item">{{ p.version }}
                <el-icon color="red" @click="deleteVersion(pluginDetail?.id, p.version)">
                  <Delete />
                </el-icon>
              </el-text>
            </div>
          </div>
        </el-col>
      </el-row>

      <el-tabs type="card">
        <el-tab-pane v-for="p in pluginDetail?.pluginDefs" :label="p.version">
          <el-row class="row-t">
            <el-col :span="2">
              <el-text class="mx-1">描述:</el-text>
            </el-col>
            <el-col :span="22">
              <el-text v-html="(pluginDetail?.description || '无').replace(/\n/g, '<br/>')" />
            </el-col>
          </el-row>

          <el-row class="row-t">
            <el-col :span="2">
              <el-text class="mx-1">编译脚本:</el-text>
            </el-col>
            <el-col :span="22">
              <textarea class="script" v-model="p.buildScript" />
            </el-col>
          </el-row>

          <el-row>
            <el-col :span="12">
              <el-card shadow="never">
                <template #header>
                  <el-text>输入参数</el-text>
                </template>
                <vue-json-pretty :data="p.inputSchema" />
              </el-card>
            </el-col>
            <el-col :span="12">
              <el-card shadow="never">
                <template #header>
                  <el-text>输出参数</el-text>
                </template>
                <vue-json-pretty :data="p.outputSchema" />
              </el-card>
            </el-col>
          </el-row>
        </el-tab-pane>
      </el-tabs>
    </div>

  </div>
</template>

<script lang="ts">
import { PluginDetail } from '@/api/dto/node-definitions';
import { getPluginByName, deletePlugin } from '@/api/plugin';
import { defineComponent, getCurrentInstance, inject, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useStore } from 'vuex';
import { IRootState } from '@/model';
import { PluginTypeEnum } from '@/api/dto/enumeration';
import PluginLabel from './plugin-label.vue';
import VueJsonPretty from 'vue-json-pretty';

export default defineComponent({
  props: {
    name: {
      type: String,
      required: true,
    },
  },
  components: {
    PluginLabel, VueJsonPretty
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
        pluginDetail.value = await getPluginByName(props.name);
      } catch (err) {
        proxy.$throw(err, proxy);
      } finally {
        loadingTop.value = false;
      }
    };

    const deleteVersion = async (id: string | undefined, version: string) => {
      try {
        if (pluginDetail.value) {
          await deletePlugin(id ?? '', version);
          pluginDetail.value.pluginDefs = pluginDetail.value?.pluginDefs?.filter(a => a.version != version);
        }
      } catch (err) {
        proxy.$throw(err, proxy);
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
      deleteVersion,
    };
  },
});
</script>

<style scoped lang="less">
.plugin-detail {
  margin-bottom: 20px;

  .right-top-btn {
    position: fixed;
    right: 20px;
    top: 78px;

    .jm-icon-button-cancel::before {
      font-weight: bold;
    }
  }

  .content {
    min-height: 58px;
    font-size: 24px;
    padding: 24px;
    background-color: #ffffff;

    .line {
      display: block;
    }

    .script {
      height: 100px;
      width: 100%;
      font-size: 15px;
    }

    .row-t {
      margin: 10px;
    }

    .top-title {
      display: flex;
      align-items: center;
      color: #082340;

      .name {
        font-size: 32px;
        font-weight: 500;
      }
    }

    .version {
      display: flex;

      .v-item {
        background-color: antiquewhite;
        margin-right: 10px;
        border: 1px solid small;
        border-radius: 5px;
      }
    }

    .description {
      margin-top: 10px;
      color: #6b7b8d;
    }
  }
}
</style>
