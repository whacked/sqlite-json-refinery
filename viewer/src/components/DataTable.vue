<template>
  <div class="main-grid-container">
    <div class="filters">
      <input v-model="quickFilterText" placeholder="Quick filter..." @input="onQuickFilterChanged" />
      <button @click="autoSizeColumns">Fit Columns</button>
    </div>
    <div class="core-columns-control">
      <details>
        <summary>Core Columns</summary>
        <div class="add-column">
          <input v-model="newColumnName" placeholder="New column name" />
          <button
            v-if="newColumnName != '' && ColumnManager.availableColumns.value.filter(col => col.key == newColumnName).length > 0"
            @click="addCoreColumn">Add Column {{ newColumnName }}</button>
        </div>
        <div
          v-for="column in ColumnManager.availableColumns.value.sort((a, b) => a.key.localeCompare(b.key))"
          :key="column.key" class="column-checkbox"
        >
          <label v-if="newColumnName == '' || column.key.includes(newColumnName)">
            <input 
              type="checkbox" 
              :checked="column.isEnabled" 
              @change="toggleCoreColumn(column.key)"
            />
            {{ column.key }}
          </label>
        </div>
      </details>
    </div>
    <div class="ag-theme-alpine">
      <div>
        {{ gridApi ? `${gridApi.getDisplayedRowCount()} / ${dataStore.totalRows} rows visible
        (${currentDisplayedRowsRange})` : '0 / 0 rows visible'
        }}; {{ totalExpandableRows }} expandable rows
      </div>
      <table>
        <tbody>
          <tr>
            <th>rows</th><td>{{ props.rowData.length }}</td>
          </tr>

          <tr>
            <th>expandable keys</th><td>{{ ColumnManager.expandableDataDetectedKeys.value }}</td>
          </tr>
          <tr>
            <th>unexpanded keys</th><td>{{ ColumnManager.expandableDataUnexpandedKeys.value }}</td>
          </tr>
          <tr>
            <th>expanded keys</th><td>{{ ExpandableDataManager.expandedExpandableDataKeys.value }}</td>
          </tr>

          <tr>
            <th>collapsible keys</th><td>{{ CollapsibleDataManager.collapsibleDataDetectedKeys.value }}</td>
          </tr>
          <tr>
            <th>uncollapsed keys</th><td>{{ CollapsibleDataManager.collapsibleDataExpandedKeys.value }}</td>
          </tr>
          <tr>
            <th>collapsed keys</th><td>{{ CollapsibleDataManager.collapsibleDataCollapsedKeys.value }}</td>
          </tr>
        </tbody>
      </table>
      <!--
      NOTE: rowModelType: infinite is not supported with rowData
      -->
      <ag-grid-vue
        v-if="props.rowData.length == 0"

        class="ag-theme-quartz main-grid"
        :columnDefs="columnDefs"
        :defaultColDef="defaultColDef"
        :components="components"
        rowModelType="infinite"
        :datasource="dataSource"
        :cacheBlockSize="cacheBlockSize"
        :infiniteInitialRowCount="infiniteInitialRowCount"
        :maxBlocksInCache="maxBlocksInCache"
        :rowBuffer="rowBuffer"
        :rowHeight="ColumnManager.COMMON_COLUMN_KEYS.value.has('photo') || ColumnManager.collapsableDataExtractedKeys.value.has('photo') ? 200 : null"
        @grid-ready="onGridReady"
        @model-updated="onModelUpdated"
        @first-data-rendered="onFirstDataRendered"
        @body-scroll="onBodyScroll"
      >
      </ag-grid-vue>

      <ag-grid-vue
        v-if="props.rowData.length > 0"

        class="ag-theme-quartz main-grid"
        :columnDefs="columnDefs"
        :rowData="rowData"
        :defaultColDef="defaultColDef"
        :components="components"

        :rowBuffer="rowBuffer"
        :rowModelType="rowModelType"
        :rowHeight="ColumnManager.COMMON_COLUMN_KEYS.value.has('photo') || ColumnManager.collapsableDataExtractedKeys.value.has('photo') ? 200 : null"

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

