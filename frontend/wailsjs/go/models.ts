export namespace imagecore {
	
	export class BitPlaneRequest {
	    imageId: string;
	    channel: string;
	    bit: number;
	
	    static createFrom(source: any = {}) {
	        return new BitPlaneRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.imageId = source["imageId"];
	        this.channel = source["channel"];
	        this.bit = source["bit"];
	    }
	}
	export class HexRow {
	    offset: number;
	    hex: string;
	    ascii: string;
	
	    static createFrom(source: any = {}) {
	        return new HexRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.offset = source["offset"];
	        this.hex = source["hex"];
	        this.ascii = source["ascii"];
	    }
	}
	export class ExtractPreview {
	    rows: HexRow[];
	    totalBytes: number;
	    previewBytes: number;
	    truncated: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ExtractPreview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rows = this.convertValues(source["rows"], HexRow);
	        this.totalBytes = source["totalBytes"];
	        this.previewBytes = source["previewBytes"];
	        this.truncated = source["truncated"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ExtractRequest {
	    imageId: string;
	    maskA: number;
	    maskR: number;
	    maskG: number;
	    maskB: number;
	    channelOrder: string[];
	    traverse: string;
	    bitOrder: string;
	
	    static createFrom(source: any = {}) {
	        return new ExtractRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.imageId = source["imageId"];
	        this.maskA = source["maskA"];
	        this.maskR = source["maskR"];
	        this.maskG = source["maskG"];
	        this.maskB = source["maskB"];
	        this.channelOrder = source["channelOrder"];
	        this.traverse = source["traverse"];
	        this.bitOrder = source["bitOrder"];
	    }
	}
	
	export class ImageInfo {
	    imageId: string;
	    name: string;
	    format: string;
	    width: number;
	    height: number;
	    fileSize: number;
	    hasAlpha: boolean;
	    previewPngB64: string;
	
	    static createFrom(source: any = {}) {
	        return new ImageInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.imageId = source["imageId"];
	        this.name = source["name"];
	        this.format = source["format"];
	        this.width = source["width"];
	        this.height = source["height"];
	        this.fileSize = source["fileSize"];
	        this.hasAlpha = source["hasAlpha"];
	        this.previewPngB64 = source["previewPngB64"];
	    }
	}
	export class SaveResult {
	    cancelled: boolean;
	    path: string;
	    bytes: number;
	
	    static createFrom(source: any = {}) {
	        return new SaveResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cancelled = source["cancelled"];
	        this.path = source["path"];
	        this.bytes = source["bytes"];
	    }
	}

}

