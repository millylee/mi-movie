export namespace main {
	
	export class Config {
	    proxyServer: string;
	    homePage: string;
	    userData: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.proxyServer = source["proxyServer"];
	        this.homePage = source["homePage"];
	        this.userData = source["userData"];
	    }
	}

}

