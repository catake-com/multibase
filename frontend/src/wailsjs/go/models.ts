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
	export class Header {
	    id: string;
	    key: string;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new Header(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.key = source["key"];
	        this.value = source["value"];
	    }
	}
	export class Form {
	    id: string;
	    address: string;
	    headers: Header[];
	    selectedMethodID: string;
	    request: string;
	    response: string;
	
	    static createFrom(source: any = {}) {
	        return new Form(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.address = source["address"];
	        this.headers = this.convertValues(source["headers"], Header);
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
	export class Project {
	    id: string;
	    splitterWidth: number;
	    forms: {[key: string]: Form};
	    formIDs: string[];
	    currentFormID: string;
	    isReflected: boolean;
	    importPathList: string[];
	    protoFileList: string[];
	    nodes: ProtoTreeNode[];
	
	    static createFrom(source: any = {}) {
	        return new Project(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.splitterWidth = source["splitterWidth"];
	        this.forms = source["forms"];
	        this.formIDs = source["formIDs"];
	        this.currentFormID = source["currentFormID"];
	        this.isReflected = source["isReflected"];
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
	export class Header {
	    id: string;
	    key: string;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new Header(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.key = source["key"];
	        this.value = source["value"];
	    }
	}
	export class Form {
	    id: string;
	    address: string;
	    headers: Header[];
	    selectedFunctionID: string;
	    isMultiplexed: boolean;
	    request: string;
	    response: string;
	
	    static createFrom(source: any = {}) {
	        return new Form(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.address = source["address"];
	        this.headers = this.convertValues(source["headers"], Header);
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
	export class Project {
	    id: string;
	    splitterWidth: number;
	    forms: {[key: string]: Form};
	    formIDs: string[];
	    currentFormID: string;
	    filePath: string;
	    nodes: ServiceTreeNode[];
	
	    static createFrom(source: any = {}) {
	        return new Project(source);
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

}

export namespace kafka {
	
	export class Project {
	    id: string;
	    address: string;
	    authMethod: string;
	    authUsername: string;
	    authPassword: string;
	    isConnected: boolean;
	    currentTab: string;
	
	    static createFrom(source: any = {}) {
	        return new Project(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.address = source["address"];
	        this.authMethod = source["authMethod"];
	        this.authUsername = source["authUsername"];
	        this.authPassword = source["authPassword"];
	        this.isConnected = source["isConnected"];
	        this.currentTab = source["currentTab"];
	    }
	}
	export class State {
	    projects: {[key: string]: Project};
	
	    static createFrom(source: any = {}) {
	        return new State(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projects = source["projects"];
	    }
	}
	export class TabConsumersConsumer {
	    name: string;
	    state: string;
	
	    static createFrom(source: any = {}) {
	        return new TabConsumersConsumer(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.state = source["state"];
	    }
	}
	export class TabConsumers {
	    isConnected: boolean;
	    count: number;
	    list: TabConsumersConsumer[];
	
	    static createFrom(source: any = {}) {
	        return new TabConsumers(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isConnected = source["isConnected"];
	        this.count = source["count"];
	        this.list = this.convertValues(source["list"], TabConsumersConsumer);
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
	export class TopicPartition {
	    id: number;
	    offsetTotalStart: number;
	    offsetTotalEnd: number;
	    offsetCurrentStart: number;
	    offsetCurrentEnd: number;
	
	    static createFrom(source: any = {}) {
	        return new TopicPartition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.offsetTotalStart = source["offsetTotalStart"];
	        this.offsetTotalEnd = source["offsetTotalEnd"];
	        this.offsetCurrentStart = source["offsetCurrentStart"];
	        this.offsetCurrentEnd = source["offsetCurrentEnd"];
	    }
	}
	export class TopicOutput {
	    countTotal: number;
	    countCurrent: number;
	    partitions: TopicPartition[];
	
	    static createFrom(source: any = {}) {
	        return new TopicOutput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.countTotal = source["countTotal"];
	        this.countCurrent = source["countCurrent"];
	        this.partitions = this.convertValues(source["partitions"], TopicPartition);
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
	export class TabTopicsTopic {
	    name: string;
	    partitionCount: number;
	    messageCount: number;
	
	    static createFrom(source: any = {}) {
	        return new TabTopicsTopic(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.partitionCount = source["partitionCount"];
	        this.messageCount = source["messageCount"];
	    }
	}
	export class TabTopics {
	    isConnected: boolean;
	    count: number;
	    list: TabTopicsTopic[];
	
	    static createFrom(source: any = {}) {
	        return new TabTopics(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isConnected = source["isConnected"];
	        this.count = source["count"];
	        this.list = this.convertValues(source["list"], TabTopicsTopic);
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
	export class TabBrokersBroker {
	    id: number;
	    rack: string;
	    host: string;
	    port: number;
	
	    static createFrom(source: any = {}) {
	        return new TabBrokersBroker(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.rack = source["rack"];
	        this.host = source["host"];
	        this.port = source["port"];
	    }
	}
	export class TabBrokers {
	    isConnected: boolean;
	    count: number;
	    list: TabBrokersBroker[];
	
	    static createFrom(source: any = {}) {
	        return new TabBrokers(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isConnected = source["isConnected"];
	        this.count = source["count"];
	        this.list = this.convertValues(source["list"], TabBrokersBroker);
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

export namespace project {
	
	export class Project {
	    id: string;
	    type: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new Project(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.name = source["name"];
	    }
	}
	export class Stats {
	    grpcProjectCount: number;
	    thriftProjectCount: number;
	    kafkaProjectCount: number;
	
	    static createFrom(source: any = {}) {
	        return new Stats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.grpcProjectCount = source["grpcProjectCount"];
	        this.thriftProjectCount = source["thriftProjectCount"];
	        this.kafkaProjectCount = source["kafkaProjectCount"];
	    }
	}
	export class Module {
	    // Go type: Stats
	    stats?: any;
	    projects: {[key: string]: Project};
	    openedProjectIDs: string[];
	    currentProjectID: string;
	
	    static createFrom(source: any = {}) {
	        return new Module(source);
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

