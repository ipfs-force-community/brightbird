
export const namespace = 'worker-editor';
export default {
  namespaced: true,
  state:()=>({
    fileList:[],
    isUploadCancel:true,
  }),
  mutations:{
    setFileList(state:any, payload:[]){
      state.fileList = payload;
    },
    setUploadCancel(state:any, payload:boolean){
      state.isUploadCancel = payload;
    },
  },
};