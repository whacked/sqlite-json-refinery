<template>
    <div>
        <button v-if="renderedPayload !== ''" @click="toggleExpand">
            <span class="key-size-indicator">
                {{ Object.keys(props.params.data.payload).length }}
            </span>
            <span class="rendered-content">
                {{ renderedPayload }}
            </span>
        </button>
    </div>
</template>

<style scoped>
</style>

<script setup lang="ts">
import { computed } from 'vue';
import { RenderParams } from '@/utils/columnManager';
const hasKeys = computed(() => Object.keys(props.params.data.payload).length > 0);

const props = defineProps<{
    params: RenderParams
}>();

const renderedPayload = computed(() => renderParams(props.params.data.payload));

function renderParams(payload: any) {
    if (!hasKeys.value) return "";
    let numKeys = 0;
    const reducedPayload: Record<string, any> = {};
    for (const key of Object.keys(payload)) {
        if (props.params.payloadExtractedKeys.has(key)) {
            continue;
        }
        reducedPayload[key] = payload[key];
        numKeys++;
    }
    if (numKeys == 0) {
        return "";
    }
    const stringified = JSON.stringify(reducedPayload);
    if (stringified.length <= 100) {
        return stringified;
    } else {
        return stringified.substring(0, 100) + "...";
    }
}

const toggleExpand = () => props.params.toggleExpandPayloadKeys(props.params.data.id);

</script>