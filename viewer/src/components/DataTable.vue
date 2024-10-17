<template>
  <div>
    <div class="filters">
      <input v-model="quickFilterText" placeholder="Quick filter..." @input="onQuickFilterChanged" />
      <button @click="expandAllExpandableRows">Expand All</button>
      <button @click="contractAllExpandableRows">Contract All</button>
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

const expandedPayloadColumnHeader = makeExpandedColumnHeader(ColumnManager.expandedExpandableDataKeys);
const expandedCollapsibleDataColumnHeader = makeExpandedColumnHeader(ColumnManager.collapsableDataExtractedKeys);

const components = {
  expandedPayloadColumnHeader,
  expandedExtraDataColumnHeader: expandedCollapsibleDataColumnHeader,
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
    [ColumnManager.EXPANDABLE_DATA_COLUMN_SHADOW]: row[ColumnManager.EXPANDABLE_DATA_COLUMN],
    [ColumnManager.EXPANDABLE_DATA_COLUMN]: row[ColumnManager.EXPANDABLE_DATA_COLUMN] ? JSON.parse(row[ColumnManager.EXPANDABLE_DATA_COLUMN]) : null,
  }));
  updateColumnDefs();
};

const updateColumnDefs = () => {
  const cellRendererParams: ColumnManager.RenderParams = {
    coreDisplayParams: ColumnManager.coreDisplayParams.value,
    expandableDataExtractedKeys: ColumnManager.expandedExpandableDataKeys.value,
    collapsedDataKeys: ColumnManager.collapsableDataExtractedKeys.value,
    toggleExpandCollapsableDataKeys: toggleExpandCollapsibleKeys,
    toggleContractCollapsableDataKeys: toggleContractCollapsibleKeys,
    toggleExpandExpandableKeys: toggleExpandPayloadKeys,
    toggleContractExpandableKeys: toggleContractPayloadKeys,
  };

  const coreDisplayParamSettings: Record<string, ColDef> = {
    id: { field: 'id', headerName: 'ID', width: 100 },
    country: { field: 'country', headerName: 'Country', width: 150 },
    createdAt: { field: 'createdAt', headerName: 'Created At', width: 200 },
    [ColumnManager.COLLAPSABLE_DATA_COLUMN_SHADOW]: {
      field: ColumnManager.COLLAPSABLE_DATA_COLUMN_SHADOW,
      headerName: 'Extra Data String (not shown; for filtering)',
      hide: true,
    },
    [ColumnManager.COLLAPSABLE_DATA_COLUMN]: { 
      field: ColumnManager.COLLAPSABLE_DATA_COLUMN, 
      headerName: 'Extra Data',
      width: 200,
      headerClass: 'my-ag-table-extra-data-header',
      cellRenderer: 'extractedDataCellRenderer',
      cellClass: 'my-ag-table-extra-data-cell',
      cellRendererParams: cellRendererParams,
    },
    [ColumnManager.EXPANDABLE_DATA_COLUMN_SHADOW]: { 
      field: ColumnManager.EXPANDABLE_DATA_COLUMN_SHADOW,
      headerName: 'Expandable Data String (not shown; for filtering)', 
      hide: true 
    },
  };
  const baseColumns: ColDef[] = (
    Array.from(ColumnManager.coreDisplayParams.value)
    .filter(key => key !== ColumnManager.EXPANDABLE_DATA_COLUMN)  // this one is handled separately
    .map(key => coreDisplayParamSettings[key] ?? {
      field: key,
      headerName: key,
    })
    .filter(colDef => colDef)
  );

  console.log("extractedDataExtractedKeys", toRaw(ColumnManager.collapsableDataExtractedKeys.value));
  const collapsedDataExtractedColumns: ColDef[] = Array.from(ColumnManager.collapsableDataExtractedKeys.value)
    .sort((a, b) => a.localeCompare(b))
    .map(key => ({
      field: key,
      headerName: `${key}`,
      width: 150,
      headerClass: 'my-ag-table-extra-data-expanded-column',
      cellClass: 'my-ag-table-extra-data-expanded-cell',
      cellRenderer: (params: ColumnManager.RenderParams, _: any) => {
        return params.data[key]?.toString()
      },
      headerComponent: expandedCollapsibleDataColumnHeader,
      headerComponentParams: {
        key: key,
      },
    }));

    const expandableDataExtractedColumns: ColDef[] = Array.from(ColumnManager.expandedExpandableDataKeys.value)
    .sort((a, b) => a.localeCompare(b))
    .map(key => ({
      field: `${ColumnManager.EXPANDABLE_DATA_COLUMN}.${key}`,
      headerName: key,
      width: 150,
      headerClass: 'my-ag-table-payload-expanded-column',
      cellClass: 'my-ag-table-payload-expanded-cell',
      headerComponent: expandedPayloadColumnHeader,
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
      flex: 0,
      width: 40,
    },
    ...baseColumns,
    ...collapsedDataExtractedColumns,
    { 
      field: ColumnManager.EXPANDABLE_DATA_COLUMN, 
      headerName: 'Payload (expandable)', 
      width: 300, 
      cellRenderer: 'payloadCellRenderer',
      headerClass: 'my-ag-table-payload-header',
      cellClass: 'my-ag-table-payload-cell',
      cellRendererParams: cellRendererParams,
    },
    ...expandableDataExtractedColumns,
  ];

};

const toggleExpandPayloadKeys = (rowIndex: number) => {
  if (ColumnManager.expandedExpandableDataRows.value.has(rowIndex)) {
    ColumnManager.expandedExpandableDataRows.value.delete(rowIndex);
  } else {
    ColumnManager.expandedExpandableDataRows.value.add(rowIndex);
    const row = rowData.value[rowIndex];
    if (row) {
      Object.keys(row.payload).forEach(key => ColumnManager.expandedExpandableDataKeys.value.add(key));
    }
  }
  updateColumnDefs();
};

const toggleContractPayloadKeys = (rowIndex: number) => {
  if (!ColumnManager.expandedExpandableDataRows.value.has(rowIndex)) {
    return;
  } else {
    ColumnManager.expandedExpandableDataRows.value.delete(rowIndex);
    const row = rowData.value[rowIndex];
    if (row) {
      Object.keys(row.payload).forEach(key => ColumnManager.expandedExpandableDataKeys.value.delete(key));
    }
  }
  updateColumnDefs();
}

const toggleExpandCollapsibleKeys = (rowIndex: number) => {
  console.log("toggleExpandExtraDataKeys", rowIndex);
  if (ColumnManager.collapsableDataExpandedRows.value.has(rowIndex)) {
    ColumnManager.collapsableDataExpandedRows.value.delete(rowIndex);
  } else {
    ColumnManager.collapsableDataExpandedRows.value.add(rowIndex);
    const row = rowData.value[rowIndex];
    if (row) {
      console.log("row", row);
      Object.keys(row).filter(
        key => !ColumnManager.coreDisplayParams.value.has(key)
      ).forEach(key => {
        console.log("adding key", key);
        ColumnManager.collapsableDataExtractedKeys.value.add(key)
      });
    }
  }
  updateColumnDefs();
}

const toggleContractCollapsibleKeys = (rowIndex: number) => {
  if (!ColumnManager.collapsableDataExpandedRows.value.has(rowIndex)) {
    return;
  } else {
    ColumnManager.collapsableDataExpandedRows.value.delete(rowIndex);
    const row = rowData.value[rowIndex];
    if (row) {
      Object.keys(row).filter(
        key => !ColumnManager.coreDisplayParams.value.has(key)
      ).forEach(key => ColumnManager.collapsableDataExtractedKeys.value.delete(key));
    }
  }
  updateColumnDefs();
}

const expandAllExpandableRows = () => {
  rowData.value.forEach(row => {
    ColumnManager.expandedExpandableDataRows.value.add(row.id);
    Object.keys(row.payload).forEach(key => ColumnManager.expandedExpandableDataKeys.value.add(key));
  });
  updateColumnDefs();
};

const contractAllExpandableRows = () => {
  ColumnManager.expandedExpandableDataRows.value.clear();
  updateColumnDefs();
};

const onQuickFilterChanged = () => {
  console.log(quickFilterText.value.trim());
  gridApi.value?.setColumnFilterModel(ColumnManager.EXPANDABLE_DATA_COLUMN_SHADOW, {
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