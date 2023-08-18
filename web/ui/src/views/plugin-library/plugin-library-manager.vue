<template>
  <jm-scrollbar>
    <router-view v-if="childRoute"></router-view>
    <div class="plugin-manager" v-else>
      <div class="right-top-btn">
        <router-link :to="{ name: 'index' }">
          <jm-button type="primary" class="jm-icon-button-cancel" size="small">
            关闭
          </jm-button>
        </router-link>
      </div>

      <div class="upload">
        <el-upload v-model:file-list="fileList" ref="unloadRef" :auto-upload="false" multiple>
          <template #trigger>
            <el-button type="primary" style="margin-right: 10px; box-shadow: none;">选择文件</el-button>
          </template>
          <el-button 
            style="margin-right: 10px; 
            margin-left: 10px;" 
            type="success" 
            :loading="uploading"
            @click="submitUpload()">上传</el-button>
          </el-upload>
        <el-button style="margin-left: 10px;" @click="openTemplateUrl">查看模版文件</el-button>
      </div>

      <div class="title" style="display: flex; justify-content: space-between;">
        <span>测试插件</span>
        <span class="desc">（共有 {{ execPlugins.total }} 个插件定义）</span>
      </div>

      <div class="content">
        <jm-empty v-if="execPlugins.list.length === 0" />
        <div v-else v-for="(i, idx) in execPlugins.list" :key="idx" class="item">
          <router-link :to="{ path: `/plugin/detail/${i.name}` }">
            <div class="item-t">
              <span class="item-t-t">
                <jm-text-viewer :value="i.name" />
              </span>

              <p class="item-t-btm">
                <jm-text-viewer :value="`${i.description || '无'}`" />
              </p>
            </div>
          </router-link>

          <div class="item-pos">
            <el-button 
              type="danger" 
              :icon="Delete" 
              :loading="i.isDeleting"
              @click="handleDelete(i)"
              class="small-delete-button"></el-button>
          </div>
        </div>
      </div>

      <div class="title" style="display: flex; justify-content: space-between;">
        <span>部署插件</span>
        <span class="desc">（共有 {{ deployPlugins.total }} 个插件定义）</span>
      </div>

      <div class="content">
        <jm-empty v-if="deployPlugins.list.length === 0" />
        <div v-else v-for="(i, idx) in deployPlugins.list" :key="idx" class="item">
          <router-link :to="{ path: `/plugin/detail/${i.name}` }">
          <div class="item-t">
            <span class="item-t-t">
              <jm-text-viewer :value="i.name" />
            </span>

            <p class="item-t-btm">
              <jm-text-viewer :value="`${i.description || '无'}`" />
            </p>
          </div>
          </router-link>
          <div class="item-pos">
            <el-button 
              type="danger" 
              :icon="Delete" 
              :loading="i.isDeleting"
              @click="handleDelete(i)"
            class="small-delete-button"></el-button>
          </div>
        </div>
      </div>
    </div>
  </jm-scrollbar>
</template>

<script lang="ts">
import {
  defineComponent,
  ref,
  getCurrentInstance,
  Ref,
} from 'vue';
import { INode } from '@/model/modules/plugin-library';
import { Mutable } from '@/utils/lib';
import { fetchDeployPlugins, fetchExecPlugins, uploadPlugin, deletePluginAllVersion } from '@/api/plugin';
import { PluginDetail } from '@/api/dto/node-definitions.js';
import { ElButton, ElUpload } from 'element-plus';
import JmEmpty from '@/components/data/empty/index.vue';
import JmTextViewer from '@/components/text-viewer/index.vue';
import { ElMessageBox } from 'element-plus';
import { Delete } from '@element-plus/icons-vue';
import { downloadPublicZip } from '@/api/plugin';
import { PluginTypeEnum } from '@/api/dto/enumeration';
import type { UploadProps, UploadUserFile } from 'element-plus'

import {
  onBeforeRouteUpdate,
  RouteLocationNormalized,
  RouteLocationNormalizedLoaded,
  useRoute,
} from 'vue-router';