.main-grid-container {
  width: 100%;
  height: 80em;
  background: beige;
}

.ag-theme-alpine {
  height: 60em;
}

.main-grid {
  background: orange;
  height: 100%;
}

.add-column {
  display: inline-block;
}

.filters {
  margin-bottom: 10px;
}

</style>

<script setup lang="ts">
import { ref, onMounted, reactive, defineComponent, h, Ref, toRaw, watch } from 'vue';
import { AgGridVue } from 'ag-grid-vue3';
import { BodyScrollEvent, ColDef, GridApi, GridReadyEvent, ValueGetterParams } from 'ag-grid-community';
import { useDataStore } from '@/stores/dataStore';
import ExpandableCell from '@/components/ExpandableCell.vue';
import CollapsableCell from '@/components/CollapsableCell.vue';
import PhotoCell from '@/components/PhotoCell.vue';
import * as ColumnManager from '@/utils/columnManager';
import { loadRemoteData } from '@/stores/remoteDataLoader';
import chroma from 'chroma-js';
import TimeCell from '@/components/TimeCell.vue';
import ColorizedCategoricalCell from '@/components/ColorizedCategoricalCell.vue';
import { colorHash } from './styling';
import { CollapsibleDataManager, ExpandableDataManager } from '@/utils/columnManager';


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

      return () => h('div', {
        class: 'ag-header-cell-label my-ag-table-expanded-column-header',
      }, [
        ...(
          params.displayName.split('.').map((word: string, index: number) => {
            return h('span', {
              class: 'ag-header-cell-text',
              style: { background: colorHash.hex(word) },
            }, index > 0 ? '.' + word : word)
          })
        ),
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
const expandedCollapsibleDataColumnHeader = makeExpandedColumnHeader(ColumnManager.collapsibleDataExpandedKeys);

const components = {
  expandedPayloadColumnHeader,
  expandedExtraDataColumnHeader: expandedCollapsibleDataColumnHeader,
  payloadCellRenderer: ExpandableCell,
  extractedDataCellRenderer: CollapsableCell,
  photoCellRenderer: PhotoCell,
};

const dataStore = useDataStore();

const columnDefs = ref<ColDef[]>([]);
const quickFilterText = ref('');
const gridApi = ref<GridApi | null>(null);
const rowData = ref<any[]>([]);

const maxBlocksInCache = ref(100);

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
    const colId = cursorPosition?.column.getColId();

    if (rowIndex != null && colId != null) {
      // Get the displayed row corresponding to the focused cell
      const displayedRow = gridApi.value?.getDisplayedRowAtIndex(rowIndex);

      if (displayedRow) {
        // Retrieve the cell value from the displayed row
        const cellValue = displayedRow.data[colId] ?? JSON.stringify(displayedRow.data);
        navigator.clipboard.writeText(cellValue);
      }
    }
  }
}


const currentDisplayedRowsRange = ref('');

const onModelUpdated = () => {
  console.log("%cupdated", "color: red; font-weight: bold; font-size: 2em;");
  console.log(">>> detectedKeys", toRaw(CollapsibleDataManager.collapsibleDataDetectedKeys.value));
  ColumnManager.availableColumns.value = Array.from(CollapsibleDataManager.collapsibleDataDetectedKeys.value)
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
  processRows();
};

const autoSizeColumns = () => {
  // gridApi.value?.autoSizeColumns(Array.from(ColumnManager.COMMON_COLUMN_KEYS.value));
  gridApi.value?.autoSizeAllColumns();
}

