<template>
    <div class="time-cell"
    :style="{ background: backgroundColor }"
    >
        {{ myDate?.toISOString() }}
    </div>
</template>

<!-- somehow scoped doesn't work here -->
<style>
.time-cell {
    font-weight: bold;
}
</style>

<script setup lang="ts">
import { parseTimeValue, RenderParams } from '@/utils/columnManager';
import { useDataStore } from '@/stores/dataStore';
import { computed } from 'vue';

const dataStore = useDataStore();
const props = defineProps<{
    params: RenderParams;
}>();

const myDate = computed(() => {
  return parseTimeValue(props.params.value);
});

interface TimeBreakpoints {
  week: number;
  month: number;
  year: number;
  fiveYears: number;
}

const defaultBreakpoints: TimeBreakpoints = {
  week: 7,
  month: 30,
  year: 365,
  fiveYears: 1825
};

const getAgeColor = (date: Date, now: Date, breakpoints: TimeBreakpoints = defaultBreakpoints) => {
  const ageInDays = (now.getTime() - date.getTime()) / (1000 * 60 * 60 * 24);
  
  if (ageInDays <= 0) return '#FF0000'; // Today or future: full red
  if (ageInDays >= breakpoints.fiveYears) return '#CCCCCC'; // 5 years or older: light gray

  let startColor, endColor, startDay, endDay;

  if (ageInDays < breakpoints.week) {
    startColor = [255, 0, 0];
    endColor = [255, 128, 128];
    startDay = 0;
    endDay = breakpoints.week;
  } else if (ageInDays < breakpoints.month) {
    startColor = [255, 128, 128];
    endColor = [255, 192, 192];
    startDay = breakpoints.week;
    endDay = breakpoints.month;
  } else if (ageInDays < breakpoints.year) {
    startColor = [255, 192, 192];
    endColor = [255, 224, 224];
    startDay = breakpoints.month;
    endDay = breakpoints.year;
  } else {
    startColor = [255, 224, 224];
    endColor = [204, 204, 204];
    startDay = breakpoints.year;
    endDay = breakpoints.fiveYears;
  }

  const fraction = (ageInDays - startDay) / (endDay - startDay);
  const r = Math.round(startColor[0] + fraction * (endColor[0] - startColor[0]));
  const g = Math.round(startColor[1] + fraction * (endColor[1] - startColor[1]));
  const b = Math.round(startColor[2] + fraction * (endColor[2] - startColor[2]));

  return `rgb(${r}, ${g}, ${b})`;
};

const backgroundColor = computed(() => {
  if (myDate.value == null) return null;
  return getAgeColor(myDate.value, dataStore.now);
});
</script> 