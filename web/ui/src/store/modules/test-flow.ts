export const namespace = 'test-flow';
export default {
  namespaced: true,
  state: () => ({
    toggle:Boolean,
  }),
  mutations: {
    mutate(state: any, payload: { id: string, status: boolean }) {
      console.log('+========projectGroupFoldingMapping===========', payload);
      const { id, status } = payload;
      // 根据项目组id存入对应的折叠状态
      state[id] = status;
    },
  },
  actions: {},
};