const totalExpandableRows = ref<number>(0)
const coreDetectedKeys = ref<Set<string>>(new Set());
const processRows = async () => {
  const sourceData = props.rowData?.length > 0 ? props.rowData : (await dataStore.fetchData(0, dataStore.totalRows));
  CollapsibleDataManager.clearAll();
  ExpandableDataManager.clearAll();

  rowData.value = sourceData.map(row => {
    Object.keys(row).forEach(key => {
      if (ColumnManager.COMMON_COLUMN_KEYS.value.has(key)) {
        coreDetectedKeys.value.add(key);
      } else {
        CollapsibleDataManager.collapsibleDataDetectedKeys.value.add(key);
      }
    });
    let expandableData: object | null = null;
    if (row[ColumnManager.EXPANDABLE_DATA_COLUMN]) {
      totalExpandableRows.value++;
      expandableData = JSON.parse(row[ColumnManager.EXPANDABLE_DATA_COLUMN]);
      if (expandableData) {
        Object.keys(expandableData).forEach(key => {
          ExpandableDataManager.expandableDataDetectedKeys.value.add(key);
          ExpandableDataManager.expandableDataUnexpandedKeys.value.add(key);
        });
      }
    }
    CollapsibleDataManager.resetKeys()
    return {
      ...row,
      [ColumnManager.COLLAPSABLE_DATA_COLUMN_SHADOW]: JSON.stringify(ColumnManager.objectWithoutKeys(row, ColumnManager.SPECIAL_COLUMN_KEYS)),
      [ColumnManager.EXPANDABLE_DATA_COLUMN_SHADOW]: row[ColumnManager.EXPANDABLE_DATA_COLUMN],
      [ColumnManager.EXPANDABLE_DATA_COLUMN]: expandableData,
    }
  });
  updateColumnDefs();
  updateRowCount();
};

