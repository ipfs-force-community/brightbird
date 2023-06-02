<template>
  <jm-scrollbar>
    <div class="plugin-manager">
      <div class="right-top-btn">
        <router-link :to="{ name: 'index' }">
          <jm-button type="primary" class="jm-icon-button-cancel" size="small">
            关闭
          </jm-button>
        </router-link>
      </div>
      <div>
        <el-upload
          class="upload-demo"
          action="http://192.168.200.171/brightbird"
          :auto-upload="false"
          :on-change="handleFileChange"
        >
          <template #trigger>
              <el-button type="primary">select file</el-button>
          </template>
          <el-button
            style="margin-left: 10px"
            size="primary"
            type="success"
            @click="uploadFiles"
          >upload to server</el-button>
          <div class="el-upload__tip">
              jpg/png files with a size less than 500kb
          </div>
        </el-upload>
      </div>
      <div class="title" style="display: flex; justify-content: space-between;">
        <span>测试插件</span>
        <span class="desc">（共有 {{ execPlugins.total }} 个插件定义）</span>
      </div>

      <div class="content">
        <jm-empty v-if="execPlugins.list.length === 0"/>
        <div
            v-else
            v-for="(i, idx) in execPlugins.list"
            :key="idx"
            class="item"
        >
          <div class="item-t">
            <span class="item-t-t">
              <jm-text-viewer :value="i.name"/>
            </span>

            <p class="item-t-btm">
              <jm-text-viewer :value="`${i.description || '无'}`"/>
            </p>
          </div>

          <div class="item-pos" >
              <button class="delete-label" @click="handleDelete(idx.toString())"></button>
          </div>
        </div>
      </div>

      <div class="title" style="display: flex; justify-content: space-between;">
        <span>部署插件</span>
        <span class="desc">（共有 {{ deployPlugins.total }} 个插件定义）</span>
      </div>

      <div class="content">
        <jm-empty v-if="deployPlugins.list.length === 0"/>
        <div
          v-else
          v-for="(i, idx) in deployPlugins.list"
          :key="idx"
          class="item"
        >
          <div class="item-t">
            <span class="item-t-t">
              <jm-text-viewer :value="i.name"/>
            </span>

            <p class="item-t-btm">
              <jm-text-viewer :value="`${i.description || '无'}`"/>
            </p>
          </div>

          <div class="item-pos" >
              <button class="delete-label" @click="handleDelete(idx.toString())"></button>
          </div>
        </div>
      </div>
    </div>
  </jm-scrollbar>
</template>

<script lang="ts">
import {
  defineComponent,
  getCurrentInstance,
  reactive,
    Ref,
    ref,
} from 'vue';
import { INode } from '@/model/modules/node-library';
import { Mutable } from '@/utils/lib';
import {deletePlugin, fetchDeployPlugins, fetchExecPlugins, uploadPlugin} from "@/api/view-no-auth";
import { PluginDetail } from "@/api/dto/testflow.js";
import { ElButton, ElUpload } from 'element-plus';
import JmEmpty from "@/components/data/empty/index.vue";
import JmTextViewer from "@/components/text-viewer/index.vue";


export default defineComponent({
    components: {JmTextViewer, JmEmpty, ElButton, ElUpload },
    setup() {
        const { proxy } = getCurrentInstance() as any;
        const deployPlugins = reactive<Mutable<INode<PluginDetail>>>({ total: 0, list: [] });
        const execPlugins = reactive<Mutable<INode<PluginDetail>>>({ total: 0, list: [] });

        const fileList: Ref<FileList> = ref(new FileList());

        const handleFileChange = (event: Event) => {
            const inputElement = event.target as HTMLInputElement;
            if (inputElement.files) {
                fileList.value = inputElement.files;
            }
        };


        const uploadFiles = async () => {
            try {
                await uploadPlugin(fileList.value);
            } catch (error) {
            }
        };

        const handleDelete = (index: string) => {
            deletePlugin(index)
                .then(() => {
                    console.log('插件删除成功');
                })
                .catch((error: Error) => {
                    console.error('插件删除失败', error);
                });
        };

        fetchDeployPlugins()
            .then(res => {
                deployPlugins.list = res;
                deployPlugins.total = res.length;
            })
            .catch((err: Error) => {
                proxy.$throw(err, proxy);
            });


        fetchExecPlugins()
            .then(res => {
                execPlugins.list = res;
                execPlugins.total = res.length;
            })
            .catch((err: Error) => {
                proxy.$throw(err, proxy);
            });



        return {
            deployPlugins,
            execPlugins,
            fileList,
            handleFileChange,
            uploadFiles,
            handleDelete,
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

        > p {
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

        & > div {
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
        border-radius: 25.5%;
        overflow: hidden;

        img {
          width: 100%;
          height: 100%;
        }
      }
    }
  }
}
</style>
