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






  <ag-grid-vue
    class="ag-theme-alpine"
    style="height: 500px; width: 100%;"
    :columnDefs="columnDefs2"
    :rowModelType="'infinite'"
    :cacheBlockSize="10"
    :infiniteInitialRowCount="10"
    :maxConcurrentDatasourceRequests="1"
    :maxBlocksInCache="10"
    :datasource="dataSource2"
    @grid-ready="onGridReady2"
  />
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


import { AgGridVue } from "ag-grid-vue3"; // Vue Data Grid Component
/** hello world table */
/*

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









import { ColDef, GridApi, GridReadyEvent, IDatasource } from 'ag-grid-community';
import { faker } from '@faker-js/faker';

// Fake data generator
const generateFakeRow = (index: number) => ({
  id: index,
  name: faker.person.fullName(),
  email: faker.internet.email(),
  city: faker.location.city()
});

// Fake async data fetcher
const fetchData2 = (startRow: number, endRow: number): Promise<any[]> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      const rowData = [];
      for (let i = startRow; i < endRow; i++) {
        rowData.push(generateFakeRow(i));
      }
      resolve(rowData);
    }, 200); // Simulate network delay
  });
};

const gridApi2 = ref<GridApi | null>(null);
const columnDefs2 = ref<ColDef[]>([
  { field: 'id' },
  { field: 'name' },
  { field: 'email' },
  { field: 'city' }
]);

const onGridReady2 = (params: GridReadyEvent) => {
  gridApi2.value = params.api;
/* 
  const dataSource: IDatasource = {
    getRows: (params) => {
      console.log('Fetching rows:', params.startRow, 'to', params.endRow);
      fetchData2(params.startRow, params.endRow).then(rowData => {
        params.successCallback(rowData, 100000); // Assume 100,000 total rows
      });
    }
  };

  console.log("setting datasource", params.api);
  params.api.setDatasource(dataSource); */
};


const dataSource2 = ref<IDatasource>({
  getRows: (params) => {
    console.log('Fetching rows:', params.startRow, 'to', params.endRow);
    fetchData2(params.startRow, params.endRow).then(rowData => {
      params.successCallback(rowData, 100000); // Assume 100,000 total rows
    });
  }
});

</script>