/// <reference types="@dcloudio/types" />

interface ImportMetaEnv {
  readonly VITE_CHUYAJI_API_BASE: string;
  /** 为 "true" 时登录请求固定发 code=dev，须后端 CHUYAJI_DEV_MODE=true */
  readonly VITE_DEV_WECHAT_LOGIN?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}

declare module "*.vue" {
  import type { DefineComponent } from "vue";
  const component: DefineComponent<{}, {}, any>;
  export default component;
}
