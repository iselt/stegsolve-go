import { useCallback, useRef, useState } from 'react';
import { clamp } from '../utils';

export function useViewport(initialScale = 1) {
  const [scale, setScale] = useState(initialScale);
  const [offset, setOffset] = useState({ x: 0, y: 0 });
  const [fitMode, setFitMode] = useState(true);
  const dragRef = useRef<{ x: number; y: number; ox: number; oy: number } | null>(null);

  const setZoomPercent = useCallback((pct: number) => {
    setFitMode(false);
    setScale(clamp(pct / 100, 0.1, 10));
  }, []);

  const zoomTo = useCallback((s: number) => {
    setFitMode(false);
    setScale(clamp(s, 0.1, 10));
  }, []);

  const fit = useCallback(() => {
    setFitMode(true);
    setOffset({ x: 0, y: 0 });
  }, []);

  const actual = useCallback(() => {
    setFitMode(false);
    setScale(1);
    setOffset({ x: 0, y: 0 });
  }, []);

  const onPointerDown = useCallback(
    (e: React.PointerEvent) => {
      if (e.button !== 0) return;
      (e.target as HTMLElement).setPointerCapture?.(e.pointerId);
      dragRef.current = { x: e.clientX, y: e.clientY, ox: offset.x, oy: offset.y };
    },
    [offset],
  );

  const onPointerMove = useCallback((e: React.PointerEvent) => {
    const d = dragRef.current;
    if (!d) return;
    setFitMode(false);
    setOffset({
      x: d.ox + (e.clientX - d.x),
      y: d.oy + (e.clientY - d.y),
    });
  }, []);

  const onPointerUp = useCallback(() => {
    dragRef.current = null;
  }, []);

  const onWheel = useCallback((e: React.WheelEvent) => {
    e.preventDefault();
    const delta = e.deltaY > 0 ? 0.9 : 1.1;
    setFitMode(false);
    setScale((s) => clamp(s * delta, 0.1, 10));
  }, []);

  return {
    scale,
    offset,
    fitMode,
    setZoomPercent,
    zoomTo,
    fit,
    actual,
    onPointerDown,
    onPointerMove,
    onPointerUp,
    onWheel,
  };
}
