import { ref } from 'vue';

export const coreDisplayParams = ref(new Set<string>([
    'id',
    'country',
    'createdAt',
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
    data?: any;
    coreDisplayParams: Set<string>;
    extraDataExtractedKeys: Set<string>;
    payloadExtractedKeys: Set<string>;
    toggleExpandExtraDataKeys: (id: string) => void;
    toggleContractExtraDataKeys: (id: string) => void;
    toggleExpandPayloadKeys: (id: string) => void;
    toggleContractPayloadKeys: (id: string) => void;
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

