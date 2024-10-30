import { Ref, ref } from 'vue';
import { IRowNode } from 'ag-grid-community';

export const EXPANDABLE_DATA_COLUMN = 'payload';
export const EXPANDABLE_DATA_COLUMN_SHADOW = 'payloadString';
export const COLLAPSABLE_DATA_COLUMN = 'collapsedData';
export const COLLAPSABLE_DATA_COLUMN_SHADOW = 'collapsedDataString';

export function makeExpandableDataColumnKey(key: string): string {
    return `${EXPANDABLE_DATA_COLUMN}.${key}`;
}

export function isExpandableDataColumnKey(key: string): boolean {
    return key.startsWith(EXPANDABLE_DATA_COLUMN + '.');
}

export function getExpandableDataColumnSubKey(key: string): string {
    return key.substring(EXPANDABLE_DATA_COLUMN.length + 1);
}

export const CUSTOMARY_COLUMN_KEYS = ref(new Set<string>([
    'id',
    'country',
    'createdAt',

    'time',
    'topic',
    'category',
    'entry',
    /* 'price', */
]));

export const SPECIAL_COLUMN_KEYS = new Set([
    EXPANDABLE_DATA_COLUMN,
    // derived columns
    EXPANDABLE_DATA_COLUMN_SHADOW,
    COLLAPSABLE_DATA_COLUMN,
    COLLAPSABLE_DATA_COLUMN_SHADOW,
]);

export interface DispalyableColumn {
    key: string;
    isEnabled: boolean;
}


export interface ColumnKey {
    effectiveLookupPath: string[];  // the actual code-friendly lookup path
    apparentLookupPath: string;  // the that the human uses to refer to the data
    displayString: string;
    shouldDisplay: boolean;
}

export namespace UsableColumns {
    export const columnsSet = ref<Set<ColumnKey>>(new Set());

    const stringifiedLookup = new Map<string, ColumnKey>();

    export function reset() {
        stringifiedLookup.clear();
        columnsSet.value.clear();
    }

    export function getColumn(keyPath: string[]): ColumnKey | undefined {
        return stringifiedLookup.get(JSON.stringify(keyPath));
    }

    export function setDisplayOn(keyPath: string[]) {
        const column = getColumn(keyPath);
        if (column) {
            column.shouldDisplay = true;
        }
    }

    export function setDisplayOff(keyPath: string[]) {
        console.log("setDisplayOff", keyPath);
        const column = getColumn(keyPath);
        if (column) {
            column.shouldDisplay = false;
        }
    }

    export function getAllSingleLevelColumns(): ColumnKey[] {
        return Array.from(columnsSet.value).filter(col => col.effectiveLookupPath.length == 1);
    }

    export function getTotalVisibleSingleLevelColumns(): number {
        return Array.from(columnsSet.value).filter(col => col.effectiveLookupPath.length == 1 && col.shouldDisplay).length;
    }

    export function getTotalHiddenSingleLevelColumns(): number {
        return getAllSingleLevelColumns().length - getTotalVisibleSingleLevelColumns();
    }

    export function getTotalCollapsedCountString(): string {
        const totalColumns = getAllSingleLevelColumns().length;
        const totalHiddenColumns = getTotalHiddenSingleLevelColumns();
        return `Collapsed (${totalHiddenColumns}/${totalColumns})`;
    }

    export function getAllNestedDepthColumns(): ColumnKey[] {
        return Array.from(columnsSet.value).filter(col => col.effectiveLookupPath.length > 1);
    }

    export function addColumn(columnKey: ColumnKey) {
        const stringifiedEffectiveLookupPath = JSON.stringify(columnKey.effectiveLookupPath);
        if (stringifiedLookup.has(stringifiedEffectiveLookupPath)) {
            return;
        }
        stringifiedLookup.set(stringifiedEffectiveLookupPath, columnKey);
        columnsSet.value.add(columnKey);
    }
}


export const DEPRECATE_coreDetectedKeys = ref<Set<ColumnKey>>(new Set());
export const availableColumns = ref<DispalyableColumn[]>(
    ([] as DispalyableColumn[])
        .concat(
            Array.from(CUSTOMARY_COLUMN_KEYS.value).map(key => ({ key, isEnabled: true }))
        )
    /* .concat(
        Array.from(SPECIAL_COLUMN_KEYS).map(key => ({ key, isEnabled: true }))
    ) */
);


export interface DerivedColumnGroup {
    originalColumn: string;
    isFromExpandedData: boolean;  // if this is extracted from the json-parsed data or the original payload
    derivedColumns: string[];
}


export namespace ColumnTypeTracker {
    export const timeColumns = ref(new Set<string>(['time', 'timestamp', 'createdAt', 'updatedAt']));
    export const categoricalColumns = ref(new Set<string>(['v', 'topic']));
    export const coerceToNumberColumns = ref(new Set<string>());
    export const derivedColumnGroups = ref<DerivedColumnGroup[]>([]);
}


export abstract class ContractableColumnsManager {
    public detectedKeys: Ref<Set<string>>;
    public hiddenKeys: Ref<Set<string>>;

    constructor(
        public expandedRows: Ref<Set<number>>,
        public expandedKeys: Ref<Set<string>>,
        public collapsedKeys: Ref<Set<string>>,
    ) {
        this.detectedKeys = ref(new Set<string>());
        this.hiddenKeys = ref(new Set<string>());
    }

