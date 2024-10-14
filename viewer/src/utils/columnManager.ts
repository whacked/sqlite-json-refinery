import { ref } from 'vue';
import { IRowNode } from 'ag-grid-community';

export const coreDisplayParams = ref(new Set<string>([
    // FIXME -- need more flexible setup
    // sample columns
    /* 'id',
    'country',
    'createdAt', */

    'time',
    'topic',
    'category',
    'entry',

    'payload',
    // derived columns
    'payloadString',
    'extraData',
    'extraDataString',
]));

export const expandedPayloadRows = ref(new Set<string>());
export const expandedPayloadKeys = ref(new Set<string>());
export const extractedDataExpandedRows = ref(new Set<string>());
export const extractedDataExtractedKeys = ref(new Set<string>());


export interface RenderParams {
    node: IRowNode<any>;
    data?: any;
    coreDisplayParams: Set<string>;
    extraDataExtractedKeys: Set<string>;
    payloadExtractedKeys: Set<string>;
    toggleExpandExtraDataKeys: (rowIndex: number) => void;
    toggleContractExtraDataKeys: (rowIndex: number) => void;
    toggleExpandPayloadKeys: (rowIndex: number) => void;
    toggleContractPayloadKeys: (rowIndex: number) => void;
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

