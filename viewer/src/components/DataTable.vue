<template>
  <div>
    <div class="filters">
      <input v-model="quickFilterText" placeholder="Quick filter..." @input="onQuickFilterChanged" />
    </div>
    <div class="ag-theme-alpine" style="height: 500px; width: 100%;">
      <div>
        {{ dataStore.totalRows }} rows
      </div>
      <ag-grid-vue
        style="height: 100%; border: 4px dashed orange;"
        class="ag-theme-quartz"

        :columnDefs="columnDefs"
        :datasource="datasource"
        
        :defaultColDef="defaultColDef"
        :rowBuffer="rowBuffer"
        :rowModelType="rowModelType"
        :paginationPageSize="paginationPageSize"
        :cacheBlockSize="cacheBlockSize"
        :infiniteInitialRowCount="infiniteInitialRowCount"
        @grid-ready="onGridReady"
      />
    </div>
  </div>
</template>

<style scoped>
.filters {
  margin-bottom: 10px;
}

.ag-theme-quartz {
    --ag-header-column-separator-display: block;
    --ag-header-column-separator-height: 100%;
    --ag-header-column-separator-width: 2px;
    --ag-header-column-separator-color: purple;

    --ag-header-column-resize-handle-display: block;
    --ag-header-column-resize-handle-height: 25%;
    --ag-header-column-resize-handle-width: 5px;
    --ag-header-column-resize-handle-color: orange;
    
}

.ag-header-group-cell.ag-column-first {
    background-color: #2244CC66;
    color: white;
}
.ag-header-cell.ag-column-first {
    background-color: #2244CC44;
    color: white;
}
.ag-floating-filter.ag-column-first {
    background-color: #2244CC22;
}

.ag-header-group-cell.ag-column-last {
    background-color: #33CC3366;
}
.ag-header-cell.ag-column-last {
    background-color: #33CC3344;
}
.ag-floating-filter.ag-column-last {
    background-color: #33CC3322;
}
</style>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue';
import { AgGridVue } from 'ag-grid-vue3';
import { ColDef, GridApi, GridReadyEvent, IDatasource } from 'ag-grid-community';
import { useDataStore } from '@/stores/dataStore';
import { useWindowingService } from '@/services/windowingService';

const dataStore = useDataStore();
const windowingService = useWindowingService(50); // pageSize of 50

const columnDefs = ref<ColDef[]>([]);
const quickFilterText = ref('');

const gridApi = ref<GridApi | null>(null);

const defaultColDef = reactive({
  flex: 1,
  minWidth: 100,
  sortable: true,
  filter: true,
});

const rowBuffer = 0;
const rowModelType = 'infinite';
const paginationPageSize = 10;
const cacheBlockSize = 10;
const infiniteInitialRowCount = 1;

 
const datasource: IDatasource = {
  getRows: (params) => {
    const startRow = params.startRow;
    const endRow = params.endRow;
    
    console.log(`Requesting rows from ${startRow} to ${endRow}`);
    
    dataStore.fetchData(startRow, endRow).then((data) => {
      if (columnDefs.value.length === 0) {
        columnDefs.value = generateColumnDefs(data);
        console.log('Generated column definitions:', columnDefs.value);
      }
      
      const lastRow = data.length < endRow - startRow ? startRow + data.length : undefined;
      
      console.log(`Received ${data.length} rows; lastRow: ${lastRow}`);
      console.log(params)
      console.log(data.slice(0, 3))
      params.successCallback(data, lastRow);
    });
  }
};
// */

const onGridReady = (params: GridReadyEvent) => {
  gridApi.value = params.api;
  // columnApi.value = params.columnApi;
  console.log('Grid ready', params, dataStore.totalRows);
  windowingService.setTotalItems(dataStore.totalRows);
};

const generateColumnDefs = (dataRows: any[]): ColDef[] => {
  const discoveredColumns = new Set<string>()
  const orderedColumns: ColDef[] = []

  dataRows.forEach((row) => {
    Object.keys(row).forEach((key) => {
      if(discoveredColumns.has(key)) {
        return
      }
      discoveredColumns.add(key)  
      orderedColumns.push({
        field: key,
        headerName: key,
        width: Math.floor(Math.random() * 150) + 100, // Random width between 100 and 250
        resizable: Math.random() > 0.5, // Randomly resizable
        sortable: Math.random() > 0.3, // More likely to be sortable
        filter: Math.random() > 0.4 ? 'agTextColumnFilter' : false, // Randomly add text filter
        cellStyle: { 
          color: ['#000000', '#333333', '#666666'][Math.floor(Math.random() * 3)], // Random text color
          fontWeight: Math.random() > 0.7 ? 'bold' : 'normal', // Occasionally bold
          textAlign: ['left', 'center', 'right'][Math.floor(Math.random() * 3)] // Random alignment
        },
        headerClass: Math.random() > 0.5 ? 'custom-header' : '', // Randomly add a custom header class
      })
    })
  })

  return orderedColumns
};

const onQuickFilterChanged = () => {
  quickFilterText.value = quickFilterText.value.trim();
  console.log('Quick filter changed to:', quickFilterText.value);
};

onMounted(() => {
  dataStore.initializeData();
});
</script>
