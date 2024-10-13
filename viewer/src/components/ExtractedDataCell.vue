<template>
    <div>
        <button v-if="rendereredExtraData !== ''" @click="toggleExpand">
            {{ rendereredExtraData }}
        </button>
    </div>
</template>

<style scoped>
</style>

<script setup lang="ts">
import { computed } from 'vue';
import { RenderParams } from '@/utils/columnManager';
const hasKeys = computed(() => Object.keys(props.params.data.payload).length > 0);
import * as ColumnManager from '@/utils/columnManager';

const props = defineProps<{
    params: RenderParams
}>();

const rendereredExtraData = computed(() => renderParams(props.params.data));

function renderParams(data: any) {
    const subsetData = ColumnManager.objectWithoutKeys(
        data,
        new Set([
            ...ColumnManager.coreDisplayParams.value,
            ...ColumnManager.extractedDataExtractedKeys.value
        ])
    );
    if (Object.keys(subsetData).length === 0) {
        return "";
    }
    return JSON.stringify(subsetData);
}

const toggleExpand = () => props.params.toggleExpandExtraDataKeys(props.params.data.id);
</script>