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
    if(
        props.params.value != null &&
        props.params.colDef?.field != null &&ColumnManager.CUSTOMARY_COLUMN_KEYS.has(props.params.colDef.field)
    ) {
        return props.params.value
    }

    let lookupKey = props.params.colDef?.headerComponentParams?.key;
    if(lookupKey == null && props.params.colDef?.field != null) {
      lookupKey = props.params.colDef.field;
    }

    let value: string;
    if(typeof lookupKey == "string") {
      value = props.params.data[lookupKey]?.toString() ?? "";
    } else {
      value = props.params.data[lookupKey.apparentLookupPath]?.toString() ?? "";
    }
    return value;
});
</script>