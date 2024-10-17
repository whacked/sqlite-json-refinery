payloadRenderable.rendered<template>
    <div>
        <button v-if="collapsedData.rendered !== ''" @click="toggleExpand">
            <span class="key-size-indicator">
                {{ collapsedData.keyCount }}
            </span>
            <span class="rendered-content">
                {{ collapsedData.rendered }}
            </span>
        </button>
    </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { RenderParams } from '@/utils/columnManager';
import * as ColumnManager from '@/utils/columnManager';

const props = defineProps<{
    params: RenderParams;
}>();

const collapsedData = computed(() => {
    const subsetData = ColumnManager.objectWithoutKeys(
        props.params.data,
        new Set([
            ...ColumnManager.coreDisplayParams.value,
            ...ColumnManager.collapsableDataExtractedKeys.value
        ])
    );
    const keyCount = Object.keys(subsetData).length;
    const rendered = keyCount === 0 ? "" : JSON.stringify(subsetData);
    return { rendered, keyCount };
});

const toggleExpand = () => {
    if (props.params.node.rowIndex != null) {
        props.params.toggleExpandCollapsableDataKeys(props.params.node.rowIndex);
    }
};
</script>