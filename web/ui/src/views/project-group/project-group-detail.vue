<template>
  <div class="project-group-detail">
    <div class="right-top-btn">
      <jm-button type="primary" class="jm-icon-button-cancel" size="small" @click="close">关闭</jm-button>
    </div>
    <div class="top-card" v-loading="loadingTop">
      <div class="top-title">
        <div class="name">{{ testflowGroup?.name }}</div>
        <div class="count">
          （共有 {{ testflowGroup?.testFlowCount }} 个测试流）
        </div>
      </div>
      <div class="description">
        <span v-html="(testflowGroup?.description || '无').replace(/\n/g, '<br/>')"/>
      </div>
    </div>
    <div class="content">
      <div class="menu-bar">
        <button class="add" @click="add">
          <div class="label">添加测试流</div>
        </button>
      </div>
      <div class="title">
        <div>
          <span>测试流列表</span>
        </div>
      </div>
      <div class="group-list-wrapper">
        <project-group
          v-if="initialized"
          :is-detail="true"
          :project-group="testflowGroup"
          :pageable="true"
        />
      </div>
    </div>
    <project-adder
      :id="id"
      v-if="creationActivated"
      @closed="creationActivated = false"
      @completed="addCompleted"
    />
  </div>
</template>

<script lang="ts">
import { IProjectGroupVo } from '@/api/dto/project-group';
import {getProjectGroup, queryTestFlow} from '@/api/view-no-auth';
import { defineComponent, getCurrentInstance, inject, onMounted, ref } from 'vue';
import ProjectAdder from '@/views/project-group/project-adder.vue';
import ProjectGroup from '@/views/common/project-group.vue';
import { useRouter } from 'vue-router';
import { useStore } from 'vuex';
import { IRootState } from '@/model';
import {IPageVo} from "@/api/dto/common";
import {IProjectVo} from "@/api/dto/project";

export default defineComponent({
  props: {
    id: {
      type: String,
      required: true,
    },
  },
  components: {
    ProjectAdder,
    ProjectGroup,
  },
  setup(props) {
    const { proxy } = getCurrentInstance() as any;
    const router = useRouter();
    const store = useStore();
    const rootState = store.state as IRootState;
    const initialized = ref<boolean>(false);
    const loadingTop = ref<boolean>(false);
    const isShow = ref<boolean>(true);
    const creationActivated = ref<boolean>(false);
    const testflowGroup = ref<IProjectGroupVo>();
    const reloadMain = inject('reloadMain') as () => void;
    const add = () => {
      creationActivated.value = true;
    };
    const fetchProjectGroupDetail = async () => {
      try {
        loadingTop.value = true;
        testflowGroup.value = await getProjectGroup(props.id)
        initialized.value = true;
      } catch (err) {
        proxy.$throw(err, proxy);
      } finally {
        loadingTop.value = false;
      }
    };
    onMounted(async () => {
      await fetchProjectGroupDetail();
    });
    const addCompleted = async () => {
      reloadMain();
    };
    return {
      initialized,
      isShow,
      loadingTop,
      creationActivated,
      close: () => {
        if (!['/', '/project-group'].includes(rootState.fromRoute.path)) {
          router.push({ name: 'index' });
          return;
        }
        router.push(rootState.fromRoute.fullPath);
      },
      add,
      addCompleted,
      testflowGroup,
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
    font-size: 14px;
    padding: 24px;
    background-color: #ffffff;

    .top-title {
      display: flex;
      align-items: center;
      color: #082340;

      .name {
        font-size: 20px;
        font-weight: 500;
      }

      .count {
        font-weight: 400;
        opacity: 0.45;
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
