<template>
  <div>
    <div class="filters">
      <input v-model="quickFilterText" placeholder="Quick filter..." @input="onQuickFilterChanged" />
      <button @click="autoSizeColumns">Fit Columns</button>
    </div>
    <div class="core-columns-control">
      <h4>Core Columns</h4>
      <div v-for="column in ColumnManager.availableColumns.value" :key="column.key" class="column-checkbox">
        <label v-if="newColumnName == '' || column.key.includes(newColumnName)">
          <input 
            type="checkbox" 
            :checked="column.isEnabled" 
            @change="toggleCoreColumn(column.key)"
          />
          {{ column.key }}
        </label>
      </div>
      <div class="add-column">
        <input v-model="newColumnName" placeholder="New column name" />
        <button
          v-if="newColumnName != '' && ColumnManager.availableColumns.value.filter(col => col.key == newColumnName).length > 0"
          @click="addCoreColumn">Add Column {{ newColumnName }}</button>
      </div>
    </div>
    <div class="ag-theme-alpine" style="height: 500px; width: 100%;">
      <div>
        {{ gridApi ? `${gridApi.getDisplayedRowCount()} / ${dataStore.totalRows} rows visible
        (${currentDisplayedRowsRange})` : '0 / 0 rows visible'
        }}; {{ totalExpandableRows }} expandable rows
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

        @keydown="onKeyDown"

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
import { BodyScrollEvent, ColDef, GetRowIdParams, GridApi, GridReadyEvent, paramValueToCss, ValueGetterParams } from 'ag-grid-community';
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

const onKeyDown = (event: KeyboardEvent) => {
  if (event.ctrlKey && event.key === 'c') {
    const cursorPosition = gridApi.value?.getFocusedCell();
    const rowIndex = cursorPosition?.rowIndex;
    const colIndex = cursorPosition?.column.getColId();
    if (rowIndex != null && colIndex != null) {
      navigator.clipboard.writeText(rowData.value[rowIndex][colIndex]);
    }
  }
}

const currentDisplayedRowsRange = ref('');

