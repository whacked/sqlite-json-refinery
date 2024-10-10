<template>
  <div>
    <div class="filters">
      <input v-model="quickFilterText" placeholder="Quick filter..." @input="onQuickFilterChanged" />
      <button @click="expandAllRows">Expand All</button>
      <button @click="contractAllRows">Contract All</button>
    </div>
    <div class="ag-theme-alpine" style="height: 500px; width: 100%;">
      <div>
        {{ gridApi ? `${gridApi.getDisplayedRowCount()} / ${dataStore.totalRows} rows visible
        (${currentDisplayedRowsRange})` : '0 / 0 rows visible'
        }}
      </div>
      <ag-grid-vue
        style="height: 100%; border: 4px dashed orange;"
        class="ag-theme-quartz"
        :columnDefs="columnDefs"
        :rowData="rowData"
        :defaultColDef="defaultColDef"
        :components="components"

        :rowBuffer="rowBuffer"
        :rowModelType="rowModelType"
        :paginationPageSize="paginationPageSize"
        :cacheBlockSize="cacheBlockSize"
        :infiniteInitialRowCount="infiniteInitialRowCount"

        @grid-ready="onGridReady"
        @model-updated="onModelUpdated"
        @first-data-rendered="onFirstDataRendered"
        @body-scroll="onBodyScroll"
      >
      </ag-grid-vue>
    </div>
  </div>
</template>

<style scoped>
.filters {
  margin-bottom: 10px;
}

</style>

<script setup lang="ts">
import { ref, onMounted, reactive, defineComponent, h, computed } from 'vue';
import { AgGridVue } from 'ag-grid-vue3';
import { BodyScrollEvent, ColDef, GridApi, GridReadyEvent, paramValueToCss } from 'ag-grid-community';
import { useDataStore } from '@/stores/dataStore';
import PayloadCell from '@/components/PayloadCell.vue';
import * as ColumnManager from '@/utils/columnManager';

const customHeader = defineComponent({
  props: ['displayName', 'onCustomAction'],
  setup(props) {
    //@ts-ignore  this is correct, but flagged by the linter
    const params = props.params;
    const sortOrder = ref(params.column.getSort());

    const collapseKey = () => {
      const keyToRemove = params.key;
      ColumnManager.expandedKeys.value.delete(keyToRemove);
      updateColumnDefs();
    };

    const onSortClicked = (event: MouseEvent) => {
      params.progressSort(event.shiftKey);
      sortOrder.value = params.column.getSort();
    };

    params.column.addEventListener('sortChanged', () => {
      sortOrder.value = params.column.getSort();
    });

    return () => h('div', { class: 'ag-header-cell-label' }, [
      h('span', { class: 'ag-header-cell-text' }, params.displayName),
      h('button', {
        class: 'ag-my-sort-button',
        onClick: onSortClicked
      }, (
        sortOrder.value === "asc" ? '▲' : 
        sortOrder.value === "desc" ? '▼' : '⇅'
      )),
      h('button', {
        class: 'ag-my-collapse-button',
        onClick: collapseKey
      }, '✖'),
    ]);
  }
});

const components = {
  customHeader,
  payloadCellRenderer: PayloadCell,
};

const dataStore = useDataStore();

const columnDefs = ref<ColDef[]>([]);
const quickFilterText = ref('');
const gridApi = ref<GridApi | null>(null);
const rowData = ref<any[]>([]);


namespace TableRowPositionStatus {

export const totalRows = ref(0);
export const visibleRows = ref(0);
export const firstVisibleRow = ref(0);
export const lastVisibleRow = ref(0);
}


const onBodyScroll = (event: BodyScrollEvent) => {
  updateRowCount();
}

const currentDisplayedRowsRange = ref('');

const onModelUpdated = () => {
  console.log("updated")
  updateRowCount();
};

const onFirstDataRendered = () => {
  updateRowCount();
};

