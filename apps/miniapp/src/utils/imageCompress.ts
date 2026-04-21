/** 与线上 Nginx `client_max_body_size 1m` 对齐；上传前压到此大小以下 */
export const MAX_UPLOAD_IMAGE_BYTES = 1 * 1024 * 1024;

function getFileInfoPromise(filePath: string): Promise<{ size: number }> {
  return new Promise((resolve, reject) => {
    uni.getFileInfo({
      filePath,
      success: (res) => resolve({ size: res.size }),
      fail: reject,
    });
  });
}

function getImageInfoPromise(src: string): Promise<{ width: number; height: number }> {
  return new Promise((resolve, reject) => {
    uni.getImageInfo({
      src,
      success: (res) => resolve({ width: res.width, height: res.height }),
      fail: reject,
    });
  });
}

function compressImagePromise(
  src: string,
  opts: { quality: number; compressedWidth?: number }
): Promise<string> {
  return new Promise((resolve, reject) => {
    uni.compressImage({
      src,
      quality: opts.quality,
      compressedWidth: opts.compressedWidth,
      success: (res) => resolve(res.tempFilePath),
      fail: reject,
    });
  });
}

/**
 * 若本地临时文件超过 maxBytes，则依次降低质量 / 缩小宽度再压缩，直至不超过上限（尽力而为）。
 */
export async function ensureImageUnderMaxBytes(
  filePath: string,
  maxBytes = MAX_UPLOAD_IMAGE_BYTES
): Promise<string> {
  let path = filePath;
  const { size } = await getFileInfoPromise(path);
  if (size <= maxBytes) return path;

  const qualities = [75, 65, 55, 45, 35, 28, 22];
  for (const q of qualities) {
    path = await compressImagePromise(path, { quality: q });
    const { size: s } = await getFileInfoPromise(path);
    if (s <= maxBytes) return path;
  }

  let img = await getImageInfoPromise(path);
  let w = img.width;
  while (w > 360) {
    w = Math.floor(w * 0.72);
    path = await compressImagePromise(path, { quality: 70, compressedWidth: w });
    const { size: s } = await getFileInfoPromise(path);
    if (s <= maxBytes) return path;
    img = await getImageInfoPromise(path);
  }

  path = await compressImagePromise(path, { quality: 18 });
  const { size: last } = await getFileInfoPromise(path);
  if (last > maxBytes) {
    uni.showToast({ title: "图片仍较大，请换一张或裁剪后再试", icon: "none" });
  }
  return path;
}
