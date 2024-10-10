import { defineStore } from 'pinia';
import { generateData } from '@/utils/dataGenerators';
import { ref } from 'vue';

export const useDataStore = defineStore('data', () => {
  const data = ref<any[]>([]);
  const totalRows = ref(1000); // Simulating a large dataset

  function initializeData() {
    // In a real scenario, this might involve setting up WebSocket connections or other initializations
    console.log('Data store initialized');
  }

  async function fetchData(startRow: number, endRow: number) {
    // Simulate server delay
    await new Promise(resolve => setTimeout(resolve, 100));

    // Generate data if it doesn't exist
    if (data.value.length < endRow) {
      const newData = generateData(endRow - data.value.length);
      data.value = [...data.value, ...newData];
    }

    console.log(`Fetching data from ${startRow} to ${endRow}:`, data.value.slice(startRow, endRow));
    return data.value.slice(startRow, endRow);
  }

  return { initializeData, fetchData, totalRows };
});