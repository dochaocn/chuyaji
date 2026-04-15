import { API_BASE } from "./config";

type Method = "GET" | "POST" | "PATCH" | "DELETE";

export interface RequestOptions {
  path: string;
  method?: Method;
  data?: Record<string, unknown>;
  token?: string | null;
}

export function getStoredToken(): string {
  try {
    const t = uni.getStorageSync("chuyaji_token") as string;
    return t || "";
  } catch {
    return "";
  }
}

export function request<T>(opts: RequestOptions): Promise<T> {
  const header: Record<string, string> = {
    "Content-Type": "application/json",
  };
  const tok = opts.token !== undefined ? opts.token : getStoredToken();
  if (tok) {
    header.Authorization = `Bearer ${tok}`;
  }
  const url = `${API_BASE}${opts.path}`;
  return new Promise((resolve, reject) => {
    uni.request({
      url,
      method: (opts.method || "GET") as UniApp.RequestOptions["method"],
      data: opts.data,
      header,
      timeout: 60000,
      success: (res) => {
        const status = res.statusCode || 0;
        if (status === 401) {
          try {
            uni.removeStorageSync("chuyaji_token");
          } catch {
            /* ignore */
          }
        }
        if (status >= 200 && status < 300) {
          resolve(res.data as T);
          return;
        }
        reject(new Error(`HTTP ${status}: ${JSON.stringify(res.data)}`));
      },
      fail: (err) => {
        const msg =
          err && typeof err === "object" && "errMsg" in err
            ? String((err as { errMsg?: string }).errMsg)
            : String(err);
        console.error("[chuyaji] request fail", url, msg);
        reject(new Error(msg));
      },
    });
  });
}
