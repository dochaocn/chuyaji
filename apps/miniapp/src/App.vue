<script setup lang="ts">
import { onLaunch, onShow, onHide } from "@dcloudio/uni-app";
import { useAuthStore } from "@/store/auth";
import { useSessionStore } from "@/store/session";

function extractInviteToken(options?: Record<string, any> | string): string {
  if (!options) return "";
  if (typeof options === "string") {
    const m = options.match(/[?&]token=([^&]+)/);
    return m ? decodeURIComponent(m[1]) : "";
  }
  const q = options.query || {};
  if (q.token) return String(q.token);
  if (options.token) return String(options.token);
  for (const raw of [options.path, options.fullPath, options.referrerInfo?.extraData?.path]) {
    const path = String(raw || "");
    const m = path.match(/[?&]token=([^&]+)/);
    if (m) return decodeURIComponent(m[1]);
  }
  return "";
}

function navigateToJoinIfPossible(token: string) {
  const pages = getCurrentPages();
  const route = String(pages[pages.length - 1]?.route || "");
  if (!route.includes("family/join")) {
    uni.navigateTo({ url: `/pages/family/join?token=${encodeURIComponent(token)}` });
  }
}

async function handlePendingInvite() {
  const session = useSessionStore();
  const auth = useAuthStore();
  const token = session.pendingInviteToken;
  if (!token || !auth.token) return;
  navigateToJoinIfPossible(token);
}

onLaunch(async (options) => {
  const auth = useAuthStore();
  auth.loadToken();
  const session = useSessionStore();
  session.load();
  const token = extractInviteToken(options as any);
  if (token) session.setPendingInvite(token);
  if (!auth.token) {
    await auth.ensureWeChatSession();
  }
  if (token && auth.token) {
    navigateToJoinIfPossible(token);
  } else {
    await handlePendingInvite();
  }
});
onShow((options) => {
  const auth = useAuthStore();
  auth.loadToken();
  const session = useSessionStore();
  session.load();
  const token = extractInviteToken(options as any);
  if (!token) return;
  session.setPendingInvite(token);
  navigateToJoinIfPossible(token);
});
onHide(() => {
  console.log("App Hide");
});
</script>

<style>
/* #ifdef H5 */
/* H5 预览专用：pages.json 里为兼容 uni-h5 保留了单 tab 配置，这里把它藏掉，
   让浏览器预览与小程序真实观感一致（小程序端不生成 tabBar）。 */
uni-tabbar {
  display: none !important;
}
/* #endif */
</style>
