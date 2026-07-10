export function formatFileSize(n: number): string {
  if (n < 1024) return `${n} B`;
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`;
  return `${(n / (1024 * 1024)).toFixed(2)} MB`;
}

export function formatOffset(n: number): string {
  return n.toString(16).toUpperCase().padStart(8, '0');
}

export function pngDataUrl(b64: string): string {
  return `data:image/png;base64,${b64}`;
}

export function selectedBitCount(masks: {
  maskA: number;
  maskR: number;
  maskG: number;
  maskB: number;
}): number {
  const pop = (v: number) => {
    let n = 0;
    let x = v & 0xff;
    while (x) {
      n += x & 1;
      x >>= 1;
    }
    return n;
  };
  return pop(masks.maskA) + pop(masks.maskR) + pop(masks.maskG) + pop(masks.maskB);
}

export function clamp(v: number, lo: number, hi: number): number {
  return Math.max(lo, Math.min(hi, v));
}
