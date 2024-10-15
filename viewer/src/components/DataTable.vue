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
        colon-paginationPageSize="paginationPageSize"
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
import { ref, onMounted, reactive, defineComponent, h, computed, Ref, toRaw } from 'vue';
import { AgGridVue } from 'ag-grid-vue3';
import { BodyScrollEvent, ColDef, GridApi, GridReadyEvent, paramValueToCss, ValueGetterParams } from 'ag-grid-community';
import { useDataStore } from '@/stores/dataStore';
import ExpandableCell from '@/components/ExpandableCell.vue';
import CollapsableCell from '@/components/CollapsableCell.vue';
import * as ColumnManager from '@/utils/columnManager';

const props = defineProps<{
  rowData: any[];
}>();

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
  payloadCellRenderer: ExpandableCell,
  extractedDataCellRenderer: CollapsableCell,
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

const rowBuffer = 40;
const rowModelType = 'clientSide';
// const rowModelType = 'infinite';
const paginationPageSize = 10;
const cacheBlockSize = 30;
const infiniteInitialRowCount = 20;

const onGridReady = (params: GridReadyEvent) => {
  gridApi.value = params.api;
  fetchData();
  updateRowCount();
};


const fetchData = async () => {
  const sourceData = props.rowData?.length > 0 ? props.rowData : (await dataStore.fetchData(0, dataStore.totalRows));
  rowData.value = sourceData.map(row => ({
    ...row,
    payloadString: row.payload,
    payload: row.payload ? JSON.parse(row.payload) : null,
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
    .filter(key => key !== 'payload')  // this one is handled separately
    .map(key => coreDisplayParamSettings[key] ?? {
      field: key,
      headerName: key,
    })
    .filter(colDef => colDef)
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

  console.log("extractedDataExtractedKeys", toRaw(ColumnManager.extractedDataExtractedKeys.value));
  const extraDataExtractedColumns: ColDef[] = Array.from(ColumnManager.extractedDataExtractedKeys.value)
    .sort((a, b) => a.localeCompare(b))
    .map(key => ({
      field: key,
      headerName: `${key}`,
      width: 150,
      headerClass: 'my-ag-table-extra-data-expanded-column',
      cellClass: 'my-ag-table-extra-data-expanded-cell',
      cellRenderer: (params, _) => {
        return params.data[key]?.toString()
      },
      headerComponent: expandedExtraDataColumnHeader,
      headerComponentParams: {
        key: key,
      },
    }));

  columnDefs.value = [
    {
      headerName: "#",
      valueGetter: (params: ValueGetterParams) => {
        return (params.node?.rowIndex ?? 0) + 1;
      },
      width: 80,
    },
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

const toggleExpandPayloadKeys = (rowIndex: number) => {
  if (ColumnManager.expandedPayloadRows.value.has(rowIndex)) {
    ColumnManager.expandedPayloadRows.value.delete(rowIndex);
  } else {
    ColumnManager.expandedPayloadRows.value.add(rowIndex);
    const row = rowData.value[rowIndex];
    if (row) {
      Object.keys(row.payload).forEach(key => ColumnManager.expandedPayloadKeys.value.add(key));
    }
  }
  updateColumnDefs();
};

const toggleContractPayloadKeys = (rowIndex: number) => {
  if (!ColumnManager.expandedPayloadRows.value.has(rowIndex)) {
    return;
  } else {
    ColumnManager.expandedPayloadRows.value.delete(rowIndex);
    const row = rowData.value[rowIndex];
    if (row) {
      Object.keys(row.payload).forEach(key => ColumnManager.expandedPayloadKeys.value.delete(key));
    }
  }
  updateColumnDefs();
}

const toggleExpandExtraDataKeys = (rowIndex: number) => {
  console.log("toggleExpandExtraDataKeys", rowIndex);
  if (ColumnManager.extractedDataExpandedRows.value.has(rowIndex)) {
    ColumnManager.extractedDataExpandedRows.value.delete(rowIndex);
  } else {
    ColumnManager.extractedDataExpandedRows.value.add(rowIndex);
    const row = rowData.value[rowIndex];
    if (row) {
      console.log("row", row);
      Object.keys(row).filter(
        key => !ColumnManager.coreDisplayParams.value.has(key)
      ).forEach(key => {
        console.log("adding key", key);
        ColumnManager.extractedDataExtractedKeys.value.add(key)
      });
    }
  }
  updateColumnDefs();
}

const toggleContractExtraDataKeys = (rowIndex: number) => {
  if (!ColumnManager.extractedDataExpandedRows.value.has(rowIndex)) {
    return;
  } else {
    ColumnManager.extractedDataExpandedRows.value.delete(rowIndex);
    const row = rowData.value[rowIndex];
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