import { useCallback, useEffect, useMemo, useState } from 'react';
import {
  LoadDroppedImage,
  OpenImage,
  PreviewLSB,
  RenderBitPlane,
  SaveBitPlane,
  SaveLSB,
} from '../wailsjs/go/main/App';
import { OnFileDrop, OnFileDropOff } from '../wailsjs/runtime/runtime';
import {
  BitPlanePanel,
  movePlaneInMatrix,
  planeIndex,
  planeSequence,
  stepPlane,
  type MatrixDir,
  type PlaneSel,
} from './components/BitPlanePanel';
import { ImageViewer } from './components/ImageViewer';
import { defaultLSBState, LSBPanel, type LSBState } from './components/LSBPanel';
import { Toolbar } from './components/Toolbar';
import type { ExtractPreview, ImageInfo, ViewMode } from './types';
import { pngDataUrl } from './utils';
import './App.css';

function errMessage(e: unknown): string {
  if (e instanceof Error) return e.message;
  if (typeof e === 'string') return e;
  try {
    return JSON.stringify(e);
  } catch {
    return String(e);
  }
}

function App() {
  const [image, setImage] = useState<ImageInfo | null>(null);
  const [viewMode, setViewMode] = useState<ViewMode>('bitplane');
  const [plane, setPlane] = useState<PlaneSel>({ kind: 'original' });
  const [viewSrc, setViewSrc] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [status, setStatus] = useState<string | null>(null);
  const [lsb, setLsb] = useState<LSBState>(defaultLSBState);
  const [preview, setPreview] = useState<ExtractPreview | null>(null);

  const applyImage = useCallback((info: ImageInfo | null) => {
    if (!info) return;
    setImage(info);
    setPlane({ kind: 'original' });
    setViewSrc(pngDataUrl(info.previewPngB64));
    setViewMode('bitplane');
    setPreview(null);
    setLsb(defaultLSBState());
    setError(null);
    setStatus(`已加载 ${info.name}`);
  }, []);

  const withBusy = useCallback(async <T,>(fn: () => Promise<T>): Promise<T | undefined> => {
    setBusy(true);
    setError(null);
    try {
      return await fn();
    } catch (e) {
      setError(errMessage(e));
      return undefined;
    } finally {
      setBusy(false);
    }
  }, []);

  const handleOpen = useCallback(async () => {
    const info = await withBusy(() => OpenImage());
    if (info === undefined) return;
    if (info === null) {
      setStatus('已取消打开');
      return;
    }
    applyImage(info);
  }, [applyImage, withBusy]);

  const handleDropPaths = useCallback(
    async (paths: string[]) => {
      if (!paths || paths.length === 0) return;
      if (paths.length > 1) {
        setError('一次仅支持拖入一个图像文件');
        return;
      }
      const info = await withBusy(() => LoadDroppedImage(paths[0]));
      if (info) applyImage(info);
    },
    [applyImage, withBusy],
  );

  useEffect(() => {
    OnFileDrop((_x, _y, paths) => {
      void handleDropPaths(paths);
    }, true);
    return () => {
      OnFileDropOff();
    };
  }, [handleDropPaths]);

  useEffect(() => {
    if (!image || viewMode === 'lsb') return;
    let cancelled = false;
    const run = async () => {
      if (plane.kind === 'original') {
        setViewSrc(pngDataUrl(image.previewPngB64));
        return;
      }
      setBusy(true);
      setError(null);
      try {
        const b64 = await RenderBitPlane({
          imageId: image.imageId,
          channel: plane.channel,
          bit: plane.bit,
        });
        if (!cancelled) {
          setViewSrc(pngDataUrl(b64));
        }
      } catch (e) {
        if (!cancelled) setError(errMessage(e));
      } finally {
        if (!cancelled) setBusy(false);
      }
    };
    void run();
    return () => {
      cancelled = true;
    };
  }, [image, plane, viewMode]);

  const stepPlaneBy = useCallback(
    (delta: number) => {
      if (!image || viewMode === 'lsb') return;
      setPlane((p) => stepPlane(p, image.hasAlpha, delta));
    },
    [image, viewMode],
  );

  const moveInMatrix = useCallback(
    (dir: MatrixDir) => {
      if (!image || viewMode === 'lsb') return;
      setPlane((p) => movePlaneInMatrix(p, image.hasAlpha, dir));
    },
    [image, viewMode],
  );

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (!image || viewMode === 'lsb') return;
      // Ignore when typing in form controls
      const t = e.target as HTMLElement | null;
      if (t && (t.tagName === 'INPUT' || t.tagName === 'TEXTAREA' || t.tagName === 'SELECT' || t.isContentEditable)) {
        return;
      }
      // Matrix navigation: arrows move within RGBA × bit0–7 grid
      if (e.key === 'ArrowLeft') {
        e.preventDefault();
        moveInMatrix('left');
      } else if (e.key === 'ArrowRight') {
        e.preventDefault();
        moveInMatrix('right');
      } else if (e.key === 'ArrowUp') {
        e.preventDefault();
        moveInMatrix('up');
      } else if (e.key === 'ArrowDown') {
        e.preventDefault();
        moveInMatrix('down');
      } else if (e.key === '[') {
        // Linear prev/next (sequence order) still available
        e.preventDefault();
        stepPlaneBy(-1);
      } else if (e.key === ']') {
        e.preventDefault();
        stepPlaneBy(1);
      }
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [image, viewMode, stepPlaneBy, moveInMatrix]);

  const planeLabel = useMemo(() => {
    if (!image) return '';
    if (plane.kind === 'original') return '原图';
    const tag = plane.bit === 0 ? 'LSB' : plane.bit === 7 ? 'MSB' : '';
    return `${plane.channel} bit ${plane.bit}${tag ? ` (${tag})` : ''}`;
  }, [image, plane]);

  const planeIndexLabel = useMemo(() => {
    if (!image) return '';
    const total = planeSequence(image.hasAlpha).length;
    const idx = planeIndex(plane, image.hasAlpha);
    return `${idx + 1}/${total}`;
  }, [image, plane]);

  const handleExportPlane = useCallback(async () => {
    if (!image) return;
    const req =
      plane.kind === 'original'
        ? { imageId: image.imageId, channel: 'ORIG' as const, bit: 0 }
        : { imageId: image.imageId, channel: plane.channel, bit: plane.bit };
    const res = await withBusy(() => SaveBitPlane(req));
    if (!res) return;
    if (res.cancelled) {
      setStatus('已取消导出');
      return;
    }
    setStatus(`已导出 ${res.path}（${res.bytes} 字节）`);
  }, [image, plane, withBusy]);

  const buildExtractReq = useCallback(() => {
    if (!image) throw new Error('尚未加载图像');
    const order = image.hasAlpha
      ? lsb.channelOrder
      : lsb.channelOrder.filter((c) => c !== 'A');
    return {
      imageId: image.imageId,
      maskA: image.hasAlpha ? lsb.maskA : 0,
      maskR: lsb.maskR,
      maskG: lsb.maskG,
      maskB: lsb.maskB,
      channelOrder: order,
      traverse: lsb.traverse,
      bitOrder: lsb.bitOrder,
    };
  }, [image, lsb]);

  const handlePreviewLSB = useCallback(async () => {
    const res = await withBusy(async () => PreviewLSB(buildExtractReq()));
    if (res) {
      setPreview(res);
      setStatus(`预览 ${res.previewBytes} / ${res.totalBytes} 字节`);
    }
  }, [buildExtractReq, withBusy]);

  const handleExportLSB = useCallback(async () => {
    const res = await withBusy(async () => SaveLSB(buildExtractReq()));
    if (!res) return;
    if (res.cancelled) {
      setStatus('已取消导出');
      return;
    }
    setStatus(`已导出 ${res.path}（${res.bytes} 字节）`);
  }, [buildExtractReq, withBusy]);

  return (
    <div className="app" style={{ ['--wails-drop-target' as string]: 'drop' }}>
      <Toolbar
        image={image}
        busy={busy}
        viewMode={viewMode}
        onOpen={handleOpen}
        onViewMode={setViewMode}
        onExportPlane={handleExportPlane}
        canExportPlane={!!image}
      />

      {(error || status) && (
        <div className={`banner ${error ? 'error' : 'ok'}`}>
          <span>{error || status}</span>
          <button
            type="button"
            className="btn sm"
            onClick={() => {
              setError(null);
              setStatus(null);
            }}
          >
            关闭
          </button>
        </div>
      )}

      <div className="main">
        {viewMode === 'lsb' ? (
          <LSBPanel
            hasAlpha={!!image?.hasAlpha}
            state={lsb}
            onChange={(s) => {
              setLsb(s);
              setPreview(null);
            }}
            preview={preview}
            busy={busy}
            onPreview={handlePreviewLSB}
            onExport={handleExportLSB}
            disabled={!image}
          />
        ) : (
          <BitPlanePanel
            hasAlpha={!!image?.hasAlpha}
            selection={plane}
            onSelect={setPlane}
            onPrev={() => stepPlaneBy(-1)}
            onNext={() => stepPlaneBy(1)}
            disabled={!image}
          />
        )}

        <ImageViewer
          src={viewMode === 'lsb' ? (image ? pngDataUrl(image.previewPngB64) : null) : viewSrc}
          naturalWidth={image?.width}
          naturalHeight={image?.height}
          label={viewMode === 'lsb' ? '原图参考' : planeLabel}
          showPlaneNav={viewMode !== 'lsb' && !!image}
          planeIndexLabel={planeIndexLabel}
          onPrevPlane={() => stepPlaneBy(-1)}
          onNextPlane={() => stepPlaneBy(1)}
          planeNavDisabled={!image}
        />
      </div>
    </div>
  );
}

export default App;
