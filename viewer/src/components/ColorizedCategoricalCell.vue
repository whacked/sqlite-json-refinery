<template>
    <div :style="textContainerStyle">
        {{ params.value }}
    </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { Colorizer } from './styling';
import { RenderParams } from '@/utils/columnManager';

const props = defineProps<{
    params: RenderParams;
}>();

const textContainerStyle = computed(() => {
    return Colorizer.makeTextContainerStyle(props.params.value);
});
</script>