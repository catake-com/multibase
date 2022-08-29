export namespace kafka {
	
	export class StateProject {
	    id: string;
	    currentTab: string;
	    address: string;
	    authMethod: string;
	    authUsername: string;
	    authPassword: string;
	
	    static createFrom(source: any = {}) {
	        return new StateProject(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.currentTab = source["currentTab"];
	        this.address = source["address"];
	        this.authMethod = source["authMethod"];
	        this.authUsername = source["authUsername"];
	        this.authPassword = source["authPassword"];
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

export namespace project {
	
	export class StateProject {
	    id: string;
	    type: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new StateProject(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.name = source["name"];
	    }
	}
	export class StateStats {
	    grpcProjectCount: number;
	    thriftProjectCount: number;
	    kafkaProjectCount: number;
	
	    static createFrom(source: any = {}) {
	        return new StateStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.grpcProjectCount = source["grpcProjectCount"];
	        this.thriftProjectCount = source["thriftProjectCount"];
	        this.kafkaProjectCount = source["kafkaProjectCount"];
	    }
	}
	export class State {
	    // Go type: StateStats
	    stats?: any;
	    projects: {[key: string]: StateProject};
	    openedProjectIDs: string[];
	    currentProjectID: string;
	
	    static createFrom(source: any = {}) {
	        return new State(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.stats = this.convertValues(source["stats"], null);
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
	export class StateProjectFormHeader {
	    id: string;
	    key: string;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new StateProjectFormHeader(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.key = source["key"];
	        this.value = source["value"];
	    }
	}
	export class StateProjectForm {
	    id: string;
	    address: string;
	    headers: StateProjectFormHeader[];
	    selectedMethodID: string;
	    request: string;
	    response: string;
	
	    static createFrom(source: any = {}) {
	        return new StateProjectForm(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.address = source["address"];
	        this.headers = this.convertValues(source["headers"], StateProjectFormHeader);
	        this.selectedMethodID = source["selectedMethodID"];
	        this.request = source["request"];
	        this.response = source["response"];
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
	export class StateProject {
	    id: string;
	    splitterWidth: number;
	    forms: {[key: string]: StateProjectForm};
	    formIDs: string[];
	    currentFormID: string;
	    importPathList: string[];
	    protoFileList: string[];
	    nodes: ProtoTreeNode[];
	
	    static createFrom(source: any = {}) {
	        return new StateProject(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.splitterWidth = source["splitterWidth"];
	        this.forms = source["forms"];
	        this.formIDs = source["formIDs"];
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

export namespace thrift {
	
	export class ServiceTreeNode {
	    id: string;
	    label: string;
	    selectable: boolean;
	    children: ServiceTreeNode[];
	
	    static createFrom(source: any = {}) {
	        return new ServiceTreeNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	        this.selectable = source["selectable"];
	        this.children = this.convertValues(source["children"], ServiceTreeNode);
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
	export class StateProjectFormHeader {
	    id: string;
	    key: string;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new StateProjectFormHeader(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.key = source["key"];
	        this.value = source["value"];
	    }
	}
	export class StateProjectForm {
	    id: string;
	    address: string;
	    headers: StateProjectFormHeader[];
	    selectedFunctionID: string;
	    isMultiplexed: boolean;
	    request: string;
	    response: string;
	
	    static createFrom(source: any = {}) {
	        return new StateProjectForm(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.address = source["address"];
	        this.headers = this.convertValues(source["headers"], StateProjectFormHeader);
	        this.selectedFunctionID = source["selectedFunctionID"];
	        this.isMultiplexed = source["isMultiplexed"];
	        this.request = source["request"];
	        this.response = source["response"];
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
	export class StateProject {
	    id: string;
	    splitterWidth: number;
	    forms: {[key: string]: StateProjectForm};
	    formIDs: string[];
	    currentFormID: string;
	    filePath: string;
	    nodes: ServiceTreeNode[];
	
	    static createFrom(source: any = {}) {
	        return new StateProject(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.splitterWidth = source["splitterWidth"];
	        this.forms = source["forms"];
	        this.formIDs = source["formIDs"];
	        this.currentFormID = source["currentFormID"];
	        this.filePath = source["filePath"];
	        this.nodes = this.convertValues(source["nodes"], ServiceTreeNode);
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

