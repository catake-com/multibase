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
	export class OpenProtoFileResult {
	    protoFilePath: string;
	    currentDir: string;
	
	    static createFrom(source: any = {}) {
	        return new OpenProtoFileResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.protoFilePath = source["protoFilePath"];
	        this.currentDir = source["currentDir"];
	    }
	}

}