const onModelUpdated = () => {
  console.log("%cupdated", "color: red; font-weight: bold; font-size: 2em;");
  console.log(">>> detectedKeys", toRaw(detectedKeys.value));
  ColumnManager.availableColumns.value = Array.from(detectedKeys.value)
  .filter(key => (
    key !== ColumnManager.EXPANDABLE_DATA_COLUMN
  ))
  // .filter(key => ColumnManager.COMMON_COLUMN_KEYS.value.has(key))
  .map(key => ({ key, isEnabled: ColumnManager.COMMON_COLUMN_KEYS.value.has(key) }));
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


const newColumnName = ref('');

const toggleCoreColumn = (column: string) => {
  if (ColumnManager.COMMON_COLUMN_KEYS.value.has(column)) {
    ColumnManager.COMMON_COLUMN_KEYS.value.delete(column);
  } else {
    ColumnManager.COMMON_COLUMN_KEYS.value.add(column);
  }
  updateColumnDefs();
};

const addCoreColumn = () => {
  if (newColumnName.value && !ColumnManager.COMMON_COLUMN_KEYS.value.has(newColumnName.value)) {
    ColumnManager.COMMON_COLUMN_KEYS.value.add(newColumnName.value);
    newColumnName.value = ''; // Clear the input after adding
    updateColumnDefs();
  }
};

const removeColumnName = (column: string) => {
  if (ColumnManager.COMMON_COLUMN_KEYS.value.has(column)) {
    ColumnManager.COMMON_COLUMN_KEYS.value.delete(column);
    updateColumnDefs();
  }
};




const defaultColDef = reactive({
  flex: 1,
  minWidth: 100,
  sortable: true,
  filter: true,
});

const rowBuffer = 40;
const rowModelType = 'clientSide';
const cacheBlockSize = 30;
const infiniteInitialRowCount = 20;

const onGridReady = (params: GridReadyEvent) => {
  gridApi.value = params.api;
  fetchData();
  updateRowCount();
};

const autoSizeColumns = () => {
  // gridApi.value?.autoSizeColumns(Array.from(ColumnManager.COMMON_COLUMN_KEYS.value));
  gridApi.value?.autoSizeAllColumns();
}

const totalExpandableRows = ref<number>(0)
const detectedKeys = ref(new Set<string>());
const fetchData = async () => {
  const sourceData = props.rowData?.length > 0 ? props.rowData : (await dataStore.fetchData(0, dataStore.totalRows));
  rowData.value = sourceData.map(row => {
    Object.keys(row).forEach(key => detectedKeys.value.add(key));
    if (row[ColumnManager.EXPANDABLE_DATA_COLUMN]) {
      totalExpandableRows.value++;
    }
    return {
      ...row,
      [ColumnManager.EXPANDABLE_DATA_COLUMN_SHADOW]: row[ColumnManager.EXPANDABLE_DATA_COLUMN],
      [ColumnManager.EXPANDABLE_DATA_COLUMN]: row[ColumnManager.EXPANDABLE_DATA_COLUMN] ? JSON.parse(row[ColumnManager.EXPANDABLE_DATA_COLUMN]) : null,
    }
  });
  updateColumnDefs();
};

const updateColumnDefs = () => {
  const cellRendererParams: ColumnManager.RenderParams = {
    coreDisplayParams: ColumnManager.COMMON_COLUMN_KEYS.value,
    expandableDataExtractedKeys: ColumnManager.expandedExpandableDataKeys.value,
    collapsedDataKeys: ColumnManager.collapsableDataExtractedKeys.value,
    toggleExpandCollapsibleKeys,
    toggleContractCollapsibleKeys,
    toggleExpandExpandableKeys,
    toggleContractExpandableKeys,
  };

  const coreDisplayParamSettings: Record<string, ColDef> = {
    id: { field: 'id', headerName: 'ID', width: 100 },
    country: { field: 'country', headerName: 'Country', width: 150 },
    createdAt: { field: 'createdAt', headerName: 'Created At', width: 200 },
    [ColumnManager.COLLAPSABLE_DATA_COLUMN_SHADOW]: {
      field: ColumnManager.COLLAPSABLE_DATA_COLUMN_SHADOW,
      headerName: 'collapsed data String (not shown; for filtering)',
      hide: true,
    },
    [ColumnManager.COLLAPSABLE_DATA_COLUMN]: { 
      field: ColumnManager.COLLAPSABLE_DATA_COLUMN, 
      headerName: 'Collapsed Data',
      width: 200,
      headerClass: 'my-ag-table-collapsible-data-header',
      cellRenderer: 'extractedDataCellRenderer',
      cellClass: 'my-ag-table-collapsible-data-cell',
      cellRendererParams: cellRendererParams,
      headerComponent: defineComponent({
        props: ['displayName', 'onCustomAction'],
        setup(props) {
          //@ts-ignore  this is correct, but flagged by the linter
          const params = props.params;

          return () => h('div', { class: 'my-ag-table-collapsible-data-header' }, [
            h('div', {}, [
              h('div', { class: 'ag-header-cell-text' }, params.displayName),
            ]),
            h('div', { class: 'button-container' }, [
              h('div', {}, [
                ColumnManager.collapsableDataExtractedKeys.value.size > 0 && h('button', {
                  class: 'ag-my collapse all',
                  title: 'Collapse all',
                  onClick: collapseAllCollapsibleRows
                }, '◀'),
                h('button', {
                  class: 'ag-my expand all',
                  title: 'Expand all',
                  onClick: restoreAllCollapsedRows
                }, '▶'),
              ]),
            ]),
          ]);
        }
      }),
    },
    [ColumnManager.EXPANDABLE_DATA_COLUMN_SHADOW]: { 
      field: ColumnManager.EXPANDABLE_DATA_COLUMN_SHADOW,
      headerName: 'Expandable Data String (not shown; for filtering)', 
      hide: true 
    },
  };

  const baseColumns: ColDef[] = (
    Array.from(ColumnManager.COMMON_COLUMN_KEYS.value)
    .filter(key => detectedKeys.value.has(key))
    .concat([
      ColumnManager.COLLAPSABLE_DATA_COLUMN,
      // needed for filtering
      ColumnManager.EXPANDABLE_DATA_COLUMN_SHADOW,
      ColumnManager.COLLAPSABLE_DATA_COLUMN_SHADOW,
    ])
    .map(key => coreDisplayParamSettings[key] ?? {
      field: key,
      headerName: key,
    })
    .filter(colDef => colDef)
  );

  const collapsedDataExtractedColumns: ColDef[] = Array.from(ColumnManager.collapsableDataExtractedKeys.value)
    .sort((a, b) => a.localeCompare(b))
    .map(key => ({
      field: key,
      headerName: `${key}`,
      headerClass: 'my-ag-table-collapsible-data-expanded-column',
      cellClass: 'my-ag-table-collapsible-data-expanded-cell',
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
      headerClass: 'my-ag-table-expandable-data-expanded-column',
      cellClass: 'my-ag-table-expandable-data-expanded-cell',
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
    ...(totalExpandableRows.value > 0 ? [{ 
      field: ColumnManager.EXPANDABLE_DATA_COLUMN, 
      headerName: 'Expandable Data', 
      width: 300, 
      cellRenderer: 'payloadCellRenderer',
      headerClass: 'my-ag-table-expandable-data-header',
      cellClass: 'my-ag-table-expandable-data-cell',
      cellRendererParams: cellRendererParams,
      headerComponent: defineComponent({
        props: ['displayName', 'onCustomAction'],
        setup(props) {
          //@ts-ignore  this is correct, but flagged by the linter
          const params = props.params;

          return () => h('div', { class: 'my-ag-table-expandable-data-header' }, [
            h('div', {}, [
              h('div', { class: 'ag-header-cell-text' }, params.displayName),
            ]),
            h('div', { class: 'button-container' }, [
              h('div', {}, [
                ColumnManager.expandedExpandableDataKeys.value.size > 0 && h('button', {
                  class: 'ag-my collapse all',
                  title: 'Collapse all',
                  onClick: contractAllExpandableRows
                }, '◀'),
                h('button', {
                  class: 'ag-my expand all',
                  title: 'Expand all',
                  onClick: expandAllExpandableRows
                }, '▶'),
              ]),
            ]),
          ]);
        }
      }),
    }] : []),
    ...expandableDataExtractedColumns,
  ];

};

const toggleExpandExpandableKeys = (rowIndex: number) => {
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

const toggleContractExpandableKeys = (rowIndex: number) => {
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
        key => !ColumnManager.COMMON_COLUMN_KEYS.value.has(key)
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
        key => (
          !ColumnManager.COMMON_COLUMN_KEYS.value.has(key) &&
          !ColumnManager.SPECIAL_COLUMN_KEYS.has(key)
        )
      ).forEach(key => ColumnManager.collapsableDataExtractedKeys.value.delete(key));
    }
  }
  updateColumnDefs();
}

const restoreAllCollapsedRows = () => {
  rowData.value.forEach(row => {
    ColumnManager.collapsableDataExpandedRows.value.add(row.id);
    Object.keys(row).filter(
      key => (
        !ColumnManager.COMMON_COLUMN_KEYS.value.has(key) &&
        !ColumnManager.SPECIAL_COLUMN_KEYS.has(key)
      )
    ).forEach(key => ColumnManager.collapsableDataExtractedKeys.value.add(key));
  });
  updateColumnDefs();
};

const collapseAllCollapsibleRows = () => {
  ColumnManager.collapsableDataExpandedRows.value.clear();
  ColumnManager.collapsableDataExtractedKeys.value.clear();
  updateColumnDefs();
};


const expandAllExpandableRows = () => {
  let isDirty = false;
  rowData.value.forEach(row => {
    if (row.payload) {
      const payloadKeys = Object.keys(row.payload);
      if (payloadKeys.length > 0) {
        ColumnManager.expandedExpandableDataRows.value.add(row.id);
        payloadKeys.forEach(key => ColumnManager.expandedExpandableDataKeys.value.add(key));
        isDirty = true;
      }
    }
  });
  if (isDirty) {
    updateColumnDefs();
  }
};

const contractAllExpandableRows = () => {
  ColumnManager.expandedExpandableDataKeys.value.clear();
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