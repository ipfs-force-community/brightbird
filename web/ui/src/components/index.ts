import { App } from 'vue';
import  ElementPlus from 'element-plus';
import lang from 'element-plus/dist/locale/zh-cn.js';
import 'dayjs/locale/zh-cn';
// 设置element-plus自定义主题色样式
import './theme/custom-element-plus/index.scss';
import 'bootstrap/dist/css/bootstrap.min.css';
// 设置公共组件全局样式

// 必须晚于element-plus全局样式加载，否则，无法覆盖
import './theme/index.less';
// 设置jm-icon样式
import './theme/icon/button/css/jm-icon-button.css';
import './theme/icon/tabs/css/jm-icon-tab.css';
import './theme/icon/link/css/jm-icon-link.css';
import './theme/icon/menu/css/jm-icon-menu.css';
import './theme/icon/input/css/jm-icon-input.css';
import 'vue-json-pretty/lib/styles.css';
import './theme/icon/workflow/css/jm-icon-workflow.css';
import * as ElementPlusIconsVue from '@element-plus/icons-vue';

import JmLoading from './notice/loading';
import JmMessage from './notice/message';
import JmMessageBox from './notice/message-box';
import JmInfiniteScroll from './infinite-scroll';

import {VueJsonPretty} from 'vue-json-pretty'
export default {
  // app.use()触发install的调用
  install: (app: App) => {
    app.use(ElementPlus, {
      locale: lang,
    });
    for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
      app.component(key, component);
    }
    // 动态加载组件
    Object.values(import.meta.glob('./**/index.vue', { eager: true }))
      .concat(Object.values(import.meta.glob('./**/index.ts', { eager: true })))
      // 全局注册组件
      .forEach(({ default: component }) => {
        if (!component.name) {
          return;
        }
        app.component(component.name, component);
      });

    app.use(VueJsonPretty);
    
    // 全局注册指令
    app.directive('scroll', JmInfiniteScroll.directive);
    // 全局注册变量
    app.config.globalProperties.$loading = JmLoading.service;
    app.config.globalProperties.$success = JmMessage.success;
    app.config.globalProperties.$warning = JmMessage.warning;
    app.config.globalProperties.$info = JmMessage.info;
    app.config.globalProperties.$error = JmMessage.error;
    app.config.globalProperties.$alert = JmMessageBox.alert;
    app.config.globalProperties.$confirm = JmMessageBox.confirm;
    app.config.globalProperties.$prompt = JmMessageBox.prompt;
  },
};
