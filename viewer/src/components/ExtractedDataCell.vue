<template>
    <div>
        <button v-if="extraData.rendered !== ''" @click="toggleExpand">
            <span class="key-size-indicator">
                {{ extraData.keyCount }}
            </span>
            <span class="rendered-content">
                {{ extraData.rendered }}
            </span>
        </button>
    </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { RenderParams } from '@/utils/columnManager';
import * as ColumnManager from '@/utils/columnManager';

const props = defineProps<{
    params: RenderParams
}>();

const extraData = computed(() => {
    const subsetData = ColumnManager.objectWithoutKeys(
        props.params.data,
        new Set([
            ...ColumnManager.coreDisplayParams.value,
            ...ColumnManager.extractedDataExtractedKeys.value
        ])
    );
    const keyCount = Object.keys(subsetData).length;
    const rendered = keyCount === 0 ? "" : JSON.stringify(subsetData);
    return { rendered, keyCount };
});

const toggleExpand = () => props.params.toggleExpandExtraDataKeys(props.params.data.id);
</script>