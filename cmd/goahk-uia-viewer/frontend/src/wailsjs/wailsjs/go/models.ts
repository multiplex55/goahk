export namespace inspect {
	
	export class ActivateWindowRequest {
	    hwnd: string;
	
	    static createFrom(source: any = {}) {
	        return new ActivateWindowRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hwnd = source["hwnd"];
	    }
	}
	export class ActivateWindowResponse {
	    activated: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ActivateWindowResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.activated = source["activated"];
	    }
	}
	export class ClearHighlightRequest {
	
	
	    static createFrom(source: any = {}) {
	        return new ClearHighlightRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class ClearHighlightResponse {
	    cleared: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ClearHighlightResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cleared = source["cleared"];
	    }
	}
	export class CopyBestSelectorRequest {
	    nodeID: string;
	
	    static createFrom(source: any = {}) {
	        return new CopyBestSelectorRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.nodeID = source["nodeID"];
	    }
	}
	export class CopyBestSelectorResponse {
	    selector: string;
	    clipboardUpdated: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CopyBestSelectorResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.selector = source["selector"];
	        this.clipboardUpdated = source["clipboardUpdated"];
	    }
	}
	export class GetElementUnderCursorRequest {
	
	
	    static createFrom(source: any = {}) {
	        return new GetElementUnderCursorRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class TreeNodeDTO {
	    nodeID: string;
	    name?: string;
	    controlType?: string;
	    className?: string;
	    hasChildren: boolean;
	    parentNodeID?: string;
	    patterns?: string[];
	    childCount?: number;
	    expanded?: boolean;
	    cycle?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new TreeNodeDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.nodeID = source["nodeID"];
	        this.name = source["name"];
	        this.controlType = source["controlType"];
	        this.className = source["className"];
	        this.hasChildren = source["hasChildren"];
	        this.parentNodeID = source["parentNodeID"];
	        this.patterns = source["patterns"];
	        this.childCount = source["childCount"];
	        this.expanded = source["expanded"];
	        this.cycle = source["cycle"];
	    }
	}
	export class GetElementUnderCursorResponse {
	    element: TreeNodeDTO;
	
	    static createFrom(source: any = {}) {
	        return new GetElementUnderCursorResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.element = this.convertValues(source["element"], TreeNodeDTO);
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
	export class GetFocusedElementRequest {
	
	
	    static createFrom(source: any = {}) {
	        return new GetFocusedElementRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class GetFocusedElementResponse {
	    element: TreeNodeDTO;
	
	    static createFrom(source: any = {}) {
	        return new GetFocusedElementResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.element = this.convertValues(source["element"], TreeNodeDTO);
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
	export class GetNodeChildrenRequest {
	    nodeID: string;
	
	    static createFrom(source: any = {}) {
	        return new GetNodeChildrenRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.nodeID = source["nodeID"];
	    }
	}
	export class GetNodeChildrenResponse {
	    parentNodeID: string;
	    children: TreeNodeDTO[];
	
	    static createFrom(source: any = {}) {
	        return new GetNodeChildrenResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.parentNodeID = source["parentNodeID"];
	        this.children = this.convertValues(source["children"], TreeNodeDTO);
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
	export class GetNodeDetailsRequest {
	    nodeID: string;
	
	    static createFrom(source: any = {}) {
	        return new GetNodeDetailsRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.nodeID = source["nodeID"];
	    }
	}
	export class PatternActionDTO {
	    name: string;
	    payloadSchema?: string;
	
	    static createFrom(source: any = {}) {
	        return new PatternActionDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.payloadSchema = source["payloadSchema"];
	    }
	}
	export class PropertyDTO {
	    name: string;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new PropertyDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.value = source["value"];
	    }
	}
	export class WindowSummary {
	    hwnd: string;
	    title: string;
	    processName?: string;
	    className?: string;
	    processID?: number;
	
	    static createFrom(source: any = {}) {
	        return new WindowSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hwnd = source["hwnd"];
	        this.title = source["title"];
	        this.processName = source["processName"];
	        this.className = source["className"];
	        this.processID = source["processID"];
	    }
	}
	export class GetNodeDetailsResponse {
	    windowInfo: WindowSummary;
	    properties: PropertyDTO[];
	    patterns: PatternActionDTO[];
	    statusText?: string;
	    bestSelector?: string;
	    path?: TreeNodeDTO[];
	
	    static createFrom(source: any = {}) {
	        return new GetNodeDetailsResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.windowInfo = this.convertValues(source["windowInfo"], WindowSummary);
	        this.properties = this.convertValues(source["properties"], PropertyDTO);
	        this.patterns = this.convertValues(source["patterns"], PatternActionDTO);
	        this.statusText = source["statusText"];
	        this.bestSelector = source["bestSelector"];
	        this.path = this.convertValues(source["path"], TreeNodeDTO);
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
	export class GetPatternActionsRequest {
	    nodeID: string;
	
	    static createFrom(source: any = {}) {
	        return new GetPatternActionsRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.nodeID = source["nodeID"];
	    }
	}
	export class GetPatternActionsResponse {
	    nodeID: string;
	    actions: PatternActionDTO[];
	
	    static createFrom(source: any = {}) {
	        return new GetPatternActionsResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.nodeID = source["nodeID"];
	        this.actions = this.convertValues(source["actions"], PatternActionDTO);
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
	export class GetTreeRootRequest {
	    hwnd: string;
	    refresh?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new GetTreeRootRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hwnd = source["hwnd"];
	        this.refresh = source["refresh"];
	    }
	}
	export class GetTreeRootResponse {
	    root: TreeNodeDTO;
	
	    static createFrom(source: any = {}) {
	        return new GetTreeRootResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = this.convertValues(source["root"], TreeNodeDTO);
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
	export class HighlightNodeRequest {
	    nodeID: string;
	
	    static createFrom(source: any = {}) {
	        return new HighlightNodeRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.nodeID = source["nodeID"];
	    }
	}
	export class HighlightNodeResponse {
	    highlighted: boolean;
	
	    static createFrom(source: any = {}) {
	        return new HighlightNodeResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.highlighted = source["highlighted"];
	    }
	}
	export class InspectWindowRequest {
	    hwnd: string;
	
	    static createFrom(source: any = {}) {
	        return new InspectWindowRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hwnd = source["hwnd"];
	    }
	}
	export class InspectWindowResponse {
	    window: WindowSummary;
	    rootNodeID?: string;
	
	    static createFrom(source: any = {}) {
	        return new InspectWindowResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.window = this.convertValues(source["window"], WindowSummary);
	        this.rootNodeID = source["rootNodeID"];
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
	export class InvokePatternRequest {
	    nodeID: string;
	    action: string;
	    payload?: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new InvokePatternRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.nodeID = source["nodeID"];
	        this.action = source["action"];
	        this.payload = source["payload"];
	    }
	}
	export class InvokePatternResponse {
	    nodeID: string;
	    action: string;
	    invoked: boolean;
	    result?: string;
	
	    static createFrom(source: any = {}) {
	        return new InvokePatternResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.nodeID = source["nodeID"];
	        this.action = source["action"];
	        this.invoked = source["invoked"];
	        this.result = source["result"];
	    }
	}
	export class ListWindowsRequest {
	    titleContains?: string;
	    className?: string;
	
	    static createFrom(source: any = {}) {
	        return new ListWindowsRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.titleContains = source["titleContains"];
	        this.className = source["className"];
	    }
	}
	export class ListWindowsResponse {
	    windows: WindowSummary[];
	
	    static createFrom(source: any = {}) {
	        return new ListWindowsResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.windows = this.convertValues(source["windows"], WindowSummary);
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
	
	
	export class RefreshWindowsRequest {
	    filter?: string;
	    visibleOnly: boolean;
	    titleOnly: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RefreshWindowsRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filter = source["filter"];
	        this.visibleOnly = source["visibleOnly"];
	        this.titleOnly = source["titleOnly"];
	    }
	}
	export class RefreshWindowsResponse {
	    windows: WindowSummary[];
	
	    static createFrom(source: any = {}) {
	        return new RefreshWindowsResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.windows = this.convertValues(source["windows"], WindowSummary);
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
	export class SelectNodeRequest {
	    nodeID: string;
	
	    static createFrom(source: any = {}) {
	        return new SelectNodeRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.nodeID = source["nodeID"];
	    }
	}
	export class SelectNodeResponse {
	    selected: TreeNodeDTO;
	
	    static createFrom(source: any = {}) {
	        return new SelectNodeResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.selected = this.convertValues(source["selected"], TreeNodeDTO);
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
	export class ToggleFollowCursorRequest {
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ToggleFollowCursorRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	    }
	}
	export class ToggleFollowCursorResponse {
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ToggleFollowCursorResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	    }
	}
	

}

