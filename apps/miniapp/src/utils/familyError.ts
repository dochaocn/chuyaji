export function parseFamilyError(err: unknown): string {
  const msg = String((err as { message?: string })?.message || (err as { error?: string })?.error || err || "");
  if (!msg || msg.includes("request:fail") || msg.includes("connect")) {
    return "无法连接服务器，请确认后端 API 已启动且地址配置正确。";
  }
  if (msg.includes("timeout")) return "请求超时，请检查网络后重试。";
  if (msg.includes("401") || msg.includes("unauthorized")) return "登录已过期，请重新打开小程序。";
  if (msg.includes("expired")) return "邀请已过期，请让家人重新分享。";
  if (msg.includes("already used") || msg.includes("revoked")) return "该邀请已失效，请让家人重新生成。";
  if (msg.includes("invite not found")) return "邀请无效，请让家人重新分享。";
  if (msg.includes("family has other members")) {
    return "你是当前家庭的主人且还有其他成员，请先处理现有家庭后再加入。";
  }
  if (msg.includes("403") || msg.includes("forbidden")) return "仅家庭主人可以生成邀请。";
  if (msg.includes("404") && msg.includes("no family")) return "还没有家庭，请先创建家庭。";
  if (msg.includes("500") && msg.includes("create")) return "服务器写入失败，请重启后端以完成数据库迁移。";
  if (msg.includes("500")) return "服务器异常，请稍后重试。";
  return "操作失败，请稍后重试。";
}
