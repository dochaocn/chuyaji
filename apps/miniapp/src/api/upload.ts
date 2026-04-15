import { API_BASE } from "./config";
import { getStoredToken } from "./http";

export interface UploadResult {
  id: number;
  record_id: number;
  url: string;
  thumb_url?: string;
  sort_order: number;
  size: number;
}

export function uploadRecordAttachment(recordId: number, filePath: string): Promise<UploadResult> {
  const token = getStoredToken();
  return new Promise((resolve, reject) => {
    uni.uploadFile({
      url: `${API_BASE}/api/v1/records/${recordId}/attachments/upload`,
      filePath,
      name: "file",
      header: token ? { Authorization: `Bearer ${token}` } : {},
      success: (res) => {
        const status = res.statusCode || 0;
        if (status >= 200 && status < 300) {
          try {
            resolve(JSON.parse(res.data as string) as UploadResult);
          } catch (e) {
            reject(e);
          }
          return;
        }
        reject(new Error(`upload ${status}: ${res.data}`));
      },
      fail: (e) => reject(e),
    });
  });
}
