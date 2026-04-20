/** 调用 uni.login，将返回结果交给后端完成就绪校验。 */
export function getWeChatLoginCode(): Promise<string> {
  return new Promise((resolve, reject) => {
    uni.login({
      provider: "weixin",
      success: (res) => {
        if (res.code) resolve(res.code);
        else reject(new Error("missing wx.temp"));
      },
      fail: (err) => reject(err ?? new Error("uni.login failed")),
    });
  });
}

/** 登出时调用，与旧缓存逻辑兼容（现为 no-op） */
export function clearWeChatLoginCodeCache(): void {}
