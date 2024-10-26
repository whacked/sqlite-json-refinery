<template>
    <div :style="{ background: backgroundColor, color: textColor }">
        {{ params.value }}
    </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { colorHash } from './styling';
import { RenderParams } from '@/utils/columnManager';
import chroma from 'chroma-js';

const props = defineProps<{
    params: RenderParams;
}>();

const backgroundColor = computed(() => {
    if (!props.params.value) return "";
    return colorHash.hex(props.params.value);
});

const textColor = computed(() => {
    if (!backgroundColor.value) return "";
    const isDark = chroma(backgroundColor.value).luminance() < 0.6;
    return isDark ? 'white' : 'black';
});
</script>