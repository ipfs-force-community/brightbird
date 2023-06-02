import { ConfigEnv, UserConfigExport } from 'vite';
import vue from '@vitejs/plugin-vue';
import { resolve } from 'path';
// https://vitejs.dev/config/
export default ({ command, mode }: ConfigEnv): UserConfigExport => {
  return {
    plugins: [vue()],
    // base public path
    base: '/',
    resolve: {
      alias: {
        '@': resolve(__dirname, 'src'),
        // 解决：
        // Component provided template option but runtime compilation is not supported in this build of Vue
        // Configure your bundler to alias "vue" to "vue/dist/vue.esm-bundler.js".
        vue: 'vue/dist/vue.esm-bundler.js',
      },
    },
    server: {},
    build: {
      outDir: process.env.PUBLICDIR
    }
  };
};
