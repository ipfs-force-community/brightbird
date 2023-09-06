import { fetchLabel } from '@/api/plugin';
import { Commit } from 'vuex';
export const namespace = 'worker-editor';
export default {
  namespaced: true,
  state:()=>({
    fileList:[],
    isUploadCancel:true,
    labels: [],
  }),
  mutations:{
    setFileList(state:any, payload:[]){
      state.fileList = payload;
    },
    setUploadCancel(state:any, payload:boolean){
      state.isUploadCancel = payload;
    },
    setLabels(state:any, payload:string[]){
      state.labels = payload;
    },
  },
  actions:{
    async getLabels({ commit }){
      const res =  await fetchLabel();
      commit('setLabels', res);
    },
  },
};