<template>
    <div>
        <button v-if="payloadRenderable.rendered !== ''" @click="toggleExpandExpandableKeys">
            <span class="key-size-indicator">
                {{ payloadRenderable.keyCount }}
            </span>
            <span class="rendered-content">
                {{ payloadRenderable.rendered }}
            </span>
        </button>
    </div>
</template>

<style scoped>
</style>

<script setup lang="ts">
import { computed } from 'vue';
import { RenderParams, EXPANDABLE_DATA_COLUMN } from '@/utils/columnManager';
const hasKeys = computed(() => 
  props.params?.data?.[EXPANDABLE_DATA_COLUMN] && Object.keys(props.params.data[EXPANDABLE_DATA_COLUMN]).length > 0);

const props = defineProps<{
    params: RenderParams;
}>();

const payloadRenderable = computed(() => {
    if (!hasKeys.value) return { rendered: "", keyCount: 0 };
    let numKeys = 0;
    const reducedPayload: Record<string, any> = {};

    for (const key of Object.keys(props.params.data[EXPANDABLE_DATA_COLUMN])) {
        if (
            props.params.currentExpandedKeys.has(key)
        ) {
            continue;
        }
        reducedPayload[key] = props.params.data[EXPANDABLE_DATA_COLUMN][key];
        numKeys++;
    }
    if (numKeys == 0) {
        return { rendered: "", keyCount: 0 };
    }
    const stringified = JSON.stringify(reducedPayload);
    const rendered = stringified.length <= 100 ? stringified : stringified.substring(0, 100) + "...";
    return {
        rendered, keyCount: numKeys,
    }
})

const toggleExpandExpandableKeys = () => {
    if (props.params.node?.rowIndex != null) {
        props.params.toggleExpandExpandableKeys(props.params.node.rowIndex);
    }
};

</script>