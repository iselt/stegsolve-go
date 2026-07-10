import type { Channel } from '../types';

export type PlaneSel =
  | { kind: 'original' }
  | { kind: 'plane'; channel: Channel; bit: number };

interface Props {
  hasAlpha: boolean;
  selection: PlaneSel;
  onSelect: (s: PlaneSel) => void;
  onPrev?: () => void;
  onNext?: () => void;
  disabled?: boolean;
}

const CHANNELS: Channel[] = ['R', 'G', 'B', 'A'];

export function BitPlanePanel({
  hasAlpha,
  selection,
  onSelect,
  onPrev,
  onNext,
  disabled,
}: Props) {
  const isOrig = selection.kind === 'original';
  const curCh = selection.kind === 'plane' ? selection.channel : null;
  const curBit = selection.kind === 'plane' ? selection.bit : -1;
  const seq = planeSequence(hasAlpha);
  const idx = planeIndex(selection, hasAlpha);

  return (
    <aside className="side-panel">
      <h3>位平面</h3>
      <p className="hint">0 = LSB，7 = MSB；位为 1 显示白，0 显示黑</p>

      <div className="plane-nav">
        <button
          type="button"
          className="btn plane-nav-btn"
          disabled={disabled}
          title="上一模式（←）"
          onClick={onPrev}
        >
          ◀ 上一
        </button>
        <span className="plane-nav-pos" title="当前显示模式序号">
          {idx + 1}/{seq.length}
        </span>
        <button
          type="button"
          className="btn plane-nav-btn"
          disabled={disabled}
          title="下一模式（→）"
          onClick={onNext}
        >
          下一 ▶
        </button>
      </div>
      <p className="hint">顺序：原图 → R0…R7 → G0…G7 → B0…B7{hasAlpha ? ' → A0…A7' : ''}</p>

      <button
        type="button"
        className={`plane-btn orig ${isOrig ? 'active' : ''}`}
        disabled={disabled}
        onClick={() => onSelect({ kind: 'original' })}
      >
        原图
      </button>

      {CHANNELS.map((ch) => {
        const chDisabled = disabled || (ch === 'A' && !hasAlpha);
        return (
          <div key={ch} className={`channel-row ${chDisabled ? 'disabled' : ''}`}>
            <span className={`ch-label ch-${ch}`}>{ch}</span>
            <div className="bit-grid">
              {Array.from({ length: 8 }, (_, i) => {
                const bit = i; // show 0..7 left to right (LSB→MSB)
                const active = curCh === ch && curBit === bit;
                return (
                  <button
                    key={bit}
                    type="button"
                    className={`bit-btn ${active ? 'active' : ''}`}
                    disabled={chDisabled}
                    title={`${ch} bit ${bit}${bit === 0 ? ' (LSB)' : bit === 7 ? ' (MSB)' : ''}`}
                    onClick={() => onSelect({ kind: 'plane', channel: ch, bit })}
                  >
                    {bit}
                  </button>
                );
              })}
            </div>
          </div>
        );
      })}

      {!hasAlpha && (
        <p className="hint warn">当前图像无有效 Alpha（全部为 255），A 通道已禁用</p>
      )}

      <p className="hint">
        方向键在矩阵内移动：←→ 改位，↑↓ 改通道（R/G/B/A）；[ ] 按线性顺序切换
      </p>
    </aside>
  );
}

/**
 * Walk order: 原图 → R0…R7 → G0…G7 → B0…B7 → (A0…A7 if alpha).
 * bit0 = LSB, bit7 = MSB.
 */
export function planeSequence(hasAlpha: boolean): PlaneSel[] {
  const seq: PlaneSel[] = [{ kind: 'original' }];
  const chs: Channel[] = hasAlpha ? ['R', 'G', 'B', 'A'] : ['R', 'G', 'B'];
  for (const ch of chs) {
    for (let bit = 0; bit <= 7; bit++) {
      seq.push({ kind: 'plane', channel: ch, bit });
    }
  }
  return seq;
}

export function planeIndex(cur: PlaneSel, hasAlpha: boolean): number {
  const seq = planeSequence(hasAlpha);
  const idx = seq.findIndex((s) => {
    if (s.kind === 'original' && cur.kind === 'original') return true;
    if (s.kind === 'plane' && cur.kind === 'plane') {
      return s.channel === cur.channel && s.bit === cur.bit;
    }
    return false;
  });
  return idx < 0 ? 0 : idx;
}

export function stepPlane(cur: PlaneSel, hasAlpha: boolean, delta: number): PlaneSel {
  const seq = planeSequence(hasAlpha);
  const idx = planeIndex(cur, hasAlpha);
  return seq[(idx + delta + seq.length) % seq.length];
}

/** Channels in matrix row order (top → bottom). */
export function matrixChannels(hasAlpha: boolean): Channel[] {
  return hasAlpha ? ['R', 'G', 'B', 'A'] : ['R', 'G', 'B'];
}

export type MatrixDir = 'left' | 'right' | 'up' | 'down';

/**
 * Move within the RGBA × bit0–7 matrix:
 * - left/right: same channel, bit ±1 (wrap across channels at edges)
 * - up/down: same bit, channel ±1 (wrap); from 原图, down/right → R0
 */
export function movePlaneInMatrix(cur: PlaneSel, hasAlpha: boolean, dir: MatrixDir): PlaneSel {
  const chs = matrixChannels(hasAlpha);

  if (cur.kind === 'original') {
    if (dir === 'down' || dir === 'right') {
      return { kind: 'plane', channel: 'R', bit: 0 };
    }
    // up / left from original → last cell (A7 or B7)
    const lastCh = chs[chs.length - 1];
    return { kind: 'plane', channel: lastCh, bit: 7 };
  }

  let chIdx = chs.indexOf(cur.channel);
  if (chIdx < 0) chIdx = 0;
  let bit = cur.bit;

  switch (dir) {
    case 'left':
      bit -= 1;
      if (bit < 0) {
        bit = 7;
        chIdx -= 1;
        if (chIdx < 0) {
          return { kind: 'original' };
        }
      }
      break;
    case 'right':
      bit += 1;
      if (bit > 7) {
        bit = 0;
        chIdx += 1;
        if (chIdx >= chs.length) {
          return { kind: 'original' };
        }
      }
      break;
    case 'up':
      chIdx -= 1;
      if (chIdx < 0) {
        return { kind: 'original' };
      }
      break;
    case 'down':
      chIdx += 1;
      if (chIdx >= chs.length) {
        return { kind: 'original' };
      }
      break;
  }

  return { kind: 'plane', channel: chs[chIdx], bit };
}
