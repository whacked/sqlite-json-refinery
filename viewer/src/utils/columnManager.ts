import { ref } from 'vue';

export const expandedRows = ref(new Set<string>());
export const expandedKeys = ref(new Set<string>());

export interface RenderParams {
    data?: any;
    expandedKeys: Set<string>;
    toggleExpand: (id: string) => void;
    toggleContract: (id: string) => void;
}