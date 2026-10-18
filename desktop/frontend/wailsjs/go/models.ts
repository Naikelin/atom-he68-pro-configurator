export namespace hid {
	
	export class Mapping {
	    addr: number;
	    label: string;
	    hidName: string;
	    hidUsage: number;
	
	    static createFrom(source: any = {}) {
	        return new Mapping(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.addr = source["addr"];
	        this.label = source["label"];
	        this.hidName = source["hidName"];
	        this.hidUsage = source["hidUsage"];
	    }
	}
	export class PhysicalKey {
	    label: string;
	    addr: number;
	    x: number;
	    y: number;
	    w: number;
	    h: number;
	
	    static createFrom(source: any = {}) {
	        return new PhysicalKey(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.addr = source["addr"];
	        this.x = source["x"];
	        this.y = source["y"];
	        this.w = source["w"];
	        this.h = source["h"];
	    }
	}

}

