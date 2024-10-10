<template>
  <div id="app">
      <DataTable />
  </div>
  <!-- hello world table
  <div class="myblock">
    <ag-grid-vue
      :rowData="rowData"
      :columnDefs="colDefs"
      style="height: 500px"
      class="ag-theme-quartz"
    >
    </ag-grid-vue>
  </div>
  -->
</template>

<style scoped>
#app {
  height: 600px;
  background: beige;
}

.myblock {
  width: 700px;
  border: 2px solid green;
}
</style>

<script setup lang="ts">
import DataTable from './components/DataTable.vue'
import "ag-grid-community/styles/ag-grid.css"; // Mandatory CSS required by the Data Grid
import "ag-grid-community/styles/ag-theme-quartz.css"; // Optional Theme applied to the Data Grid
import { ref } from 'vue';

/** hello world table */
/* import { AgGridVue } from "ag-grid-vue3"; // Vue Data Grid Component

const rowData = ref([
   { make: "Tesla", model: "Model Y", price: 64950, electric: true },
   { make: "Ford", model: "F-Series", price: 33850, electric: false },
   { make: "Toyota", model: "Corolla", price: 29600, electric: false },
]);

const colDefs = ref([
   { field: "make" },
   { field: "model" },
   { field: "price" },
   { field: "electric" }
]); */
</script>