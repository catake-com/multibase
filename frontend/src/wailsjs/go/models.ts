export namespace grpc {
	
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

export namespace kafka {
	
	export class State {
	    id: string;
	    address: string;
	    authMethod: string;
	    authUsername: string;
	    authPassword: string;
	    isConnected: boolean;
	    currentTab: string;
	
	    static createFrom(source: any = {}) {
	        return new State(source);
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
	export class TabBrokersDataBroker {
	    id: number;
	    rack: string;
	    host: string;
	    port: number;
	
	    static createFrom(source: any = {}) {
	        return new TabBrokersDataBroker(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.rack = source["rack"];
	        this.host = source["host"];
	        this.port = source["port"];
	    }
	}
	export class TabBrokersData {
	    isConnected: boolean;
	    count: number;
	    list: TabBrokersDataBroker[];
	
	    static createFrom(source: any = {}) {
	        return new TabBrokersData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isConnected = source["isConnected"];
	        this.count = source["count"];
	        this.list = this.convertValues(source["list"], TabBrokersDataBroker);
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
	export class TabConsumersDataConsumer {
	    name: string;
	    state: string;
	
	    static createFrom(source: any = {}) {
	        return new TabConsumersDataConsumer(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.state = source["state"];
	    }
	}
	export class TabConsumersData {
	    isConnected: boolean;
	    count: number;
	    list: TabConsumersDataConsumer[];
	
	    static createFrom(source: any = {}) {
	        return new TabConsumersData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isConnected = source["isConnected"];
	        this.count = source["count"];
	        this.list = this.convertValues(source["list"], TabConsumersDataConsumer);
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
	export class TabTopicsDataTopic {
	    name: string;
	    partitionCount: number;
	    messageCount: number;
	
	    static createFrom(source: any = {}) {
	        return new TabTopicsDataTopic(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.partitionCount = source["partitionCount"];
	        this.messageCount = source["messageCount"];
	    }
	}
	export class TabTopicsData {
	    isConnected: boolean;
	    count: number;
	    list: TabTopicsDataTopic[];
	
	    static createFrom(source: any = {}) {
	        return new TabTopicsData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isConnected = source["isConnected"];
	        this.count = source["count"];
	        this.list = this.convertValues(source["list"], TabTopicsDataTopic);
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
	    topicName: string;
	    startFromTime: string;
	    countTotal: number;
	    countCurrent: number;
	    partitions: TopicPartition[];
	
	    static createFrom(source: any = {}) {
	        return new TopicOutput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.topicName = source["topicName"];
	        this.startFromTime = source["startFromTime"];
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

}

export namespace kubernetes {
	
	export class State {
	    id: string;
	    selectedContext: string;
	    selectedNamespace: string;
	    isConnected: boolean;
	    isPortForwarded: boolean;
	    currentTab: string;
	
	    static createFrom(source: any = {}) {
	        return new State(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.selectedContext = source["selectedContext"];
	        this.selectedNamespace = source["selectedNamespace"];
	        this.isConnected = source["isConnected"];
	        this.isPortForwarded = source["isPortForwarded"];
	        this.currentTab = source["currentTab"];
	    }
	}
	export class TabOverviewDataContext {
	    isSelected: boolean;
	    name: string;
	    cluster: string;
	
	    static createFrom(source: any = {}) {
	        return new TabOverviewDataContext(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isSelected = source["isSelected"];
	        this.name = source["name"];
	        this.cluster = source["cluster"];
	    }
	}
	export class TabOverviewData {
	    isConnected: boolean;
	    contexts: TabOverviewDataContext[];
	
	    static createFrom(source: any = {}) {
	        return new TabOverviewData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isConnected = source["isConnected"];
	        this.contexts = this.convertValues(source["contexts"], TabOverviewDataContext);
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
	export class TabWorkloadsPodsDataPodPort {
	    name: string;
	    containerPort: number;
	
	    static createFrom(source: any = {}) {
	        return new TabWorkloadsPodsDataPodPort(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.containerPort = source["containerPort"];
	    }
	}
	export class TabWorkloadsPodsDataPod {
	    name: string;
	    namespace: string;
	    ports: TabWorkloadsPodsDataPodPort[];
	
	    static createFrom(source: any = {}) {
	        return new TabWorkloadsPodsDataPod(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.namespace = source["namespace"];
	        this.ports = this.convertValues(source["ports"], TabWorkloadsPodsDataPodPort);
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
	export class TabWorkloadsPodsData {
	    pods: TabWorkloadsPodsDataPod[];
	
	    static createFrom(source: any = {}) {
	        return new TabWorkloadsPodsData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pods = this.convertValues(source["pods"], TabWorkloadsPodsDataPod);
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
	    kubernetesProjectCount: number;
	
	    static createFrom(source: any = {}) {
	        return new Stats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.grpcProjectCount = source["grpcProjectCount"];
	        this.thriftProjectCount = source["thriftProjectCount"];
	        this.kafkaProjectCount = source["kafkaProjectCount"];
	        this.kubernetesProjectCount = source["kubernetesProjectCount"];
	    }
	}
	export class Module {
	    stats?: Stats;
	    projects: {[key: string]: Project};
	    openedProjectIDs: string[];
	    currentProjectID: string;
	
	    static createFrom(source: any = {}) {
	        return new Module(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.stats = this.convertValues(source["stats"], Stats);
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

export namespace thrift {
	
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

