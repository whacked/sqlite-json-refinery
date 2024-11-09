import { Ref, ref } from 'vue';
import { ColDef, IRowNode } from 'ag-grid-community';

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

export const CUSTOMARY_COLUMN_KEYS = new Set([
    'id',
    'country',
    'createdAt',

    'time',
    'topic',
    'category',
    'entry',
    /* 'price', */
]);
export const SERIALIZED_CUSTOMARY_COLUMN_KEYS = ref<Set<string>>(
    new Set(
        Array.from(CUSTOMARY_COLUMN_KEYS)
            .map(key => JSON.stringify([key]))));

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

export enum ColumnTransformation {
    NUMBER,
}

export enum ColumnRendererType {
    TIME,
    CATEGORICAL,
}

export const DEFAULT_TIME_COLUMN_KEYS = new Set(['time', 'timestamp', 'createdAt', 'updatedAt']);
export const DEFAULT_CATEGORICAL_COLUMN_KEYS = new Set(['v', 'topic']);

export interface ColumnKey {
    effectiveLookupPath: string[];  // the actual code-friendly lookup path
    apparentLookupPath: string;  // the that the human uses to refer to the data
    displayString: string;
    shouldDisplay: boolean;
    serializedEffectiveLookupPath: string;
    derivedFromColumn?: ColumnKey | null;
    transformations?: Array<ColumnTransformation | null>;
    renderType?: ColumnRendererType | null;
}

export namespace UsableColumns {
    export const columnsSet = ref<Set<ColumnKey>>(new Set());

    const stringifiedLookup = new Map<string, ColumnKey>();

    export const visibleSingleLevelColumnDataRowIds = ref<Set<number>>(new Set());
    export const visibleNestedDepthColumnDataRowIds = ref<Set<number>>(new Set());

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

    export function getTotalVisibleNestedDepthColumns(): number {
        return Array.from(columnsSet.value).filter(col => col.effectiveLookupPath.length > 1 && col.shouldDisplay).length;
    }

    export function getTotalHiddenNestedDepthColumns(): number {
        return getAllNestedDepthColumns().length - getTotalVisibleNestedDepthColumns();
    }

    export function getDerivedColumns(): ColumnKey[] {
        return Array.from(columnsSet.value).filter(col => col.derivedFromColumn != null);
    }

    export function getTotalTransformedColumns(transformationType?: ColumnTransformation): number {
        let totalColumns = 0
        for (const column of columnsSet.value) {
            if (transformationType == null) {
                totalColumns++;
            } else if (column.transformations?.includes(transformationType)) {
                totalColumns++;
            }
        }
        return totalColumns;
    }

    export function getTotalSpeciallyRenderedColumns(rendererType?: ColumnRendererType): number {
        let totalColumns = 0;
        for (const column of columnsSet.value) {
            if (rendererType == null) {
                totalColumns++;
            } else if (column.renderType == rendererType) {
                totalColumns++;
            }
        }
        return totalColumns;
    }

    export function getTotalCollapsedCountString(): string {
        let totalColumns = 0;
        let totalHiddenColumns = 0;
        for (const column of columnsSet.value) {
            if (column.effectiveLookupPath.length == 1
                && !CUSTOMARY_COLUMN_KEYS.has(column.apparentLookupPath)
            ) {
                totalColumns++;
                if (!column.shouldDisplay) {
                    totalHiddenColumns++;
                }
            }
        }
        return `Collapsed (${totalHiddenColumns}/${totalHiddenColumns})`;
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

    export function removeColumn(columnKey: ColumnKey) {
        const stringifiedEffectiveLookupPath = JSON.stringify(columnKey.effectiveLookupPath);
        stringifiedLookup.delete(stringifiedEffectiveLookupPath);
        columnsSet.value.delete(columnKey);
    }
}

export const collapsibleDataExpandedRows: Ref<Set<number>> = ref(new Set<number>());
export const expandableDataExpandedRows: Ref<Set<number>> = ref(new Set<number>());

export interface RenderParams {
    node?: IRowNode<any>;
    data?: any;
    value?: any;
    colDef?: ColDef;
    valueFormatted?: string;
    coreDisplayParams: Set<string>;

    toggleExpandCollapsibleKeys: (rowIndex: number) => void;
    toggleExpandExpandableKeys: (rowIndex: number) => void;
    toggleContractExpandableKeys: (rowIndex: number) => void;

    currentExpandedKeys: Set<string>;
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
    if (typeof value !== 'string') {
        return { value: value, unit: null };
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

export function parseTimeValue(value: string | number | null | undefined): Date | null {
    if (value == null || value == undefined || value == "") return null;
    try {
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
    } catch (e) {
        return null;
    }
}