export default defineComponent({
  components: { JmTextViewer, JmEmpty, ElButton, ElUpload },
  setup() {
    const detailActive = ref<boolean>(false);
    const { proxy } = getCurrentInstance() as any;
    const deployPlugins = ref<Mutable<INode<PluginDetail>>>({ total: 0, list: [] });
    const execPlugins = ref<Mutable<INode<PluginDetail>>>({ total: 0, list: [] });

    const fileList: Ref<UploadUserFile[]> = ref([]);
    const downloadUrl = `/public.zip`;


    const uploading = ref(false); 

    fetchDeployPlugins()
    .then(res => {
      deployPlugins.value.list = res || [];
      deployPlugins.value.total = res ? res.length : 0;
    })
    .catch((err: Error) => {
      proxy.$throw(err, proxy);
    });

    fetchExecPlugins()
    .then(res => {
      execPlugins.value.list = res || [];
      execPlugins.value.total = res ? res.length : 0;
    })
    .catch((err: Error) => {
      proxy.$throw(err, proxy);
    });

    function changeView(
      childRoute: Ref<boolean>,
      route: RouteLocationNormalizedLoaded | RouteLocationNormalized,
    ) {
      childRoute.value = route.matched.length > 2;
    }
    const childRoute = ref<boolean>(false);
    changeView(childRoute, useRoute());
    onBeforeRouteUpdate(to => changeView(childRoute, to));

    const openTemplateUrl = () => {
      window.open('https://github.com/ipfs-force-community/brightbird/templates', '_blank');
    };

    const handleDelete = async (plugin: PluginDetail) => {
      try {
        plugin.isDeleting = true;

        await deletePluginAllVersion(plugin.id);

        if (plugin.pluginType === PluginTypeEnum.Deploy) {
          deployPlugins.value.list = deployPlugins.value.list.filter(item => item.id !== plugin.id);
          deployPlugins.value.total = deployPlugins.value.list.length;
        } else if (plugin.pluginType === PluginTypeEnum.Exec) {
          execPlugins.value.list = execPlugins.value.list.filter(item => item.id !== plugin.id);
          execPlugins.value.total = execPlugins.value.list.length;
        }

      } catch (error) {
        console.error('delete plugin failed:', error);
      } finally {
        plugin.isDeleting = false;
      }
    };

    const submitUpload = async () => {
      try {
        uploading.value = true;
        if (fileList.value.length > 0) {
          const formData = new FormData(); // 创建新的FormData对象

          fileList.value.forEach((file, index) => {
            formData.append('plugins', file.raw); // 将每个文件添加到FormData对象中
          });

          await uploadPlugin(formData);
          
          // 获取最新数据
          fetchExecPlugins()
            .then(res => {
              execPlugins.value.list = res;
              execPlugins.value.total = res.length;
            });

          fetchDeployPlugins()  
            .then(res => {
              deployPlugins.value.list = res;
              deployPlugins.value.total = res.length;  
            });

          ElMessageBox.alert('上传成功', '提示', {
            confirmButtonText: '确定',
            type: 'success',
          });
        } else {
          ElMessageBox.alert('文件列表为空，请先选择文件', '提示', {
            confirmButtonText: '确定',
            type: 'warning',
          });
        }
      } catch (error) {
        ElMessageBox.alert(`上传失败: ${error}`, '错误', {
          confirmButtonText: '确定',
          type: 'error',
        });
      } finally {
        uploading.value = false;
      }
    };

    return {
      Delete,
      childRoute,
      detailActive,
      deployPlugins,
      execPlugins,
      fileList,
      uploading,
      handleDelete,
      submitUpload,
      openTemplateUrl,
    };
  },
});
</script>

<style scoped lang="less">
.plugin-manager {
  padding: 16px 20px 25px 16px;
  background-color: #ffffff;

  // height: calc(100vh - 185px);
  .right-top-btn {
    position: fixed;
    right: 20px;
    top: 78px;

    .jm-icon-button-cancel::before {
      font-weight: bold;
    }
  }

  .upload {
    display: flex;
  }
  .title {
    font-size: 20px;
    font-weight: bold;
    color: #082340;
    position: relative;
    padding-left: 20px;
    padding-right: 20px;
    margin: 30px -20px 20px;

    .desc {
      font-weight: normal;
      margin-left: 12px;
      font-size: 14px;
      color: #082340;
      opacity: 0.46;
    }
  }

  .content {
    display: flex;
    flex-wrap: wrap;
    min-height: 200px;

    .item {
      margin: 0.5%;
      width: 24%;
      min-width: 230px;
      background-color: #ffffff;
      box-shadow: 0 0 12px 4px #edf1f8;
      padding: 15px;
      position: relative;
      box-sizing: border-box;

      .deprecated {
        position: absolute;
        top: 0;
        right: 0;

        img {
          width: 45px;
          height: 45px;
        }
      }

      .item-t {
        display: flex;
        flex-direction: column;
        color: #082340;
        max-width: 75%;

        >p {
          margin-bottom: 10px;
        }

        .item-t-t {
          color: #082340;
          text-decoration: none;
          margin-bottom: 10px;
          font-size: 16px;
        }

        a.item-t-t:hover {
          color: #096dd9;
        }

        .item-t-mid {
          font-size: 14px;
        }

        .item-t-btm {
          font-size: 14px;
          color: #385775;
        }
      }

      .item-btm {
        display: flex;
        justify-content: space-between;

        &>div {
          width: 49%;
          display: flex;
          align-items: end;
        }

        button {
          width: 26px;
          height: 26px;
          background-color: transparent;
          border: 0;
          background-position: center center;
          background-repeat: no-repeat;
          margin-right: 16px;
          cursor: pointer;

          &:active {
            background-color: #eff7ff;
            border-radius: 4px;
          }
        }

        .item-btm-r {
          justify-content: end;
          color: #7c91a5;
          font-size: 14px;

          ::v-deep(.jm-text-viewer) {
            width: 100%;

            .content {
              .text-line {
                &:last-child {
                  text-align: right;

                  &::after {
                    display: none;
                  }
                }
              }
            }
          }
        }
      }

      .item-pos {
        position: absolute;
        top: 20px;
        right: 20px;
        width: 54px;
        height: 54px;

        .small-delete-button {
          width: 40px;
          padding: 0;
          font-size: 15px;
        }
      }
    }
  }
}
</style>
