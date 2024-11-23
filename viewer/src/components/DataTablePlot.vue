<template>
    <div ref="plotContainer"></div>
</template>

<style scoped>
</style>

<script setup lang="ts">

import * as Plotly from 'plotly.js-dist';
import { ref, watch } from 'vue';
import { parse } from 'vue/compiler-sfc';
import * as ColumnManager from '@/utils/columnManager';
import { ColumnKey } from '@/utils/columnManager';

const plotContainer = ref(null);

const props = defineProps<{
  rowData: any[];
  plotSettings: PlotSettings;
}>();

export interface PlotSettings {
  xColumn: ColumnKey | null;
  yColumn: ColumnKey | null;
}


</script>