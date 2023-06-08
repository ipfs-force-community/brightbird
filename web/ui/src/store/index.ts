import { ActionContext, createStore } from 'vuex';
import { IRootState, IScrollOffset } from '@/model';
import { fetchVersion } from '@/api/view-no-auth';
import { RouteLocationNormalized } from 'vue-router';

const store = createStore<IRootState>({
  // 开发环境开启严格模式，在严格模式下，无论何时发生了状态变更且不是由 mutation 函数引起的，将会抛出错误
  strict: process.env.NODE_ENV !== 'production',
  // 根状态
  state: {
    version: "",
    thirdPartyType: '',
    authMode: 'readonly',
    parameterTypes: [],
    fromRoute: {
      path: '/',
      fullPath: '/',
    },
    scrollbarOffset: {},
  },
  // 根mutation
  mutations: {
    mutateVersion(state: IRootState, payload: ""): void {
      state.version = payload;
    },

    mutateParameterTypes(state: IRootState, payload: string[]): void {
      state.parameterTypes = payload;
    },

    mutateFromRoute(
      state: IRootState,
      {
        to,
        from,
      }: {
        to: RouteLocationNormalized;
        from: RouteLocationNormalized;
      },
    ): void {
      if (to.path === from.path) {
        // 忽略重复
        return;
      }

      const { path, fullPath } = from;
      state.fromRoute = { path, fullPath };
    },
    mutateScrollbarOffset(
      state: IRootState,
      {
        fullPath,
        left,
        top,
      }: {
        fullPath: string;
      } & IScrollOffset,
    ) {
      const { scrollbarOffset } = state;
      scrollbarOffset[fullPath] = { left, top };
    },
  },
  // 根action
  actions: {
    async initialize({ state, commit }: ActionContext<IRootState, IRootState>): Promise<void> {
      if (state.version?.length === 0) {
        try {
          // 初始化版本
          commit('mutateVersion', await fetchVersion());
        } catch (err) {
          console.debug('fetch version failed', err.message);
        }
      }
    },
  },
});

// 动态加载模块
Object.values(import.meta.glob('./modules/*.ts', {eager: true})).forEach(({ default: module, namespace }) =>
  store.registerModule(namespace, module),
);

export default store;
