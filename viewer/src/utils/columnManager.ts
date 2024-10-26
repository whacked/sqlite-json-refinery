import { Ref, ref } from 'vue';
import { IRowNode } from 'ag-grid-community';

export const EXPANDABLE_DATA_COLUMN = 'payload';
export const EXPANDABLE_DATA_COLUMN_SHADOW = 'payloadString';
export const COLLAPSABLE_DATA_COLUMN = 'collapsedData';
export const COLLAPSABLE_DATA_COLUMN_SHADOW = 'collapsedDataString';

export const COMMON_COLUMN_KEYS = ref(new Set<string>([
    'id',
    'country',
    'createdAt',

    'time',
    'topic',
    'category',
    'entry',
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

export const coreDetectedKeys = ref<Set<string>>(new Set());
export const availableColumns = ref<DispalyableColumn[]>(
    ([] as DispalyableColumn[])
        .concat(
            Array.from(COMMON_COLUMN_KEYS.value).map(key => ({ key, isEnabled: true }))
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

    abstract getVisibleKeys(): Array<string>;
    abstract contractKey(key: string): void;
    abstract hideKey(key: string): void;
    abstract resetKeys(): void;
    abstract clearAll(): void;

    constructor() {
        this.detectedKeys = ref(new Set<string>());
        this.hiddenKeys = ref(new Set<string>());
    }
}

class ExpandableColumnsManager extends ContractableColumnsManager {
    public expandedExpandableDataRows = ref(new Set<number>());
    public expandedExpandableDataKeys = ref(new Set<string>());
    public expandableDataUnexpandedKeys = ref(new Set<string>());

    getVisibleKeys(): Array<string> {
        return Array.from(this.expandedExpandableDataKeys.value)
            .filter(key => !this.hiddenKeys.value.has(key));
    }

    contractKey(key: string) {
        this.expandedExpandableDataKeys.value.delete(key);
        this.expandableDataUnexpandedKeys.value.add(key);
    }

    hideKey(key: string) {
        this.expandedExpandableDataKeys.value.delete(key);
        this.expandableDataUnexpandedKeys.value.delete(key);
        this.hiddenKeys.value.add(key);
    }

    resetKeys() {
        this.expandedExpandableDataKeys.value.clear()
        this.expandableDataUnexpandedKeys.value = new Set(this.detectedKeys.value);
    }

    clearAll() {
        this.detectedKeys.value.clear();
        this.hiddenKeys.value.clear();
        this.resetKeys();
    }
}

export const expandableDataManager = new ExpandableColumnsManager();


class CollapsibleColumnsManager extends ContractableColumnsManager {
    public collapsibleDataExpandedRows = ref(new Set<number>());
    public collapsibleDataCollapsedKeys = ref(new Set<string>());
    public collapsibleDataExpandedKeys = ref(new Set<string>());

    getVisibleKeys(): Array<string> {
        return Array.from(this.collapsibleDataExpandedKeys.value)
            .filter(key => !this.hiddenKeys.value.has(key));
    }

    contractKey(key: string) {
        this.collapsibleDataExpandedKeys.value.delete(key);
        this.collapsibleDataCollapsedKeys.value.add(key);
    }

    hideKey(key: string) {
        this.collapsibleDataCollapsedKeys.value.delete(key);
        this.collapsibleDataExpandedKeys.value.delete(key);
        this.hiddenKeys.value.add(key);
    }

    resetKeys() {
        this.collapsibleDataExpandedKeys.value.clear()
        this.collapsibleDataCollapsedKeys.value = new Set(this.detectedKeys.value);
    }

    clearAll() {
        this.detectedKeys.value.clear();
        this.hiddenKeys.value.clear();
        this.resetKeys();
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
    toggleContractCollapsibleKeys: (rowIndex: number) => void;
    toggleExpandExpandableKeys: (rowIndex: number) => void;
    toggleContractExpandableKeys: (rowIndex: number) => void;
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