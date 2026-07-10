export type Channel = 'A' | 'R' | 'G' | 'B';
export type TraverseOrder = 'row' | 'col';
export type BitOrder = 'lsb' | 'msb';

export interface ImageInfo {
  imageId: string;
  name: string;
  format: string;
  width: number;
  height: number;
  fileSize: number;
  hasAlpha: boolean;
  previewPngB64: string;
}

export interface BitPlaneRequest {
  imageId: string;
  channel: Channel | 'ORIG' | '';
  bit: number;
}

export interface ExtractRequest {
  imageId: string;
  maskA: number;
  maskR: number;
  maskG: number;
  maskB: number;
  channelOrder: Channel[];
  traverse: TraverseOrder;
  bitOrder: BitOrder;
}

export interface HexRow {
  offset: number;
  hex: string;
  ascii: string;
}

export interface ExtractPreview {
  rows: HexRow[];
  totalBytes: number;
  previewBytes: number;
  truncated: boolean;
}

export interface SaveResult {
  cancelled: boolean;
  path: string;
  bytes: number;
}

export type ViewMode = 'original' | 'bitplane' | 'lsb';

export interface PlaneSelection {
  kind: 'original' | 'plane';
  channel: Channel;
  bit: number;
}
