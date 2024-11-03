<template>
  <span v-for="(word, index) in props.columnText.split('.').filter(w => w.length > 0)" 
        :key="word"
        class="ag-header-cell-text"
        :style="Colorizer.makeTextContainerStyle(word)">
    {{ index > 0 ? '.' + word : word }}
  </span>
</template>

<script setup lang="ts">
import { Colorizer } from './styling';

const props = defineProps<{
    columnText: string;
}>();
</script>