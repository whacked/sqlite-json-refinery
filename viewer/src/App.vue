<template>
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

  <button @click="loadNewFile">Load JSONL File</button>
  <DataTable
    v-if="fileCache.length > 0"
    :rowData="Array.from(fileCache.values())"
   />
   <DataTable
    v-else
    :rowData="[]"
   />
  <ag-grid-vue
    v-if="fileCache.length > 0"
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









import { ColDef, GridApi, GridReadyEvent, IDatasource, ValueGetterParams } from 'ag-grid-community';
import { faker } from '@faker-js/faker';


// TODO move this to autogen
interface CommonPayloadData {
  time: number;
  category: string;
  entry: string;
  payload: Record<string, string>;
}
const commonKeys = new Set(['time', 'category', 'entry', 'payload']);


// Fake data generator
const generateFakeRow = (index: number): CommonPayloadData => ({
  time: index,
  category: faker.location.city(),
  entry: faker.person.fullName(),
  payload: { email: faker.internet.email() },
});

// Fake async data fetcher
const fetchData2_ = (startRow: number, endRow: number): Promise<CommonPayloadData[]> => {
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
  {
    headerName: "#",
    valueGetter: (params: ValueGetterParams) => {
      return (params.node?.rowIndex ?? 0) + 1;
    },
    width: 80,
  },
  { field: 'time' },
  { field: 'category' },
  { field: 'entry' },
  {
    headerName: 'Expandable Data',
    valueGetter: (params: ValueGetterParams) => {
      // return a canonicalized json string of the payload
      // whose keys are not included in the default key list
      const allData: any = params.data ?? {};
      const filteredPayload = Object.keys(allData).filter(
        key => !commonKeys.has(key)).reduce((obj: Record<string, string>, key) => {
          obj[key] = allData[key];
          return obj;
        }, {});
      return filteredPayload;
    },
    headerClass: 'my-ag-table-derived-column',
    cellRenderer: (params: CellRendererParams) => {
      const stringRepresentation = JSON.stringify(params.value);
      const numKeys = Object.keys(params.value).length;
      return `<button onclick="alert(${numKeys})">(${numKeys}) ${stringRepresentation}</button>`;
    },
    cellRendererParams: {
      onClick: () => alert('Button clicked'),
    },
  },
  { field: 'payload' },
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

interface RowFetchWindow {
  rows: any[],
  startIndex: number,
  endIndex: number,
  totalRowCount: number,
}

const dataSource2 = ref<IDatasource>({
  getRows: (params) => {
    console.log('Fetching rows:', params.startRow, 'to', params.endRow, "out of");
    fetchData2(params.startRow, params.endRow).then(rowData => {
      console.log("got rows", rowData.rows, rowData.totalRowCount);
      if(rowData.rows.length === 0) {
        params.failCallback();
      } else {
        params.successCallback(rowData.rows, rowData.totalRowCount);
      }
    });
  }
});


async function getFileHandle() {
  if ('showOpenFilePicker' in window) {
    const [handle] = await (window as any).showOpenFilePicker();
    return handle;
  } else {
    return new Promise((resolve) => {
      const input = document.createElement('input');
      input.type = 'file';
      input.onchange = (e) => resolve((e.target as HTMLInputElement).files?.[0]);
      input.click();
    });
  }
}

const fileCache = ref<any[]>([]);

async function loadFile() {
  const fileHandle = await getFileHandle();
  const file = 'getFile' in fileHandle ? await fileHandle.getFile() : fileHandle;
  const text = await file.text();
  const lines = text.split('\n');

  for (const line of lines) {
    if (line.trim() !== '') {
      try {
        const parsedLine = JSON.parse(line);
        fileCache.value.push(parsedLine);
      } catch (error) {
        console.error('Error parsing line:', line);
        console.error(error);
      }
    }
  }
  console.log('fileCache length:', fileCache.value.length);
}


const fetchData2 = async (startRow: number, endRow: number): Promise<RowFetchWindow> => {
  console.log(fileCache.value)

  if (fileCache.value.length >= endRow) {
    return {
      rows: fileCache.value.slice(startRow, endRow),
      startIndex: startRow,
      endIndex: endRow,
      totalRowCount: fileCache.value.length,
    };
  }

  return {
    rows: [],
    startIndex: 0,
    endIndex: 0,
    totalRowCount: 0,
  };

  if (fileCache.value.length === 0) {
    await loadFile();
  }

  return {
    rows: fileCache.value.slice(startRow, endRow),
    startIndex: startRow,
    endIndex: endRow,
    totalRowCount: fileCache.value.length,
  };
};


const loadNewFile = async () => {
  await loadFile();
};
</script>