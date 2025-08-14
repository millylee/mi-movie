export namespace main {
	
	export class Config {
	    proxyServer: string;
	    homePage: string;
	    userData: string;
	    userAgent: string;
	    antiDetection: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.proxyServer = source["proxyServer"];
	        this.homePage = source["homePage"];
	        this.userData = source["userData"];
	        this.userAgent = source["userAgent"];
	        this.antiDetection = source["antiDetection"];
	    }
	}

}