const updateColumnDefs = () => {
  const cellRendererParams: ColumnManager.RenderParams = {
    coreDisplayParams: ColumnManager.COMMON_COLUMN_KEYS.value,
    expandableDataExtractedKeys: ExpandableDataManager.expandedExpandableDataKeys.value,
    collapsedDataKeys: CollapsibleDataManager.collapsibleDataExpandedKeys.value,
    toggleExpandCollapsibleKeys,
    toggleContractCollapsibleKeys,
    toggleExpandExpandableKeys,
    toggleContractExpandableKeys,
  };

  const coreDisplayParamSettings: Record<string, ColDef> = {
    id: { field: 'id', headerName: 'ID', width: 100 },
    country: { field: 'country', headerName: 'Country', width: 150 },
    createdAt: { field: 'createdAt', headerName: 'Created At', width: 200 },
    photo: {
      field: 'photo',
      headerName: 'Photo',
      width: 100,
      cellRenderer: 'photoCellRenderer',
    },
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
              h('div', { class: 'ag-header-cell-text' },
                CollapsibleDataManager.collapsibleDataCollapsedKeys.value.size == 0 ? "" : params.displayName
              ),
            ]),
            h('div', { class: 'button-container' }, [
              h('div', {}, [
                CollapsibleDataManager.collapsibleDataExpandedKeys.value.size > 0 && h('button', {
                  class: 'ag-my collapse all',
                  title: 'Collapse all',
                  onClick: collapseAllCollapsibleRows
                }, '◀'),
                CollapsibleDataManager.collapsibleDataCollapsedKeys.value.size > 0 && h('button', {
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
    Array.from(coreDetectedKeys.value)
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

  const collapsedDataExtractedColumns: ColDef[] = Array.from(CollapsibleDataManager.collapsibleDataExpandedKeys.value)
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

    const expandableDataExtractedColumns: ColDef[] = Array.from(ExpandableDataManager.expandedExpandableDataKeys.value)
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
      cellRenderer: (params: ColumnManager.RenderParams, _: any) => {
        return params.data[ColumnManager.EXPANDABLE_DATA_COLUMN][key];
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
              h('div', { class: 'ag-header-cell-text' },
                ExpandableDataManager.expandableDataUnexpandedKeys.value.size == 0 ? "" : params.displayName
              ),
            ]),
            h('div', { class: 'button-container' }, [
              h('div', {}, [
                ExpandableDataManager.expandedExpandableDataKeys.value.size > 0 && h('button', {
                  class: 'ag-my collapse all',
                  title: 'Collapse all',
                  onClick: contractAllExpandableRows
                }, '◀'),
                ExpandableDataManager.expandableDataUnexpandedKeys.value.size > 0 && h('button', {
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
  if (ExpandableDataManager.expandedExpandableDataRows.value.has(rowIndex)) {
    ExpandableDataManager.expandedExpandableDataRows.value.delete(rowIndex);
  } else {
    ExpandableDataManager.expandedExpandableDataRows.value.add(rowIndex);
    const row = rowData.value[rowIndex];
    if (row) {
      if (row[ColumnManager.EXPANDABLE_DATA_COLUMN]) {
        Object.keys(row[ColumnManager.EXPANDABLE_DATA_COLUMN]).forEach(key => {
          ExpandableDataManager.expandedExpandableDataKeys.value.add(key);
          ExpandableDataManager.expandableDataUnexpandedKeys.value.delete(key);
        });
      }
    }
  }
  updateColumnDefs();
};

const toggleContractExpandableKeys = (rowIndex: number) => {
  if (!ExpandableDataManager.expandedExpandableDataRows.value.has(rowIndex)) {
    return;
  } else {
    ExpandableDataManager.expandedExpandableDataRows.value.delete(rowIndex);
    const row = rowData.value[rowIndex];
    if (row) {
      if (row[ColumnManager.EXPANDABLE_DATA_COLUMN]) {
        Object.keys(row[ColumnManager.EXPANDABLE_DATA_COLUMN]).forEach(key => {
          ExpandableDataManager.expandedExpandableDataKeys.value.delete(key);
          ExpandableDataManager.expandableDataUnexpandedKeys.value.add(key);
        });
      }
    }
  }
  updateColumnDefs();
}

const toggleExpandCollapsibleKeys = (rowIndex: number) => {
  console.log("!!!toggleExpandExtraDataKeys", rowIndex);
  if (CollapsibleDataManager.collapsibleDataExpandedRows.value.has(rowIndex)) {
    CollapsibleDataManager.collapsibleDataExpandedRows.value.delete(rowIndex);
  } else {
    CollapsibleDataManager.collapsibleDataExpandedRows.value.add(rowIndex);
    const row = rowData.value[rowIndex];
    if (row) {
      console.log("row", row);
      Object.keys(row).filter(
        key => !ColumnManager.COMMON_COLUMN_KEYS.value.has(key) && !ColumnManager.SPECIAL_COLUMN_KEYS.has(key)
      ).forEach(key => {
        console.log("adding key", key);
        CollapsibleDataManager.collapsibleDataExpandedKeys.value.add(key)
        CollapsibleDataManager.collapsibleDataCollapsedKeys.value.delete(key);
      });
    }
  }
  updateColumnDefs();
}

const toggleContractCollapsibleKeys = (rowIndex: number) => {
  if (!CollapsibleDataManager.collapsibleDataExpandedRows.value.has(rowIndex)) {
    return;
  } else {
    CollapsibleDataManager.collapsibleDataExpandedRows.value.delete(rowIndex);
    const row = rowData.value[rowIndex];
    if (row) {
      Object.keys(row).filter(
        key => (!ColumnManager.COMMON_COLUMN_KEYS.value.has(key) && !ColumnManager.SPECIAL_COLUMN_KEYS.has(key))
      ).forEach(key => {
        CollapsibleDataManager.collapsibleDataExpandedKeys.value.delete(key);
        CollapsibleDataManager.collapsibleDataCollapsedKeys.value.add(key);
      });
    }
  }
  updateColumnDefs();
}

const restoreAllCollapsedRows = () => {
  rowData.value.forEach(row => {
    CollapsibleDataManager.collapsibleDataExpandedRows.value.add(row.id);
    Object.keys(row).filter(
      key => !ColumnManager.COMMON_COLUMN_KEYS.value.has(key) && !ColumnManager.SPECIAL_COLUMN_KEYS.has(key)
    ).forEach(key => CollapsibleDataManager.collapsibleDataExpandedKeys.value.add(key));
  });
  CollapsibleDataManager.collapsibleDataCollapsedKeys.value.clear();
  updateColumnDefs();
};

const collapseAllCollapsibleRows = () => {
  CollapsibleDataManager.collapsibleDataExpandedRows.value.clear();
  CollapsibleDataManager.resetKeys();
  updateColumnDefs();
};


const expandAllExpandableRows = () => {
  let isDirty = false;
  rowData.value.forEach(row => {
    if (row.payload) {
      const payloadKeys = Object.keys(row.payload);
      if (payloadKeys.length > 0) {
        ExpandableDataManager.expandedExpandableDataRows.value.add(row.id);
        payloadKeys.forEach(key => ExpandableDataManager.expandedExpandableDataKeys.value.add(key));
        isDirty = true;
      }
    }
  });
  if (isDirty) {
    ExpandableDataManager.expandableDataUnexpandedKeys.value.clear();
    updateColumnDefs();
  }
};

const contractAllExpandableRows = () => {
  ExpandableDataManager.expandedExpandableDataKeys.value.clear();
  ExpandableDataManager.resetKeys();
  updateColumnDefs();
};

const onQuickFilterChanged = () => {
  const filterText = quickFilterText.value.trim();
  console.log("filterText", filterText);
  // apply filter to the COLLAPSABLE_DATA_COLUMN_SHADOW
  gridApi.value?.setColumnFilterModel(ColumnManager.COLLAPSABLE_DATA_COLUMN_SHADOW, {
    filterType: 'text',
    type: 'contains',
    filter: filterText,
  });

  // DEBUG: display the first 5 rows's data in the COLLAPSABLE_DATA_COLUMN_SHADOW
  console.log("first 5 rows's data in the COLLAPSABLE_DATA_COLUMN_SHADOW", rowData.value.slice(0, 5).map(row => row[ColumnManager.COLLAPSABLE_DATA_COLUMN_SHADOW]));
  console.log("First 5 rows full data", rowData.value.slice(0, 5));

  /* if (filterText.length === 0) {
    gridApi.value?.setFilterModel({});
  } else {
    gridApi.value?.setFilterModel({
      ...({[ColumnManager.EXPANDABLE_DATA_COLUMN_SHADOW]: {
        filterType: 'text',
        type: 'contains',
        filter: quickFilterText.value.trim(),
      }} ?? {}),
      ...({[ColumnManager.COLLAPSABLE_DATA_COLUMN_SHADOW]: {
        filterType: 'text',
        type: 'contains',
        filter: quickFilterText.value.trim(),
      }} ?? {}),
    });
    } */
  gridApi.value?.onFilterChanged();
};

onMounted(() => {
});


watch(
  () => props.rowData,
  (newData, _) => {
    rowData.value = newData;
    processRows();
  },
  { deep: true }
);

const dataSource: IDatasource = {
  getRows: async (params) => {
    const { startRow, endRow, successCallback, failCallback } = params;
    
    CollapsibleDataManager.clearAll();
    ExpandableDataManager.clearAll();

    try {
      const limit = endRow - startRow;
      const result = await loadRemoteData(startRow, limit);
      
      // Assuming loadRemoteData returns an object with data and totalCount
      const { rows, totalRowCount } = result;

      // preprocess for ag-grid;
      // something weird with json string escaping
      // when triggered manually no problem;
      // when used as infinite datasource, it has problem
      for (const row of rows) {
        row.payload = JSON.parse(row.payload);
        Object.keys(row.payload).forEach(key => CollapsibleDataManager.collapsibleDataDetectedKeys.value.add(key));
      }
      
      // If this is the last block of data, pass the actual row count
      const lastRow = startRow + rows.length >= totalRowCount ? totalRowCount : undefined;
      
      successCallback(rows, lastRow);
    } catch (error) {
      console.error('Error fetching remote data:', error);
      failCallback();
    }
  }
};
</script>