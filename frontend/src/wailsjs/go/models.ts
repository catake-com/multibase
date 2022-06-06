export namespace project {
	
	export class StateProject {
	    type: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new StateProject(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.name = source["name"];
	    }
	}
	export class StateStats {
	    _: number;
	
	    static createFrom(source: any = {}) {
	        return new StateStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this._ = source["_"];
	    }
	}
	export class State {
	    // Go type: StateStats
	    _?: any;
	    projects: {[key: string]: StateProject};
	    openedProjectIDs: string[];
	    currentProjectID: string;
	
	    static createFrom(source: any = {}) {
	        return new State(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this._ = this.convertValues(source["_"], null);
	        this.projects = source["projects"];
	        this.openedProjectIDs = source["openedProjectIDs"];
	        this.currentProjectID = source["currentProjectID"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice) {
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

}

export namespace grpc {
	
	export class ProtoTreeNode {
	    id: string;
	    label: string;
	    selectable: boolean;
	    children: ProtoTreeNode[];
	
	    static createFrom(source: any = {}) {
	        return new ProtoTreeNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	        this.selectable = source["selectable"];
	        this.children = this.convertValues(source["children"], ProtoTreeNode);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice) {
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
	export class StateProjectForm {
	    address: string;
	    selectedMethodID: string;
	    request: string;
	    response: string;
	
	    static createFrom(source: any = {}) {
	        return new StateProjectForm(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.address = source["address"];
	        this.selectedMethodID = source["selectedMethodID"];
	        this.request = source["request"];
	        this.response = source["response"];
	    }
	}
	export class StateProject {
	    forms: {[key: string]: StateProjectForm};
	    currentFormID: string;
	    importPathList: string[];
	    protoFileList: string[];
	    nodes: ProtoTreeNode[];
	
	    static createFrom(source: any = {}) {
	        return new StateProject(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.forms = source["forms"];
	        this.currentFormID = source["currentFormID"];
	        this.importPathList = source["importPathList"];
	        this.protoFileList = source["protoFileList"];
	        this.nodes = this.convertValues(source["nodes"], ProtoTreeNode);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice) {
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
	export class State {
	    projects: {[key: string]: StateProject};
	
	    static createFrom(source: any = {}) {
	        return new State(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projects = source["projects"];
	    }
	}

}

