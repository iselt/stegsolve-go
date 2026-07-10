import type { BitOrder, Channel, ExtractPreview, TraverseOrder } from '../types';
import { formatOffset, selectedBitCount } from '../utils';

export interface LSBState {
  maskA: number;
  maskR: number;
  maskG: number;
  maskB: number;
  channelOrder: Channel[];
  traverse: TraverseOrder;
  bitOrder: BitOrder;
}

interface Props {
  hasAlpha: boolean;
  state: LSBState;
  onChange: (s: LSBState) => void;
  preview: ExtractPreview | null;
  busy: boolean;
  onPreview: () => void;
  onExport: () => void;
  disabled?: boolean;
}

const ALL_CH: Channel[] = ['R', 'G', 'B', 'A'];

function toggleBit(mask: number, bit: number): number {
  return mask ^ (1 << bit);
}

function setMask(state: LSBState, ch: Channel, mask: number): LSBState {
  const next = { ...state };
  if (ch === 'A') next.maskA = mask;
  if (ch === 'R') next.maskR = mask;
  if (ch === 'G') next.maskG = mask;
  if (ch === 'B') next.maskB = mask;
  return next;
}

function getMask(state: LSBState, ch: Channel): number {
  if (ch === 'A') return state.maskA;
  if (ch === 'R') return state.maskR;
  if (ch === 'G') return state.maskG;
  return state.maskB;
}

export const defaultLSBState = (): LSBState => ({
  maskA: 0,
  maskR: 0,
  maskG: 0,
  maskB: 0,
  channelOrder: ['R', 'G', 'B', 'A'],
  traverse: 'row',
  bitOrder: 'lsb',
});

export function LSBPanel({
  hasAlpha,
  state,
  onChange,
  preview,
  busy,
  onPreview,
  onExport,
  disabled,
}: Props) {
  const count = selectedBitCount(state);
  const canRun = count > 0 && !disabled && !busy;

  const applyPreset = (name: string) => {
    let next = { ...state, maskA: 0, maskR: 0, maskG: 0, maskB: 0 };
    switch (name) {
      case 'R0':
        next.maskR = 1;
        break;
      case 'G0':
        next.maskG = 1;
        break;
      case 'B0':
        next.maskB = 1;
        break;
      case 'RGB0':
        next.maskR = 1;
        next.maskG = 1;
        next.maskB = 1;
        break;
      case 'RGBA0':
        next.maskR = 1;
        next.maskG = 1;
        next.maskB = 1;
        if (hasAlpha) next.maskA = 1;
        break;
      case 'all':
        next.maskR = 0xff;
        next.maskG = 0xff;
        next.maskB = 0xff;
        next.maskA = hasAlpha ? 0xff : 0;
        break;
      case 'clear':
        break;
    }
    onChange(next);
  };

  const moveChannel = (ch: Channel, dir: -1 | 1) => {
    const order = [...state.channelOrder];
    const i = order.indexOf(ch);
    if (i < 0) return;
    const j = i + dir;
    if (j < 0 || j >= order.length) return;
    [order[i], order[j]] = [order[j], order[i]];
    onChange({ ...state, channelOrder: order });
  };

  return (
    <aside className="side-panel lsb-panel">
      <h3>LSB 提取</h3>
      <p className="hint">勾选通道位；首个提取位装入输出字节 bit7</p>

      <div className="preset-row">
        {['R0', 'G0', 'B0', 'RGB0', 'RGBA0'].map((p) => (
          <button
            key={p}
            type="button"
            className="btn sm"
            disabled={disabled || (p === 'RGBA0' && !hasAlpha)}
            onClick={() => applyPreset(p)}
          >
            {p}
          </button>
        ))}
        <button type="button" className="btn sm" disabled={disabled} onClick={() => applyPreset('all')}>
          全选
        </button>
        <button type="button" className="btn sm" disabled={disabled} onClick={() => applyPreset('clear')}>
          清空
        </button>
      </div>

      {ALL_CH.map((ch) => {
        const chOff = disabled || (ch === 'A' && !hasAlpha);
        const mask = getMask(state, ch);
        return (
          <div key={ch} className={`channel-row ${chOff ? 'disabled' : ''}`}>
            <span className={`ch-label ch-${ch}`}>{ch}</span>
            <div className="bit-grid check">
              {Array.from({ length: 8 }, (_, bit) => {
                const on = (mask & (1 << bit)) !== 0;
                return (
                  <label key={bit} className={`bit-check ${on ? 'on' : ''}`} title={`${ch}${bit}`}>
                    <input
                      type="checkbox"
                      checked={on}
                      disabled={chOff}
                      onChange={() => onChange(setMask(state, ch, toggleBit(mask, bit)))}
                    />
                    <span>{bit}</span>
                  </label>
                );
              })}
            </div>
          </div>
        );
      })}

      <div className="field">
        <label>遍历顺序</label>
        <select
          value={state.traverse}
          disabled={disabled}
          onChange={(e) => onChange({ ...state, traverse: e.target.value as TraverseOrder })}
        >
          <option value="row">逐行（Row）</option>
          <option value="col">逐列（Column）</option>
        </select>
      </div>

      <div className="field">
        <label>通道内位序</label>
        <select
          value={state.bitOrder}
          disabled={disabled}
          onChange={(e) => onChange({ ...state, bitOrder: e.target.value as BitOrder })}
        >
          <option value="lsb">LSB → MSB</option>
          <option value="msb">MSB → LSB</option>
        </select>
      </div>

      <div className="field">
        <label>通道顺序</label>
        <div className="ch-order">
          {state.channelOrder.map((ch) => {
            const off = ch === 'A' && !hasAlpha;
            return (
              <div key={ch} className={`ch-order-item ${off ? 'disabled' : ''}`}>
                <span className={`ch-label ch-${ch}`}>{ch}</span>
                <button type="button" className="btn sm" disabled={disabled || off} onClick={() => moveChannel(ch, -1)}>
                  ↑
                </button>
                <button type="button" className="btn sm" disabled={disabled || off} onClick={() => moveChannel(ch, 1)}>
                  ↓
                </button>
              </div>
            );
          })}
        </div>
      </div>

      <div className="actions">
        <button type="button" className="btn primary" disabled={!canRun} onClick={onPreview}>
          生成预览
        </button>
        <button type="button" className="btn" disabled={!canRun} onClick={onExport}>
          导出 .bin
        </button>
      </div>

      <p className="hint">已选 {count} 位{count === 0 ? ' — 预览/导出已禁用' : ''}</p>

      {preview && (
        <div className="hex-preview">
          <div className="hex-meta">
            总长度 {preview.totalBytes} 字节 · 预览 {preview.previewBytes}
            {preview.truncated ? '（已截断）' : ''}
          </div>
          <div className="hex-table">
            {preview.rows.map((row) => (
              <div key={row.offset} className="hex-row">
                <span className="off">{formatOffset(row.offset)}</span>
                <span className="hex">{row.hex}</span>
                <span className="asc">{row.ascii}</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </aside>
  );
}