    getVisibleKeys(): Array<string> {
        return Array.from(this.expandedKeys.value)
            .filter(key => !this.hiddenKeys.value.has(key));
    }

    contractKey(key: string): void {
        this.expandedKeys.value.delete(key);
        this.collapsedKeys.value.add(key);
    }

    hideKey(key: string): void {
        this.hiddenKeys.value.add(key);
        this.expandedKeys.value.delete(key);
        this.collapsedKeys.value.delete(key);
    }

    unhideKey(key: string): void {
        this.hiddenKeys.value.delete(key);
        this.expandedKeys.value.add(key);
        this.collapsedKeys.value.delete(key);
    }

    resetKeys(): void {
        this.expandedKeys.value.clear();
        this.collapsedKeys.value = new Set(this.detectedKeys.value);
    }

    clearAll(): void {
        this.detectedKeys.value.clear();
        this.hiddenKeys.value.clear();
        this.resetKeys();
    }
}

class ExpandableColumnsManager extends ContractableColumnsManager {
    public expandedExpandableDataRows: Ref<Set<number>>;
    public expandedExpandableDataKeys: Ref<Set<string>>;
    public expandableDataUnexpandedKeys: Ref<Set<string>>;

    constructor() {
        super(
            ref(new Set<number>()),
            ref(new Set<string>()),
            ref(new Set<string>()),
        );

        this.expandedExpandableDataRows = this.expandedRows;
        this.expandedExpandableDataKeys = this.expandedKeys;
        this.expandableDataUnexpandedKeys = this.collapsedKeys;
    }
}

export const expandableDataManager = new ExpandableColumnsManager();


class CollapsibleColumnsManager extends ContractableColumnsManager {
    public collapsibleDataExpandedRows: Ref<Set<number>>;
    public collapsibleDataCollapsedKeys: Ref<Set<string>>;
    public collapsibleDataExpandedKeys: Ref<Set<string>>;

    constructor() {
        super(
            ref(new Set<number>()),
            ref(new Set<string>()),
            ref(new Set<string>()),
        );

        this.collapsibleDataExpandedRows = this.expandedRows;
        this.collapsibleDataCollapsedKeys = this.collapsedKeys;
        this.collapsibleDataExpandedKeys = this.expandedKeys;
    }
}

export const collapsibleDataManager = new CollapsibleColumnsManager();


export interface RenderParams {
    node?: IRowNode<any>;
    data?: any;
    value?: any;
    coreDisplayParams: Set<string>;
    collapsedDataKeys: Set<string>;
    expandableDataExtractedKeys: Set<string>;

    expandableDataHiddenKeys: Set<string>;
    collapsibleDataHiddenKeys: Set<string>;

    expandableDataManager: ExpandableColumnsManager;
    collapsibleDataManager: CollapsibleColumnsManager;

    toggleExpandCollapsibleKeys: (rowIndex: number) => void;
    toggleExpandExpandableKeys: (rowIndex: number) => void;
    toggleContractExpandableKeys: (rowIndex: number) => void;

    currentRestoredKeys: Set<string>;
}

export function objectWithoutKeys<T>(obj: T, keySource: object | string[] | Set<string>): T {
    if (!obj) {
        return {} as T;
    }

    let excluder: (key: string) => boolean;
    if (keySource instanceof Set) {
        excluder = (key: string) => !keySource.has(key);
    } else if (Array.isArray(keySource)) {
        const keySet = new Set(keySource);
        excluder = (key: string) => !keySet.has(key);
    } else {
        const keySet = new Set(Object.keys(keySource));
        excluder = (key: string) => !keySet.has(key);
    }
    return (Object.fromEntries(Object.entries(obj as object).filter(([key]) => excluder(key))) ?? {}) as T;
}

export function objectWithKeys<T>(obj: T, keySource: object | string[] | Set<string>): T {
    let includer: (key: string) => boolean;
    if (keySource instanceof Set) {
        includer = (key: string) => keySource.has(key);
    } else if (Array.isArray(keySource)) {
        const keySet = new Set(keySource);
        includer = (key: string) => keySet.has(key);
    } else {
        const keySet = new Set(Object.keys(keySource));
        includer = (key: string) => keySet.has(key);
    }
    return Object.fromEntries(Object.entries(obj as object).filter(([key]) => includer(key))) as T;
}

export function parseValueWithUnitSuffix(value: string, shouldForceLowerCaseUnit: boolean = false): {
    value: number | null;
    unit: string | null;
} {
    if (!value) {
        return { value: null, unit: null };
    }
    const match = value.match(/^\s*(\d+(?:\.\d+)?|\.\d+)(\D*)\s*$/);
    if (!match) {
        return { value: parseFloat(value), unit: null };
    }

    return {
        value: parseFloat(match[1]),
        unit: shouldForceLowerCaseUnit ? match[2].toLowerCase() : match[2],
    };
}

export function parseTimeValue(value: string | number | null): Date | null {
    if (value == null) return null;
    if (typeof value === 'number') {
        // check if this should be adjusted
        const maybeFutureDate = value * 1000;
        if (Math.log10(maybeFutureDate) < 13.5) {
            return new Date(maybeFutureDate);
        } else {
            return new Date(value);
        }
    } else {
        return new Date(value);
    }
}