export namespace internal {
	
	export class Preferences {
	    printer_ip: string;
	    printer_port: string;
	
	    static createFrom(source: any = {}) {
	        return new Preferences(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.printer_ip = source["printer_ip"];
	        this.printer_port = source["printer_port"];
	    }
	}

}