const updateRowCount = () => {
  if (!gridApi.value) return;
  TableRowPositionStatus.totalRows.value = dataStore.totalRows;
  TableRowPositionStatus.visibleRows.value = gridApi.value.getDisplayedRowCount();
  TableRowPositionStatus.firstVisibleRow.value = gridApi.value.getFirstDisplayedRow() + 1;
  TableRowPositionStatus.lastVisibleRow.value = gridApi.value.getLastDisplayedRow() + 1;
  currentDisplayedRowsRange.value = `${TableRowPositionStatus.firstVisibleRow.value}-${TableRowPositionStatus.lastVisibleRow.value}`;
};



const defaultColDef = reactive({
  flex: 1,
  minWidth: 100,
  sortable: true,
  filter: true,
});

const rowBuffer = 0;
const rowModelType = 'clientSide';
const paginationPageSize = 10;
const cacheBlockSize = 10;
const infiniteInitialRowCount = 1;

const onGridReady = (params: GridReadyEvent) => {
  gridApi.value = params.api;
  fetchData();
  updateRowCount();
};


const fetchData = async () => {
  const data = await dataStore.fetchData(0, dataStore.totalRows);
  rowData.value = data.map(row => ({
    ...row,
    payloadString: row.payload,
    payload: JSON.parse(row.payload)
  }));
  updateColumnDefs();
};

const updateColumnDefs = () => {
  const cellRendererParams: ColumnManager.RenderParams = {
    expandedKeys: ColumnManager.expandedKeys.value,
    toggleExpand: toggleExpandKeys,
    toggleContract: toggleContractKeys,
  };
  const baseColumns: ColDef[] = [
    { field: 'id', headerName: 'ID', width: 100 },
    { field: 'country', headerName: 'Country', width: 150 },
    { field: 'createdAt', headerName: 'Created At', width: 200 },
    { 
      field: 'payloadString', 
      headerName: 'Payload String (not shown; for filtering)', 
      hide: true 
    },
    { 
      field: 'payload', 
      headerName: 'Payload', 
      width: 300, 
      cellRenderer: 'payloadCellRenderer',
      cellRendererParams: cellRendererParams,
    },
  ];

  const derivedColumns: ColDef[] = Array.from(ColumnManager.expandedKeys.value)
    .sort((a, b) => a.localeCompare(b))
    .map(key => ({
      field: `payload.${key}`,
      headerName: key,
      width: 150,
      cellStyle: { backgroundColor: '#e6f7ff' },
      /* headerComponent: 'agColumnHeader', */
      headerComponent: customHeader,
       headerComponentParams: {
        key: key,
      },
    }));

  columnDefs.value = [...baseColumns, ...derivedColumns];

};

const toggleExpandKeys = (id: string) => {
  if (ColumnManager.expandedRows.value.has(id)) {
    ColumnManager.expandedRows.value.delete(id);
  } else {
    ColumnManager.expandedRows.value.add(id);
    const row = rowData.value.find(r => r.id === id);
    if (row) {
      Object.keys(row.payload).forEach(key => ColumnManager.expandedKeys.value.add(key));
    }
  }
  updateColumnDefs();
};

const toggleContractKeys = (id: string) => {
  if (!ColumnManager.expandedRows.value.has(id)) {
    return;
  } else {
    ColumnManager.expandedRows.value.delete(id);
    const row = rowData.value.find(r => r.id === id);
    if (row) {
      Object.keys(row.payload).forEach(key => ColumnManager.expandedKeys.value.delete(key));
    }
  }
  updateColumnDefs();
}

const expandAllRows = () => {
  rowData.value.forEach(row => {
    ColumnManager.expandedRows.value.add(row.id);
    Object.keys(row.payload).forEach(key => ColumnManager.expandedKeys.value.add(key));
  });
  updateColumnDefs();
};

const contractAllRows = () => {
  ColumnManager.expandedRows.value.clear();
  ColumnManager.expandedKeys.value.clear();
  updateColumnDefs();
};

const onQuickFilterChanged = () => {
  console.log(quickFilterText.value.trim());
  gridApi.value?.setColumnFilterModel('payloadString', {
    filterType: 'text',
    type: 'contains',
    filter: quickFilterText.value.trim(),
  });
  gridApi.value?.onFilterChanged();
};

onMounted(() => {
  fetchData();
});

</script>