<template>
  <div>
    <div class="filters">
      <input v-model="quickFilterText" placeholder="Quick filter..." @input="onQuickFilterChanged" />
      <button @click="expandAllPayloadRows">Expand All</button>
      <button @click="contractAllPayloadRows">Contract All</button>
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
import { ref, onMounted, reactive, defineComponent, h, computed, Ref } from 'vue';
import { AgGridVue } from 'ag-grid-vue3';
import { BodyScrollEvent, ColDef, GridApi, GridReadyEvent, paramValueToCss } from 'ag-grid-community';
import { useDataStore } from '@/stores/dataStore';
import PayloadCell from '@/components/PayloadCell.vue';
import ExtractedDataCell from '@/components/ExtractedDataCell.vue';
import * as ColumnManager from '@/utils/columnManager';

function makeExpandedColumnHeader(keyTrackerProxy: Ref<Set<string>>) {
  return defineComponent({
    props: ['displayName', 'onCustomAction'],
    setup(props) {
      //@ts-ignore  this is correct, but flagged by the linter
      const params = props.params;
      const sortOrder = ref(params.column.getSort());

      const collapseKey = () => {
        const keyToRemove = params.key;
        keyTrackerProxy.value.delete(keyToRemove);
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
  })
}

const expandedPayloadColumnHeader = makeExpandedColumnHeader(ColumnManager.expandedPayloadKeys);
const expandedExtraDataColumnHeader = makeExpandedColumnHeader(ColumnManager.extractedDataExtractedKeys);

const components = {
  expandedPayloadColumnHeader,
  expandedExtraDataColumnHeader,
  payloadCellRenderer: PayloadCell,
  extractedDataCellRenderer: ExtractedDataCell,
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
    coreDisplayParams: ColumnManager.coreDisplayParams.value,
    payloadExtractedKeys: ColumnManager.expandedPayloadKeys.value,
    extraDataExtractedKeys: ColumnManager.extractedDataExtractedKeys.value,
    toggleExpandExtraDataKeys: toggleExpandExtraDataKeys,
    toggleContractExtraDataKeys: toggleContractExtraDataKeys,
    toggleExpandPayloadKeys: toggleExpandPayloadKeys,
    toggleContractPayloadKeys: toggleContractPayloadKeys,
  };

  const coreDisplayParamSettings: Record<string, ColDef> = {
    id: { field: 'id', headerName: 'ID', width: 100 },
    country: { field: 'country', headerName: 'Country', width: 150 },
    createdAt: { field: 'createdAt', headerName: 'Created At', width: 200 },
    extraDataString: {
      field: 'extraDataString',
      headerName: 'Extra Data String (not shown; for filtering)',
      hide: true,
    },
    extraData: { 
      field: 'extraData', 
      headerName: 'Extra Data', 
      width: 200,
      headerClass: 'my-ag-table-extra-data-header',
      cellRenderer: 'extractedDataCellRenderer',
      cellClass: 'my-ag-table-extra-data-cell',
      cellRendererParams: cellRendererParams,
    },
    payloadString: { 
      field: 'payloadString', 
      headerName: 'Payload String (not shown; for filtering)', 
      hide: true 
    },
  };
  const baseColumns: ColDef[] = (
    Array.from(ColumnManager.coreDisplayParams.value)
    .filter(key => coreDisplayParamSettings[key])
    .map(key => coreDisplayParamSettings[key])
  );

  const payloadExtractedColumns: ColDef[] = Array.from(ColumnManager.expandedPayloadKeys.value)
    .sort((a, b) => a.localeCompare(b))
    .map(key => ({
      field: `payload.${key}`,
      headerName: key,
      width: 150,
      headerClass: 'my-ag-table-payload-expanded-column',
      cellClass: 'my-ag-table-payload-expanded-cell',
      headerComponent: expandedPayloadColumnHeader,
       headerComponentParams: {
        key: key,
      },
    }));

  const extraDataExtractedColumns: ColDef[] = Array.from(ColumnManager.extractedDataExtractedKeys.value)
    .sort((a, b) => a.localeCompare(b))
    .map(key => ({
      field: key,
      headerName: `${key}`,
      width: 150,
      headerClass: 'my-ag-table-extra-data-expanded-column',
      cellClass: 'my-ag-table-extra-data-expanded-cell',
      headerComponent: expandedExtraDataColumnHeader,
      headerComponentParams: {
        key: key,
      },
    }));

  columnDefs.value = [
    ...baseColumns,
    ...extraDataExtractedColumns,
    { 
      field: 'payload', 
      headerName: 'Payload', 
      width: 300, 
      cellRenderer: 'payloadCellRenderer',
      headerClass: 'my-ag-table-payload-header',
      cellClass: 'my-ag-table-payload-cell',
      cellRendererParams: cellRendererParams,
    },
    ...payloadExtractedColumns,
  ];

};

const toggleExpandPayloadKeys = (id: string) => {
  if (ColumnManager.expandedPayloadRows.value.has(id)) {
    ColumnManager.expandedPayloadRows.value.delete(id);
  } else {
    ColumnManager.expandedPayloadRows.value.add(id);
    const row = rowData.value.find(r => r.id === id);
    if (row) {
      Object.keys(row.payload).forEach(key => ColumnManager.expandedPayloadKeys.value.add(key));
    }
  }
  updateColumnDefs();
};

const toggleContractPayloadKeys = (id: string) => {
  if (!ColumnManager.expandedPayloadRows.value.has(id)) {
    return;
  } else {
    ColumnManager.expandedPayloadRows.value.delete(id);
    const row = rowData.value.find(r => r.id === id);
    if (row) {
      Object.keys(row.payload).forEach(key => ColumnManager.expandedPayloadKeys.value.delete(key));
    }
  }
  updateColumnDefs();
}

const toggleExpandExtraDataKeys = (id: string) => {
  if (ColumnManager.extractedDataExpandedRows.value.has(id)) {
    ColumnManager.extractedDataExpandedRows.value.delete(id);
  } else {
    ColumnManager.extractedDataExpandedRows.value.add(id);
    const row = rowData.value.find(r => r.id === id);
    if (row) {
      Object.keys(row).filter(
        key => !ColumnManager.coreDisplayParams.value.has(key)
      ).forEach(key => ColumnManager.extractedDataExtractedKeys.value.add(key));
    }
  }
  updateColumnDefs();
}

const toggleContractExtraDataKeys = (id: string) => {
  if (!ColumnManager.extractedDataExpandedRows.value.has(id)) {
    return;
  } else {
    ColumnManager.extractedDataExpandedRows.value.delete(id);
    const row = rowData.value.find(r => r.id === id);
    if (row) {
      Object.keys(row).filter(
        key => !ColumnManager.coreDisplayParams.value.has(key)
      ).forEach(key => ColumnManager.extractedDataExtractedKeys.value.delete(key));
    }
  }
  updateColumnDefs();
}

const expandAllPayloadRows = () => {
  rowData.value.forEach(row => {
    ColumnManager.expandedPayloadRows.value.add(row.id);
    Object.keys(row.payload).forEach(key => ColumnManager.expandedPayloadKeys.value.add(key));
  });
  updateColumnDefs();
};

const contractAllPayloadRows = () => {
  ColumnManager.expandedPayloadRows.value.clear();
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