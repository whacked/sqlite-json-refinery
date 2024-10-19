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
export const availableColumns = ref<DispalyableColumn[]>(
    ([] as DispalyableColumn[])
        .concat(
            Array.from(COMMON_COLUMN_KEYS.value).map(key => ({ key, isEnabled: true }))
        )
    /* .concat(
        Array.from(SPECIAL_COLUMN_KEYS).map(key => ({ key, isEnabled: true }))
    ) */
);

export const expandedExpandableDataRows = ref(new Set<number>());
export const expandedExpandableDataKeys = ref(new Set<string>());
export const collapsableDataExpandedRows = ref(new Set<number>());
export const collapsableDataExtractedKeys = ref(new Set<string>());


export interface RenderParams {
    node?: IRowNode<any>;
    data?: any;
    coreDisplayParams: Set<string>;
    collapsedDataKeys: Set<string>;
    expandableDataExtractedKeys: Set<string>;
    toggleExpandCollapsibleKeys: (rowIndex: number) => void;
    toggleContractCollapsibleKeys: (rowIndex: number) => void;
    toggleExpandExpandableKeys: (rowIndex: number) => void;
    toggleContractExpandableKeys: (rowIndex: number) => void;
}

export function objectWithoutKeys<T>(obj: T, keySource: object | string[] | Set<string>): T {
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
    return Object.fromEntries(Object.entries(obj as object).filter(([key]) => excluder(key))) as T;
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

