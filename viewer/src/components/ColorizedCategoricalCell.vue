<template>
    <div :style="textContainerStyle">
        {{ params.value }}
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

const textContainerStyle = computed(() => {
    const columnKey = (props.params as any).colDef.field
    if (ColumnManager.isExpandableDataColumnKey(columnKey)) {
        return Colorizer.makeTextContainerStyle(
            props.params.data[ColumnManager.EXPANDABLE_DATA_COLUMN][ColumnManager.getExpandableDataColumnSubKey(columnKey)]);
    } else {
      return Colorizer.makeTextContainerStyle(props.params.value);
    }
});
</script>