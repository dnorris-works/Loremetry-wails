export namespace cloud {
	
	export class Account {
	    plan: string;
	    credits: number;
	    remaining: string;
	
	    static createFrom(source: any = {}) {
	        return new Account(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.plan = source["plan"];
	        this.credits = source["credits"];
	        this.remaining = source["remaining"];
	    }
	}
	export class JobStatus {
	    job_id: string;
	    status: string;
	    step: number;
	    step_total: number;
	    body: string;
	    error: string;
	    credits: number;
	    cached: boolean;
	
	    static createFrom(source: any = {}) {
	        return new JobStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.job_id = source["job_id"];
	        this.status = source["status"];
	        this.step = source["step"];
	        this.step_total = source["step_total"];
	        this.body = source["body"];
	        this.error = source["error"];
	        this.credits = source["credits"];
	        this.cached = source["cached"];
	    }
	}

}

export namespace main {
	
	export class ImportedFile {
	    name: string;
	    text: string;
	    rel: string;
	
	    static createFrom(source: any = {}) {
	        return new ImportedFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.text = source["text"];
	        this.rel = source["rel"];
	    }
	}

}

export namespace store {
	
	export class ActInput {
	    title: string;
	    summary: string;
	
	    static createFrom(source: any = {}) {
	        return new ActInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.summary = source["summary"];
	    }
	}
	export class AdminColumn {
	    name: string;
	    type: string;
	    nullable: string;
	    default?: string;
	
	    static createFrom(source: any = {}) {
	        return new AdminColumn(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.nullable = source["nullable"];
	        this.default = source["default"];
	    }
	}
	export class AdminSQLResult {
	    columns: string[];
	    rows: string[][];
	    rows_affected: number;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new AdminSQLResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.columns = source["columns"];
	        this.rows = source["rows"];
	        this.rows_affected = source["rows_affected"];
	        this.message = source["message"];
	    }
	}
	export class AdminSchema {
	    columns: AdminColumn[];
	
	    static createFrom(source: any = {}) {
	        return new AdminSchema(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.columns = this.convertValues(source["columns"], AdminColumn);
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
	export class AdminTableRows {
	    columns: string[];
	    rows: string[][];
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new AdminTableRows(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.columns = source["columns"];
	        this.rows = source["rows"];
	        this.total = source["total"];
	    }
	}
	export class AdminTables {
	    tables: string[];
	
	    static createFrom(source: any = {}) {
	        return new AdminTables(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tables = source["tables"];
	    }
	}
	export class AnalysisDetail {
	    id: string;
	    label: string;
	    group: string;
	    description: string;
	    needs: string[];
	    depends_on: string[];
	    uses_ai: boolean;
	    uses_merge: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AnalysisDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	        this.group = source["group"];
	        this.description = source["description"];
	        this.needs = source["needs"];
	        this.depends_on = source["depends_on"];
	        this.uses_ai = source["uses_ai"];
	        this.uses_merge = source["uses_merge"];
	    }
	}
	export class AnalysisFile {
	    name: string;
	    path: string;
	    rel: string;
	
	    static createFrom(source: any = {}) {
	        return new AnalysisFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.rel = source["rel"];
	    }
	}
	export class AnalysisItem {
	    id: string;
	    label: string;
	    description: string;
	    needs: string[];
	    depends_on: string[];
	    uses_ai: boolean;
	    uses_merge: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AnalysisItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	        this.description = source["description"];
	        this.needs = source["needs"];
	        this.depends_on = source["depends_on"];
	        this.uses_ai = source["uses_ai"];
	        this.uses_merge = source["uses_merge"];
	    }
	}
	export class AnalysisGroup {
	    id: string;
	    label: string;
	    items: AnalysisItem[];
	
	    static createFrom(source: any = {}) {
	        return new AnalysisGroup(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	        this.items = this.convertValues(source["items"], AnalysisItem);
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
	
	export class AnalysisReport {
	    id: number;
	    analysis_id: string;
	    analysis_label: string;
	    project_path: string;
	    uses_ai: boolean;
	    body?: string;
	    original?: string;
	    proposed?: string;
	    body_size?: number;
	    created_at: string;
	
	    static createFrom(source: any = {}) {
	        return new AnalysisReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.analysis_id = source["analysis_id"];
	        this.analysis_label = source["analysis_label"];
	        this.project_path = source["project_path"];
	        this.uses_ai = source["uses_ai"];
	        this.body = source["body"];
	        this.original = source["original"];
	        this.proposed = source["proposed"];
	        this.body_size = source["body_size"];
	        this.created_at = source["created_at"];
	    }
	}
	export class AnalysisReportSummary {
	    id: number;
	    analysis_id: string;
	    analysis_label: string;
	    project_path: string;
	    uses_ai: boolean;
	    body_size: number;
	    created_at: string;
	
	    static createFrom(source: any = {}) {
	        return new AnalysisReportSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.analysis_id = source["analysis_id"];
	        this.analysis_label = source["analysis_label"];
	        this.project_path = source["project_path"];
	        this.uses_ai = source["uses_ai"];
	        this.body_size = source["body_size"];
	        this.created_at = source["created_at"];
	    }
	}
	export class AnalysisRoleMatch {
	    role: string;
	    rel: string;
	    path: string;
	    present: boolean;
	    files: AnalysisFile[];
	
	    static createFrom(source: any = {}) {
	        return new AnalysisRoleMatch(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.role = source["role"];
	        this.rel = source["rel"];
	        this.path = source["path"];
	        this.present = source["present"];
	        this.files = this.convertValues(source["files"], AnalysisFile);
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
	export class AnalysisSources {
	    project_path: string;
	    kind: string;
	    roles: AnalysisRoleMatch[];
	
	    static createFrom(source: any = {}) {
	        return new AnalysisSources(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.project_path = source["project_path"];
	        this.kind = source["kind"];
	        this.roles = this.convertValues(source["roles"], AnalysisRoleMatch);
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
	export class AuthSession {
	    authenticated: boolean;
	    id?: string;
	    email?: string;
	    firstName?: string;
	    lastName?: string;
	    isAdmin?: boolean;
	    breakGlass?: boolean;
	    reason?: string;
	
	    static createFrom(source: any = {}) {
	        return new AuthSession(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.authenticated = source["authenticated"];
	        this.id = source["id"];
	        this.email = source["email"];
	        this.firstName = source["firstName"];
	        this.lastName = source["lastName"];
	        this.isAdmin = source["isAdmin"];
	        this.breakGlass = source["breakGlass"];
	        this.reason = source["reason"];
	    }
	}
	export class BibleDoc {
	    id: number;
	    series_id: number;
	    file_name: string;
	    text_content: string;
	
	    static createFrom(source: any = {}) {
	        return new BibleDoc(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.series_id = source["series_id"];
	        this.file_name = source["file_name"];
	        this.text_content = source["text_content"];
	    }
	}
	export class Chapter {
	    id: number;
	    story_id: number;
	    file_name: string;
	    text_content?: string;
	    sort_order?: number;
	    act_id?: number;
	    created_at: string;
	
	    static createFrom(source: any = {}) {
	        return new Chapter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.story_id = source["story_id"];
	        this.file_name = source["file_name"];
	        this.text_content = source["text_content"];
	        this.sort_order = source["sort_order"];
	        this.act_id = source["act_id"];
	        this.created_at = source["created_at"];
	    }
	}
	export class ChapterInput {
	    file_name?: string;
	    text_content?: string;
	    sort_order?: number;
	    act_id?: number;
	
	    static createFrom(source: any = {}) {
	        return new ChapterInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.file_name = source["file_name"];
	        this.text_content = source["text_content"];
	        this.sort_order = source["sort_order"];
	        this.act_id = source["act_id"];
	    }
	}
	export class CharacterInput {
	    series_id: number;
	    story_id: number;
	    character_name: string;
	    role_in_story: string;
	    story_importance: string;
	    pov_character: boolean;
	    character_want: string;
	    character_need: string;
	    arc_status_at_start: string;
	    arc_status_at_end: string;
	    key_relationships: string[];
	    defining_trait: string;
	    first_appearance_book: number;
	    primary_obstacle: string;
	    internal_vs_external: string;
	    core_wound: string;
	    age: number;
	    thematic_resonance: string;
	    character_voice_notes: string;
	
	    static createFrom(source: any = {}) {
	        return new CharacterInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.series_id = source["series_id"];
	        this.story_id = source["story_id"];
	        this.character_name = source["character_name"];
	        this.role_in_story = source["role_in_story"];
	        this.story_importance = source["story_importance"];
	        this.pov_character = source["pov_character"];
	        this.character_want = source["character_want"];
	        this.character_need = source["character_need"];
	        this.arc_status_at_start = source["arc_status_at_start"];
	        this.arc_status_at_end = source["arc_status_at_end"];
	        this.key_relationships = source["key_relationships"];
	        this.defining_trait = source["defining_trait"];
	        this.first_appearance_book = source["first_appearance_book"];
	        this.primary_obstacle = source["primary_obstacle"];
	        this.internal_vs_external = source["internal_vs_external"];
	        this.core_wound = source["core_wound"];
	        this.age = source["age"];
	        this.thematic_resonance = source["thematic_resonance"];
	        this.character_voice_notes = source["character_voice_notes"];
	    }
	}
	export class CharacterProfile {
	    id: number;
	    series_id?: number;
	    story_id?: number;
	    character_name: string;
	    role_in_story?: string;
	    story_importance?: string;
	    pov_character: boolean;
	    character_want?: string;
	    character_need?: string;
	    arc_status_at_start?: string;
	    arc_status_at_end?: string;
	    key_relationships?: string[];
	    defining_trait?: string;
	    first_appearance_book?: number;
	    primary_obstacle?: string;
	    internal_vs_external?: string;
	    core_wound?: string;
	    age?: number;
	    thematic_resonance?: string;
	    character_voice_notes?: string;
	    created_at: string;
	    updated_at: string;
	
	    static createFrom(source: any = {}) {
	        return new CharacterProfile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.series_id = source["series_id"];
	        this.story_id = source["story_id"];
	        this.character_name = source["character_name"];
	        this.role_in_story = source["role_in_story"];
	        this.story_importance = source["story_importance"];
	        this.pov_character = source["pov_character"];
	        this.character_want = source["character_want"];
	        this.character_need = source["character_need"];
	        this.arc_status_at_start = source["arc_status_at_start"];
	        this.arc_status_at_end = source["arc_status_at_end"];
	        this.key_relationships = source["key_relationships"];
	        this.defining_trait = source["defining_trait"];
	        this.first_appearance_book = source["first_appearance_book"];
	        this.primary_obstacle = source["primary_obstacle"];
	        this.internal_vs_external = source["internal_vs_external"];
	        this.core_wound = source["core_wound"];
	        this.age = source["age"];
	        this.thematic_resonance = source["thematic_resonance"];
	        this.character_voice_notes = source["character_voice_notes"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}
	export class DeletedResult {
	    deleted: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DeletedResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.deleted = source["deleted"];
	    }
	}
	export class DirFile {
	    name: string;
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new DirFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	    }
	}
	export class DirNode {
	    name: string;
	    path: string;
	    key: string;
	    label: string;
	    count: number;
	    hidden: boolean;
	    required: boolean;
	    files: DirFile[];
	    folders: DirNode[];
	
	    static createFrom(source: any = {}) {
	        return new DirNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.key = source["key"];
	        this.label = source["label"];
	        this.count = source["count"];
	        this.hidden = source["hidden"];
	        this.required = source["required"];
	        this.files = this.convertValues(source["files"], DirFile);
	        this.folders = this.convertValues(source["folders"], DirNode);
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
	export class DiskFile {
	    dir: string;
	    name: string;
	    path: string;
	    text?: string;
	
	    static createFrom(source: any = {}) {
	        return new DiskFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dir = source["dir"];
	        this.name = source["name"];
	        this.path = source["path"];
	        this.text = source["text"];
	    }
	}
	export class DiskFileWrite {
	    dir: string;
	    name: string;
	    new_name: string;
	    text: string;
	
	    static createFrom(source: any = {}) {
	        return new DiskFileWrite(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dir = source["dir"];
	        this.name = source["name"];
	        this.new_name = source["new_name"];
	        this.text = source["text"];
	    }
	}
	export class DocInput {
	    kind: string;
	    file_name: string;
	    text_content: string;
	
	    static createFrom(source: any = {}) {
	        return new DocInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.file_name = source["file_name"];
	        this.text_content = source["text_content"];
	    }
	}
	export class DocumentType {
	    code: string;
	    display_name: string;
	    sort_order: number;
	    applies_to_story: boolean;
	    applies_to_series: boolean;
	    active: boolean;
	    editor: string;
	    storage: string;
	
	    static createFrom(source: any = {}) {
	        return new DocumentType(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.display_name = source["display_name"];
	        this.sort_order = source["sort_order"];
	        this.applies_to_story = source["applies_to_story"];
	        this.applies_to_series = source["applies_to_series"];
	        this.active = source["active"];
	        this.editor = source["editor"];
	        this.storage = source["storage"];
	    }
	}
	export class FileRange {
	    dir: string;
	    name: string;
	    text: string;
	    start: number;
	    end: number;
	    file_size: number;
	
	    static createFrom(source: any = {}) {
	        return new FileRange(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dir = source["dir"];
	        this.name = source["name"];
	        this.text = source["text"];
	        this.start = source["start"];
	        this.end = source["end"];
	        this.file_size = source["file_size"];
	    }
	}
	export class FileSection {
	    index: number;
	    title: string;
	
	    static createFrom(source: any = {}) {
	        return new FileSection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.title = source["title"];
	    }
	}
	export class FileSectionContent {
	    dir: string;
	    name: string;
	    index: number;
	    title: string;
	    text: string;
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new FileSectionContent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dir = source["dir"];
	        this.name = source["name"];
	        this.index = source["index"];
	        this.title = source["title"];
	        this.text = source["text"];
	        this.total = source["total"];
	    }
	}
	export class FileSections {
	    dir: string;
	    name: string;
	    sections: FileSection[];
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new FileSections(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dir = source["dir"];
	        this.name = source["name"];
	        this.sections = this.convertValues(source["sections"], FileSection);
	        this.total = source["total"];
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
	export class FolderChange {
	    type: string;
	    id: number;
	
	    static createFrom(source: any = {}) {
	        return new FolderChange(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.id = source["id"];
	    }
	}
	export class FolderLink {
	    scope: string;
	    owner_id: number;
	    kind: string;
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new FolderLink(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.scope = source["scope"];
	        this.owner_id = source["owner_id"];
	        this.kind = source["kind"];
	        this.path = source["path"];
	    }
	}
	export class FolderLinkInput {
	    scope: string;
	    owner_id: number;
	    kind: string;
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new FolderLinkInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.scope = source["scope"];
	        this.owner_id = source["owner_id"];
	        this.kind = source["kind"];
	        this.path = source["path"];
	    }
	}
	export class FolderOverrideInput {
	    project_path: string;
	    key: string;
	    mode: string;
	
	    static createFrom(source: any = {}) {
	        return new FolderOverrideInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.project_path = source["project_path"];
	        this.key = source["key"];
	        this.mode = source["mode"];
	    }
	}
	export class FolderVisibility {
	    hidden_names: string[];
	    show_hidden: boolean;
	    overrides: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new FolderVisibility(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hidden_names = source["hidden_names"];
	        this.show_hidden = source["show_hidden"];
	        this.overrides = source["overrides"];
	    }
	}
	export class HeaderFile {
	    rel: string;
	    name: string;
	    folder?: string;
	
	    static createFrom(source: any = {}) {
	        return new HeaderFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rel = source["rel"];
	        this.name = source["name"];
	        this.folder = source["folder"];
	    }
	}
	export class HeaderFileContent {
	    rel: string;
	    name: string;
	    text: string;
	
	    static createFrom(source: any = {}) {
	        return new HeaderFileContent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rel = source["rel"];
	        this.name = source["name"];
	        this.text = source["text"];
	    }
	}
	export class HeaderFileRef {
	    project_path: string;
	    project_kind: string;
	    kind: string;
	    rel: string;
	
	    static createFrom(source: any = {}) {
	        return new HeaderFileRef(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.project_path = source["project_path"];
	        this.project_kind = source["project_kind"];
	        this.kind = source["kind"];
	        this.rel = source["rel"];
	    }
	}
	export class HeaderFileWrite {
	    project_path: string;
	    project_kind: string;
	    kind: string;
	    rel: string;
	    name: string;
	    text: string;
	
	    static createFrom(source: any = {}) {
	        return new HeaderFileWrite(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.project_path = source["project_path"];
	        this.project_kind = source["project_kind"];
	        this.kind = source["kind"];
	        this.rel = source["rel"];
	        this.name = source["name"];
	        this.text = source["text"];
	    }
	}
	export class HeaderList {
	    folder: string;
	    missing: boolean;
	    files: HeaderFile[];
	
	    static createFrom(source: any = {}) {
	        return new HeaderList(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.folder = source["folder"];
	        this.missing = source["missing"];
	        this.files = this.convertValues(source["files"], HeaderFile);
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
	export class HeaderPlaceInput {
	    project_path: string;
	    project_kind: string;
	    kind: string;
	    rels: string[];
	
	    static createFrom(source: any = {}) {
	        return new HeaderPlaceInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.project_path = source["project_path"];
	        this.project_kind = source["project_kind"];
	        this.kind = source["kind"];
	        this.rels = source["rels"];
	    }
	}
	export class IDList {
	    ids: number[];
	
	    static createFrom(source: any = {}) {
	        return new IDList(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ids = source["ids"];
	    }
	}
	export class IDResult {
	    id: number;
	
	    static createFrom(source: any = {}) {
	        return new IDResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	    }
	}
	export class IncomingFile {
	    name: string;
	    text: string;
	
	    static createFrom(source: any = {}) {
	        return new IncomingFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.text = source["text"];
	    }
	}
	export class NameList {
	    names: string[];
	
	    static createFrom(source: any = {}) {
	        return new NameList(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.names = source["names"];
	    }
	}
	export class OpenedFile {
	    dir: string;
	    name: string;
	    text?: string;
	    large: boolean;
	    file_size?: number;
	
	    static createFrom(source: any = {}) {
	        return new OpenedFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dir = source["dir"];
	        this.name = source["name"];
	        this.text = source["text"];
	        this.large = source["large"];
	        this.file_size = source["file_size"];
	    }
	}
	export class PathResult {
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new PathResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	    }
	}
	export class Pen {
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new Pen(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	    }
	}
	export class PlaceInput {
	    kind: string;
	    ids: number[];
	    story_id: number;
	
	    static createFrom(source: any = {}) {
	        return new PlaceInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.ids = source["ids"];
	        this.story_id = source["story_id"];
	    }
	}
	export class RenameCount {
	    section: string;
	    renamed: number;
	
	    static createFrom(source: any = {}) {
	        return new RenameCount(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.section = source["section"];
	        this.renamed = source["renamed"];
	    }
	}
	export class Series {
	    id: number;
	    name: string;
	    description?: string;
	    summary?: string;
	    pen_name: string;
	    created_at: string;
	    updated_at: string;
	
	    static createFrom(source: any = {}) {
	        return new Series(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.summary = source["summary"];
	        this.pen_name = source["pen_name"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}
	export class SeriesDoc {
	    id: number;
	    series_id: number;
	    category: string;
	    file_name: string;
	    created_at: string;
	
	    static createFrom(source: any = {}) {
	        return new SeriesDoc(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.series_id = source["series_id"];
	        this.category = source["category"];
	        this.file_name = source["file_name"];
	        this.created_at = source["created_at"];
	    }
	}
	export class SeriesDocInput {
	    category: string;
	    file_name: string;
	    text_content: string;
	
	    static createFrom(source: any = {}) {
	        return new SeriesDocInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.category = source["category"];
	        this.file_name = source["file_name"];
	        this.text_content = source["text_content"];
	    }
	}
	export class SeriesInput {
	    name: string;
	    description: string;
	    summary: string;
	    pen_name: string;
	
	    static createFrom(source: any = {}) {
	        return new SeriesInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.summary = source["summary"];
	        this.pen_name = source["pen_name"];
	    }
	}
	export class SettingValue {
	    key: string;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new SettingValue(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.value = source["value"];
	    }
	}
	export class Story {
	    id: number;
	    series_id?: number;
	    series_sort_order?: number;
	    name: string;
	    description?: string;
	    pen_name: string;
	    created_at: string;
	    updated_at: string;
	
	    static createFrom(source: any = {}) {
	        return new Story(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.series_id = source["series_id"];
	        this.series_sort_order = source["series_sort_order"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.pen_name = source["pen_name"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}
	export class StoryAct {
	    id: number;
	    story_id: number;
	    title: string;
	    summary?: string;
	    sort_order?: number;
	    created_at: string;
	    updated_at: string;
	
	    static createFrom(source: any = {}) {
	        return new StoryAct(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.story_id = source["story_id"];
	        this.title = source["title"];
	        this.summary = source["summary"];
	        this.sort_order = source["sort_order"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	    }
	}
	export class StoryDoc {
	    id: number;
	    story_id: number;
	    kind: string;
	    file_name: string;
	    sort_order?: number;
	    act_id?: number;
	    created_at: string;
	
	    static createFrom(source: any = {}) {
	        return new StoryDoc(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.story_id = source["story_id"];
	        this.kind = source["kind"];
	        this.file_name = source["file_name"];
	        this.sort_order = source["sort_order"];
	        this.act_id = source["act_id"];
	        this.created_at = source["created_at"];
	    }
	}
	export class StoryInput {
	    name: string;
	    description: string;
	    series_id: number;
	    pen_name: string;
	
	    static createFrom(source: any = {}) {
	        return new StoryInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.series_id = source["series_id"];
	        this.pen_name = source["pen_name"];
	    }
	}
	export class TemplateFolderLists {
	    series: string[];
	    books: string[];
	
	    static createFrom(source: any = {}) {
	        return new TemplateFolderLists(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.series = source["series"];
	        this.books = source["books"];
	    }
	}
	export class TreeBook {
	    path: string;
	    name: string;
	    folder: string;
	    tree: DirNode;
	
	    static createFrom(source: any = {}) {
	        return new TreeBook(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.folder = source["folder"];
	        this.tree = this.convertValues(source["tree"], DirNode);
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
	export class TreeProblem {
	    path: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new TreeProblem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.message = source["message"];
	    }
	}
	export class TreeSeries {
	    path: string;
	    name: string;
	    tree: DirNode;
	    books: TreeBook[];
	    problems: TreeProblem[];
	
	    static createFrom(source: any = {}) {
	        return new TreeSeries(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.tree = this.convertValues(source["tree"], DirNode);
	        this.books = this.convertValues(source["books"], TreeBook);
	        this.problems = this.convertValues(source["problems"], TreeProblem);
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
	export class TreePen {
	    name: string;
	    path: string;
	    series: TreeSeries[];
	    books: TreeBook[];
	    problems: TreeProblem[];
	
	    static createFrom(source: any = {}) {
	        return new TreePen(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	        this.series = this.convertValues(source["series"], TreeSeries);
	        this.books = this.convertValues(source["books"], TreeBook);
	        this.problems = this.convertValues(source["problems"], TreeProblem);
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
	
	
	export class UIFileSelection {
	    type: string;
	    dir: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new UIFileSelection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.dir = source["dir"];
	        this.name = source["name"];
	    }
	}
	export class UISession {
	    selection: UIFileSelection;
	    open_series: string[];
	    open_stories: string[];
	    open_folders: string[];
	    restore_open: boolean;
	    last_pen: string;
	
	    static createFrom(source: any = {}) {
	        return new UISession(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.selection = this.convertValues(source["selection"], UIFileSelection);
	        this.open_series = source["open_series"];
	        this.open_stories = source["open_stories"];
	        this.open_folders = source["open_folders"];
	        this.restore_open = source["restore_open"];
	        this.last_pen = source["last_pen"];
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
	export class UpdatedResult {
	    updated: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UpdatedResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.updated = source["updated"];
	    }
	}
	export class WritingProjectInput {
	    pen_name: string;
	    pen_path: string;
	    name: string;
	    series_path: string;
	
	    static createFrom(source: any = {}) {
	        return new WritingProjectInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pen_name = source["pen_name"];
	        this.pen_path = source["pen_path"];
	        this.name = source["name"];
	        this.series_path = source["series_path"];
	    }
	}
	export class WritingTree {
	    root: string;
	    pens: TreePen[];
	    problems: TreeProblem[];
	
	    static createFrom(source: any = {}) {
	        return new WritingTree(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.root = source["root"];
	        this.pens = this.convertValues(source["pens"], TreePen);
	        this.problems = this.convertValues(source["problems"], TreeProblem);
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

}

