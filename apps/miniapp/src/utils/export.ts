import { apiListRecords } from "@/api/chuyaji";
import type { RecordItem } from "@/api/chuyaji";

export async function exportAllRecordsJson(babyId: number): Promise<string> {
  const all: RecordItem[] = [];
  let cursor: string | undefined;
  for (;;) {
    const page = await apiListRecords(babyId, 50, cursor);
    all.push(...page.items);
    if (!page.next_cursor) break;
    cursor = page.next_cursor;
  }
  return JSON.stringify(
    {
      exported_at: new Date().toISOString(),
      baby_id: babyId,
      count: all.length,
      records: all,
    },
    null,
    2
  );
}

export function copyToClipboard(text: string) {
  return new Promise<void>((resolve, reject) => {
    uni.setClipboardData({
      data: text,
      success: () => resolve(),
      fail: (e) => reject(e),
    });
  });
}
