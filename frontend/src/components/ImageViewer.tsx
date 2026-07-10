import { useEffect, useRef, useState } from 'react';
import { useViewport } from '../hooks/useViewport';

interface Props {
  src: string | null;
  naturalWidth?: number;
  naturalHeight?: number;
  label?: string;
  /** Show prev/next plane controls (bit-plane mode). */
  showPlaneNav?: boolean;
  planeIndexLabel?: string;
  onPrevPlane?: () => void;
  onNextPlane?: () => void;
  planeNavDisabled?: boolean;
}

export function ImageViewer({
  src,
  naturalWidth,
  naturalHeight,
  label,
  showPlaneNav,
  planeIndexLabel,
  onPrevPlane,
  onNextPlane,
  planeNavDisabled,
}: Props) {
  const vp = useViewport(1);
  const stageRef = useRef<HTMLDivElement>(null);
  const [imgSize, setImgSize] = useState({ w: 0, h: 0 });

  useEffect(() => {
    if (naturalWidth && naturalHeight) {
      setImgSize({ w: naturalWidth, h: naturalHeight });
    }
  }, [naturalWidth, naturalHeight, src]);

  // Fit scale when fitMode
  const [fitScale, setFitScale] = useState(1);
  useEffect(() => {
    const el = stageRef.current;
    if (!el || !imgSize.w || !imgSize.h) return;
    const ro = new ResizeObserver(() => {
      const pad = 24;
      const sw = Math.max(1, el.clientWidth - pad);
      const sh = Math.max(1, el.clientHeight - pad);
      const s = Math.min(sw / imgSize.w, sh / imgSize.h, 1);
      setFitScale(s > 0 ? s : 1);
    });
    ro.observe(el);
    return () => ro.disconnect();
  }, [imgSize]);

  const scale = vp.fitMode ? fitScale : vp.scale;
  const pct = Math.round(scale * 100);

  return (
    <div className="viewer">
      <div className="viewer-bar">
        <div className="viewer-bar-left">
          {showPlaneNav && (
            <div className="plane-nav viewer-plane-nav">
              <button
                type="button"
                className="btn sm plane-nav-btn"
                disabled={planeNavDisabled}
                title="上一显示模式（← / [）"
                onClick={onPrevPlane}
              >
                ◀ 上一
              </button>
              <button
                type="button"
                className="btn sm plane-nav-btn"
                disabled={planeNavDisabled}
                title="下一显示模式（→ / ]）"
                onClick={onNextPlane}
              >
                下一 ▶
              </button>
              {planeIndexLabel && (
                <span className="plane-nav-pos">{planeIndexLabel}</span>
              )}
            </div>
          )}
          <span className="viewer-label">{label || '预览'}</span>
        </div>
        <div className="zoom-controls">
          <button type="button" className="btn sm" onClick={vp.fit}>
            适应
          </button>
          <button type="button" className="btn sm" onClick={vp.actual}>
            100%
          </button>
          <button
            type="button"
            className="btn sm"
            onClick={() => vp.zoomTo(scale / 1.25)}
            disabled={scale <= 0.1}
          >
            −
          </button>
          <input
            className="zoom-range"
            type="range"
            min={10}
            max={1000}
            step={5}
            value={pct}
            onChange={(e) => vp.setZoomPercent(Number(e.target.value))}
          />
          <button
            type="button"
            className="btn sm"
            onClick={() => vp.zoomTo(scale * 1.25)}
            disabled={scale >= 10}
          >
            +
          </button>
          <span className="zoom-pct">{pct}%</span>
        </div>
      </div>
      <div
        ref={stageRef}
        className="viewer-stage"
        onPointerDown={vp.onPointerDown}
        onPointerMove={vp.onPointerMove}
        onPointerUp={vp.onPointerUp}
        onPointerLeave={vp.onPointerUp}
        onWheel={vp.onWheel}
      >
        {!src ? (
          <div className="empty-stage">
            <p>打开或拖入一张图像</p>
            <p className="hint">支持 PNG / JPEG / BMP / GIF / WebP · 单次一张 · ≤ 5000 万像素</p>
          </div>
        ) : (
          <div
            className="img-layer"
            style={{
              transform: `translate(${vp.offset.x}px, ${vp.offset.y}px) scale(${scale})`,
            }}
          >
            <img
              src={src}
              alt={label || 'image'}
              draggable={false}
              style={{
                width: imgSize.w || undefined,
                height: imgSize.h || undefined,
                imageRendering: 'pixelated',
              }}
              onLoad={(e) => {
                const im = e.currentTarget;
                setImgSize({ w: im.naturalWidth, h: im.naturalHeight });
              }}
            />
          </div>
        )}
      </div>
    </div>
  );
}
