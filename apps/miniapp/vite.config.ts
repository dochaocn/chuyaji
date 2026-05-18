import { defineConfig } from "vite";
import uni from "@dcloudio/vite-plugin-uni";

export default defineConfig({
  plugins: [uni()],
  css: {
    preprocessorOptions: {
      scss: {
        // uni 会在 SFC 样式前注入片段，无法满足 @use 必须置顶；继续用 @import 引入共享片段并静默弃用告警
        silenceDeprecations: ["import"],
      },
    },
  },
});
