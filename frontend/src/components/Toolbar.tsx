import type { ImageInfo, ViewMode } from '../types';
import { formatFileSize } from '../utils';

interface Props {
  image: ImageInfo | null;
  busy: boolean;
  viewMode: ViewMode;
  onOpen: () => void;
  onViewMode: (m: ViewMode) => void;
  onExportPlane: () => void;
  canExportPlane: boolean;
}

export function Toolbar({
  image,
  busy,
  viewMode,
  onOpen,
  onViewMode,
  onExportPlane,
  canExportPlane,
}: Props) {
  return (
    <header className="toolbar">
      <div className="toolbar-left">
        <span className="brand">StegSolve Go</span>
        <button type="button" className="btn primary" onClick={onOpen} disabled={busy}>
          打开图像…
        </button>
        <div className="seg">
          <button
            type="button"
            className={viewMode === 'original' || viewMode === 'bitplane' ? 'active' : ''}
            onClick={() => onViewMode('bitplane')}
            disabled={!image || busy}
          >
            位平面
          </button>
          <button
            type="button"
            className={viewMode === 'lsb' ? 'active' : ''}
            onClick={() => onViewMode('lsb')}
            disabled={!image || busy}
          >
            LSB 提取
          </button>
        </div>
        {(viewMode === 'original' || viewMode === 'bitplane') && (
          <button
            type="button"
            className="btn"
            onClick={onExportPlane}
            disabled={!canExportPlane || busy}
          >
            导出当前视图 PNG
          </button>
        )}
      </div>
      <div className="toolbar-right">
        {busy && <span className="busy-pill">处理中…</span>}
        {image ? (
          <span className="meta" title={image.name}>
            {image.name} · {image.format} · {image.width}×{image.height} ·{' '}
            {formatFileSize(image.fileSize)}
            {image.hasAlpha ? ' · Alpha' : ' · 无 Alpha'}
          </span>
        ) : (
          <span className="meta muted">未加载图像 — 打开或拖入单个文件</span>
        )}
      </div>
    </header>
  );
}
