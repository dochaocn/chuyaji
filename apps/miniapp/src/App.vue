<script setup lang="ts">
import { onLaunch, onShow, onHide } from "@dcloudio/uni-app";
import { useAuthStore } from "@/store/auth";
import { useSessionStore } from "@/store/session";

onLaunch(async () => {
  const auth = useAuthStore();
  auth.loadToken();
  useSessionStore().load();
  if (!auth.token) {
    await auth.ensureWeChatSession();
  }
});
onShow(() => {
  const auth = useAuthStore();
  auth.loadToken();
  useSessionStore().load();
});
onHide(() => {
  console.log("App Hide");
});
</script>

