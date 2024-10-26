import { ref } from 'vue';
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

export const timeColumns = ref(new Set<string>(['time', 'timestamp', 'createdAt', 'updatedAt']));
export const categoricalColumns = ref(new Set<string>(['v', 'topic']));

export abstract class ContractableColumnsManager {
    abstract contractKey(key: string): void;
    abstract resetKeys(): void;
    abstract clearAll(): void;
}

class ExpandableColumnsManager extends ContractableColumnsManager {
    public expandedExpandableDataRows = ref(new Set<number>());
    public expandableDataDetectedKeys = ref(new Set<string>());
    public expandedExpandableDataKeys = ref(new Set<string>());
    public expandableDataUnexpandedKeys = ref(new Set<string>());

    constructor() {
        super();
    }

    contractKey(key: string) {
        this.expandedExpandableDataKeys.value.delete(key);
        this.expandableDataUnexpandedKeys.value.add(key);
    }

    resetKeys() {
        this.expandedExpandableDataKeys.value.clear()
        this.expandableDataUnexpandedKeys.value = new Set(this.expandableDataDetectedKeys.value);
    }

    clearAll() {
        this.expandableDataDetectedKeys.value.clear();
        this.resetKeys();
    }
}

export const expandableDataManager = new ExpandableColumnsManager();


class CollapsibleColumnsManager extends ContractableColumnsManager {
    public collapsibleDataExpandedRows = ref(new Set<number>());
    public collapsibleDataDetectedKeys = ref(new Set<string>());
    public collapsibleDataCollapsedKeys = ref(new Set<string>());
    public collapsibleDataExpandedKeys = ref(new Set<string>());

    contractKey(key: string) {
        this.collapsibleDataExpandedKeys.value.delete(key);
        this.collapsibleDataCollapsedKeys.value.add(key);
    }

    resetKeys() {
        this.collapsibleDataExpandedKeys.value.clear()
        this.collapsibleDataCollapsedKeys.value = new Set(this.collapsibleDataDetectedKeys.value);
    }

    clearAll() {
        this.collapsibleDataDetectedKeys.value.clear();
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

