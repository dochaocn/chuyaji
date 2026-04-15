<script setup lang="ts">
import { onLaunch, onShow, onHide } from "@dcloudio/uni-app";
import { useAuthStore } from "@/store/auth";
import { useSessionStore } from "@/store/session";

onLaunch(() => {
  useAuthStore().loadToken();
  useSessionStore().load();
});
onShow(() => {
  useAuthStore().loadToken();
  useSessionStore().load();
});
onHide(() => {
  console.log("App Hide");
});
</script>

