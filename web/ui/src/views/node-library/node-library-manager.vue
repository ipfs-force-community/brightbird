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

      <div class="title">
        <span>部署插件</span>
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
          <div class="deprecated" v-if="i.deprecated">
            <jm-tooltip placement="top-start">
              <template #content>
                <div style="line-height: 20px">
                  由于某些原因，该插件不被推荐使用（如该插件可<br/>能会导致一些已知问题或有更好的插件可替代它）
                </div>
              </template>
              <img src="~@/assets/svgs/node-library/deprecated.svg" alt="">
            </jm-tooltip>
          </div>
          <div class="item-t">
            <span class="item-t-t">
              <jm-text-viewer :value="i.name"/>
            </span>

            <p class="item-t-btm">
              <jm-text-viewer :value="`${i.description || '无'}`"/>
            </p>
          </div>

          <div
              class="item-pos"
              :class="{ 'node-definition-default-icon': !i.icon, 'deprecated-icon':i.deprecated}"
          >
            <img
                v-if="i.icon"
                :src="`${i.icon}?imageMogr2/thumbnail/81x/sharpen/1`"
            />
          </div>
        </div>
      </div>

      <div class="title">
        <span>测试插件</span>
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
          <div class="deprecated" v-if="i.deprecated">
            <jm-tooltip placement="top-start">
              <template #content>
                <div style="line-height: 20px">
                  由于某些原因，该插件不被推荐使用（如该插件可<br/>能会导致一些已知问题或有更好的插件可替代它）
                </div>
              </template>
              <img src="~@/assets/svgs/node-library/deprecated.svg" alt="">
            </jm-tooltip>
          </div>
          <div class="item-t">
            <span class="item-t-t">
              <jm-text-viewer :value="i.name"/>
            </span>

            <p class="item-t-btm">
              <jm-text-viewer :value="`${i.description || '无'}`"/>
            </p>
          </div>

          <div
            class="item-pos"
            :class="{ 'node-definition-default-icon': !i.icon, 'deprecated-icon':i.deprecated}"
          >
            <img
              v-if="i.icon"
              :src="`${i.icon}?imageMogr2/thumbnail/81x/sharpen/1`"
            />
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
  ref,
  Ref,
  inject,
} from 'vue';
import { INode } from '@/model/modules/node-library';
import { Mutable } from '@/utils/lib';
import {fetchDeployPlugins, fetchExecPlugins} from "@/api/view-no-auth";
<<<<<<< Updated upstream
<<<<<<< Updated upstream
<<<<<<< Updated upstream
import {INodeVo} from "@/api/dto/project";
=======
=======
>>>>>>> Stashed changes
=======
>>>>>>> Stashed changes
import {INodeVo} from "@/api/dto/node-library";
>>>>>>> Stashed changes

export default defineComponent({
  components: {},
  setup() {
    const { proxy } = getCurrentInstance() as any;
    const deployPlugins = reactive<Mutable<INode<INodeVo>>>({total:0, list:[]});
    const execPlugins = reactive<Mutable<INode<INodeVo>>>({total:0, list:[]});
      fetchDeployPlugins()
        .then(res => {
          deployPlugins.list = res
          deployPlugins.total = res.length
        })
        .catch((err: Error) => {
          proxy.$throw(err, proxy);
        });


        fetchExecPlugins()
        .then(res => {
          execPlugins.list = res
          execPlugins.total = res.length
        })
        .catch((err: Error) => {
          proxy.$throw(err, proxy);
        });

    return {
      deployPlugins,
      execPlugins,
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

      .item-mid {
        background: #f8fcff;
        margin: 0px -15px 0px -15px;
        padding: 6px 0px 6px 15px;
        position: relative;
        height: 62px;

        .down {
          width: 16px;
          height: 16px;
          position: absolute;
          right: 20px;
          top: 10px;
          cursor: pointer;
          z-index: 2;
          background-image: url('@/assets/svgs/nav/top/down.svg');
        }

        .down.direction-down {
          transform: rotate(180deg);
        }

        .item-mid-items {
          display: flex;
          justify-content: flex-start;
          flex-wrap: wrap;
          padding-right: 50px;
          height: 26px;
          overflow: hidden;

          .item-mid-item {
            padding: 5px;
            font-size: 12px;
            color: #385775;
            background: #eff4f8;
            border-radius: 2px;
            margin-right: 10px;
            overflow-y: auto;
            margin-bottom: 10px;
            max-width: 350px;
          }
        }

        .item-mid-items.is-scroll {
          height: auto;
          overflow: auto;
        }
      }

      .item-mid.is-background {
        background: #ffffff;
        min-height: auto;
        overflow: hidden;
        height: 26px;
        margin-bottom: 10px;
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

        .sync {
          background-image: url('@/assets/svgs/btn/sync.svg');

          &.doing {
            animation: rotating 2s linear infinite;

            &:active {
              background-color: transparent;
            }
          }
        }

        .doing {
          opacity: 0.5;
          cursor: not-allowed;
        }

        .del {
          background-image: url('@/assets/svgs/btn/del.svg');
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

        &.deprecated-icon {
          opacity: .4;
        }

        img {
          width: 100%;
          height: 100%;
        }
      }

      .item-pos.node-definition-default-icon {
        background-image: url('@/assets/svgs/node-library/node-definition-default-icon.svg');
        background-size: 100%;
      }
    }
  }

  .load-more {
    text-align: center;
  }
}
</style>
