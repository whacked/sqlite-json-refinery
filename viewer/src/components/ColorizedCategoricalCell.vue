<template>
    <div :style="Colorizer.makeTextContainerStyle(displayText)">
        {{ displayText }}
    </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { Colorizer } from './styling';
import { RenderParams } from '@/utils/columnManager';
import * as ColumnManager from '@/utils/columnManager';

const props = defineProps<{
    params: RenderParams;
}>();

const displayText = computed(() => {
    return props.params.colDef?.field != ColumnManager.EXPANDABLE_DATA_COLUMN
        ? props.params.value
        : props.params.value[props.params.colDef?.headerComponentParams.key.apparentLookupPath];
});
</script>