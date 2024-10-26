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
            <th>expandable keys</th><td>{{ expandableDataManager.expandableDataDetectedKeys.value }}</td>
          </tr>
          <tr>
            <th>unexpanded keys</th><td>{{ expandableDataManager.expandableDataUnexpandedKeys.value }}</td>
          </tr>
          <tr>
            <th>expanded keys</th><td>{{ expandableDataManager.expandedExpandableDataKeys.value }}</td>
          </tr>

          <tr>
            <th>collapsible keys</th><td>{{ collapsibleDataManager.collapsibleDataDetectedKeys.value }}</td>
          </tr>
          <tr>
            <th>uncollapsed keys</th><td>{{ collapsibleDataManager.collapsibleDataExpandedKeys.value }}</td>
          </tr>
          <tr>
            <th>collapsed keys</th><td>{{ collapsibleDataManager.collapsibleDataCollapsedKeys.value }}</td>
          </tr>
        </tbody>
      </table>

      <h4>colorize columns</h4>
      <ul class="column-options">
        <li v-for="column in (
          Array.from(ColumnManager.coreDetectedKeys.value)
            .concat(Array.from(ColumnManager.expandableDataManager.expandedExpandableDataKeys.value))
            .concat(Array.from(ColumnManager.collapsibleDataManager.collapsibleDataExpandedKeys.value))
          )">
          <label>
            <input type="checkbox" :checked="ColumnManager.categoricalColumns.value.has(column)"
              @change="toggleCategoricalColumn(column)"
            />
            {{ column }}
          </label>
        </li>
      </ul>


      <!--
      NOTE: rowModelType: infinite is not supported with rowData;
      to use generated data, x- out:
        x-rowModelType="infinite"
        :x-datasource="dataSource"
      -->
      <ag-grid-vue
        v-if="props.rowData.length == 0"

        class="ag-theme-quartz main-grid"
        :columnDefs="columnDefs"
        :defaultColDef="defaultColDef"
        :components="components"

        x-rowModelType="infinite"

        :x-datasource="rowData.length > 0 ? dataSource : null"
        :rowData="rowData"

        :cacheBlockSize="cacheBlockSize"
        :infiniteInitialRowCount="infiniteInitialRowCount"
        :maxBlocksInCache="maxBlocksInCache"
        :rowBuffer="rowBuffer"
        :rowHeight="ColumnManager.COMMON_COLUMN_KEYS.value.has('photo') || collapsibleDataManager.collapsibleDataExpandedKeys.value.has('photo') ? 200 : null"
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
        :rowHeight="ColumnManager.COMMON_COLUMN_KEYS.value.has('photo') || collapsibleDataManager.collapsibleDataExpandedKeys.value.has('photo') ? 200 : null"

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

.column-options {
  list-style-type: none;
  padding: 0;
}

.column-options li {
  display: inline-block;
  margin-right: 10px;
  border: 1px solid black;
  padding: 5px;
  font-size: 12pt;
}

</style>

<script setup lang="ts">
import { ref, onMounted, reactive, defineComponent, h, Ref, toRaw, watch } from 'vue';
import { AgGridVue } from 'ag-grid-vue3';
import { BodyScrollEvent, ColDef, GridApi, GridReadyEvent, IDatasource, ValueGetterParams } from 'ag-grid-community';
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
import { collapsibleDataManager, ContractableColumnsManager, expandableDataManager } from '@/utils/columnManager';


const props = defineProps<{
  rowData: any[];
}>();


function makeExpandedColumnHeader(columnsManager: ContractableColumnsManager) {
  return defineComponent({
    props: ['displayName', 'onCustomAction'],
    setup(props) {
      //@ts-ignore  this is correct, but flagged by the linter
      const params = props.params;
      const sortOrder = ref(params.column.getSort());

      const collapseKey = () => {
        const keyToRemove = params.key;
        columnsManager.contractKey(keyToRemove);
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
            const backgroundColor = colorHash.hex(word);
            const isDark = chroma(backgroundColor).luminance() < 0.6;
            const textColor = isDark ? 'white' : 'black';

            return h('span', {
              class: 'ag-header-cell-text',
              style: { background: backgroundColor, color: textColor },
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

const expandedPayloadColumnHeader = makeExpandedColumnHeader(expandableDataManager);
const expandedCollapsibleDataColumnHeader = makeExpandedColumnHeader(collapsibleDataManager);

const components = {
  expandedPayloadColumnHeader,
  expandedExtraDataColumnHeader: expandedCollapsibleDataColumnHeader,
  expandableCellRenderer: ExpandableCell,
  extractedDataCellRenderer: CollapsableCell,
  photoCellRenderer: PhotoCell,
  timeCellRenderer: TimeCell,
  colorizedCategoricalCellRenderer: ColorizedCategoricalCell,
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
  console.log(">>> detectedKeys", toRaw(collapsibleDataManager.collapsibleDataDetectedKeys.value));
  ColumnManager.availableColumns.value = Array.from(collapsibleDataManager.collapsibleDataDetectedKeys.value)
  .filter(key => (
    key !== ColumnManager.EXPANDABLE_DATA_COLUMN
  ))
  // .filter(key => ColumnManager.COMMON_COLUMN_KEYS.value.has(key))
  .map(key => ({ key, isEnabled: ColumnManager.COMMON_COLUMN_KEYS.value.has(key) }));
  ColumnManager.availableColumns.value = [{
    key: 'id',
    isEnabled: true,
  }];
  updateRowCount();
};

const onFirstDataRendered = () => {
  updateRowCount();
};

const updateRowCount = () => {
  if (!gridApi.value) return;
  TableRowPositionStatus.totalRows.value = dataStore.totalRows;
  TableRowPositionStatus.visibleRows.value = gridApi.value.getDisplayedRowCount();
  TableRowPositionStatus.firstVisibleRow.value = gridApi.value.getFirstDisplayedRowIndex() + 1;
  TableRowPositionStatus.lastVisibleRow.value = gridApi.value.getLastDisplayedRowIndex() + 1;
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


const toggleCategoricalColumn = (column: string) => {
  if (ColumnManager.categoricalColumns.value.has(column)) {
    ColumnManager.categoricalColumns.value.delete(column);
  } else {
    ColumnManager.categoricalColumns.value.add(column);
  }
  updateColumnDefs();
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
const processRows = async () => {
  const sourceData = props.rowData?.length > 0 ? props.rowData : (await dataStore.fetchData(0, dataStore.totalRows));
  ColumnManager.coreDetectedKeys.value.clear();
  collapsibleDataManager.clearAll();
  expandableDataManager.clearAll();

  rowData.value = sourceData.map(row => {
    Object.keys(row).forEach(key => {
      if (ColumnManager.COMMON_COLUMN_KEYS.value.has(key)) {
        ColumnManager.coreDetectedKeys.value.add(key);
      } else if(key != ColumnManager.EXPANDABLE_DATA_COLUMN) {
        collapsibleDataManager.collapsibleDataDetectedKeys.value.add(key);
      }
    });
    let expandableData: object | null = null;
    if (row[ColumnManager.EXPANDABLE_DATA_COLUMN]) {
      totalExpandableRows.value++;
      expandableData = JSON.parse(row[ColumnManager.EXPANDABLE_DATA_COLUMN]);
      if (expandableData) {
        Object.keys(expandableData).forEach(key => {
          expandableDataManager.expandableDataDetectedKeys.value.add(key);
          expandableDataManager.expandableDataUnexpandedKeys.value.add(key);
        });
      }
    }
    collapsibleDataManager.resetKeys()
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
    expandableDataExtractedKeys: expandableDataManager.expandedExpandableDataKeys.value,
    collapsedDataKeys: collapsibleDataManager.collapsibleDataExpandedKeys.value,
    toggleExpandCollapsibleKeys,
    toggleContractCollapsibleKeys,
    toggleExpandExpandableKeys,
    toggleContractExpandableKeys,
  };

  const coreDisplayParamSettings: Record<string, ColDef> = {
    id: { field: 'id', headerName: 'ID', width: 100 },
    country: { field: 'country', headerName: 'Country', width: 150 },
    photo: {
      field: 'photo',
      headerName: 'Photo',
      cellRenderer: 'photoCellRenderer',
    },
    [ColumnManager.COLLAPSABLE_DATA_COLUMN_SHADOW]: {
      field: ColumnManager.COLLAPSABLE_DATA_COLUMN_SHADOW,
      headerName: 'collapsed data String (not shown; for filtering)',
      hide: true,
    },
    [ColumnManager.COLLAPSABLE_DATA_COLUMN]: { 
      field: ColumnManager.COLLAPSABLE_DATA_COLUMN, 
      headerName: `Collapsed (${collapsibleDataManager.collapsibleDataCollapsedKeys.value.size}/${collapsibleDataManager.collapsibleDataDetectedKeys.value.size})`,
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
                collapsibleDataManager.collapsibleDataCollapsedKeys.value.size == 0 ? "" : params.displayName
              ),
            ]),
            h('div', { class: 'button-container' }, [
              h('div', {}, [
                collapsibleDataManager.collapsibleDataExpandedKeys.value.size > 0 && h('button', {
                  class: 'ag-my collapse all',
                  title: 'Collapse all',
                  onClick: collapseAllCollapsibleRows
                }, '◀'),
                collapsibleDataManager.collapsibleDataCollapsedKeys.value.size > 0 && h('button', {
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

  function selectRenderer(key: string) {
    if (ColumnManager.timeColumns.value.has(key)) {
      return {
        field: key,
        headerName: key,
        cellRenderer: 'timeCellRenderer',
      }
    } else if (ColumnManager.categoricalColumns.value.has(key)) {
      return {
        field: key,
        headerName: key,
        cellRenderer: 'colorizedCategoricalCellRenderer',
      }
    } else {
      return coreDisplayParamSettings[key] ?? {
        field: key,
        headerName: key,
        cellRenderer: null,
      }
    }
  }

  const baseColumns: ColDef[] = (
    Array.from(ColumnManager.coreDetectedKeys.value)
    .concat([
      ColumnManager.COLLAPSABLE_DATA_COLUMN,
      // needed for filtering
      ColumnManager.EXPANDABLE_DATA_COLUMN_SHADOW,
      ColumnManager.COLLAPSABLE_DATA_COLUMN_SHADOW,
    ])
    .map(key => selectRenderer(key))
    .filter(colDef => colDef)
  );

  const collapsedDataExtractedColumns: ColDef[] = Array.from(collapsibleDataManager.collapsibleDataExpandedKeys.value)
    .sort((a, b) => a.localeCompare(b))
    .map(key => ({
      field: key,
      headerName: `${key}`,
      headerClass: 'my-ag-table-collapsible-data-expanded-column',
      cellClass: 'my-ag-table-collapsible-data-expanded-cell',
      cellRenderer: selectRenderer(key).cellRenderer,
      headerComponent: expandedCollapsibleDataColumnHeader,
      headerComponentParams: {
        key: key,
      },
    }));

    const expandableDataExtractedColumns: ColDef[] = Array.from(expandableDataManager.expandedExpandableDataKeys.value)
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
      cellRenderer: selectRenderer(key).cellRenderer,
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
      cellDataType: 'text',
      headerName: 'Expandable Data', 
      cellRenderer: 'expandableCellRenderer',
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
                expandableDataManager.expandableDataUnexpandedKeys.value.size == 0 ? "" : params.displayName
              ),
            ]),
            h('div', { class: 'button-container' }, [
              h('div', {}, [
                expandableDataManager.expandedExpandableDataKeys.value.size > 0 && h('button', {
                  class: 'ag-my collapse all',
                  title: 'Collapse all',
                  onClick: contractAllExpandableRows
                }, '◀'),
                expandableDataManager.expandableDataUnexpandedKeys.value.size > 0 && h('button', {
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
  if (expandableDataManager.expandedExpandableDataRows.value.has(rowIndex)) {
    expandableDataManager.expandedExpandableDataRows.value.delete(rowIndex);
  } else {
    expandableDataManager.expandedExpandableDataRows.value.add(rowIndex);
    const row = rowData.value[rowIndex];
    if (row) {
      if (row[ColumnManager.EXPANDABLE_DATA_COLUMN]) {
        Object.keys(row[ColumnManager.EXPANDABLE_DATA_COLUMN]).forEach(key => {
          expandableDataManager.expandedExpandableDataKeys.value.add(key);
          expandableDataManager.expandableDataUnexpandedKeys.value.delete(key);
        });
      }
    }
  }
  updateColumnDefs();
};

const toggleContractExpandableKeys = (rowIndex: number) => {
  if (!expandableDataManager.expandedExpandableDataRows.value.has(rowIndex)) {
    return;
  } else {
    expandableDataManager.expandedExpandableDataRows.value.delete(rowIndex);
    const row = rowData.value[rowIndex];
    if (row) {
      if (row[ColumnManager.EXPANDABLE_DATA_COLUMN]) {
        Object.keys(row[ColumnManager.EXPANDABLE_DATA_COLUMN]).forEach(key => {
          expandableDataManager.expandedExpandableDataKeys.value.delete(key);
          expandableDataManager.expandableDataUnexpandedKeys.value.add(key);
        });
      }
    }
  }
  updateColumnDefs();
}

const toggleExpandCollapsibleKeys = (rowIndex: number) => {
  console.log("!!!toggleExpandExtraDataKeys", rowIndex);
  if (collapsibleDataManager.collapsibleDataExpandedRows.value.has(rowIndex)) {
    collapsibleDataManager.collapsibleDataExpandedRows.value.delete(rowIndex);
  } else {
    collapsibleDataManager.collapsibleDataExpandedRows.value.add(rowIndex);
    const row = rowData.value[rowIndex];
    if (row) {
      console.log("row", row);
      Object.keys(row).filter(
        key => !ColumnManager.COMMON_COLUMN_KEYS.value.has(key) && !ColumnManager.SPECIAL_COLUMN_KEYS.has(key)
      ).forEach(key => {
        console.log("adding key", key);
        collapsibleDataManager.collapsibleDataExpandedKeys.value.add(key)
        collapsibleDataManager.collapsibleDataCollapsedKeys.value.delete(key);
      });
    }
  }
  updateColumnDefs();
}

const toggleContractCollapsibleKeys = (rowIndex: number) => {
  if (!collapsibleDataManager.collapsibleDataExpandedRows.value.has(rowIndex)) {
    return;
  } else {
    collapsibleDataManager.collapsibleDataExpandedRows.value.delete(rowIndex);
    const row = rowData.value[rowIndex];
    if (row) {
      Object.keys(row).filter(
        key => (!ColumnManager.COMMON_COLUMN_KEYS.value.has(key) && !ColumnManager.SPECIAL_COLUMN_KEYS.has(key))
      ).forEach(key => {
        collapsibleDataManager.collapsibleDataExpandedKeys.value.delete(key);
        collapsibleDataManager.collapsibleDataCollapsedKeys.value.add(key);
      });
    }
  }
  updateColumnDefs();
}

const restoreAllCollapsedRows = () => {
  rowData.value.forEach(row => {
    collapsibleDataManager.collapsibleDataExpandedRows.value.add(row.id);
    Object.keys(row).filter(
      key => !ColumnManager.COMMON_COLUMN_KEYS.value.has(key) && !ColumnManager.SPECIAL_COLUMN_KEYS.has(key)
    ).forEach(key => collapsibleDataManager.collapsibleDataExpandedKeys.value.add(key));
  });
  collapsibleDataManager.collapsibleDataCollapsedKeys.value.clear();
  updateColumnDefs();
};

const collapseAllCollapsibleRows = () => {
  collapsibleDataManager.collapsibleDataExpandedRows.value.clear();
  collapsibleDataManager.resetKeys();
  updateColumnDefs();
};


const expandAllExpandableRows = () => {
  let isDirty = false;
  rowData.value.forEach(row => {
    if (row.payload) {
      const payloadKeys = Object.keys(row.payload);
      if (payloadKeys.length > 0) {
        expandableDataManager.expandedExpandableDataRows.value.add(row.id);
        payloadKeys.forEach(key => expandableDataManager.expandedExpandableDataKeys.value.add(key));
        isDirty = true;
      }
    }
  });
  if (isDirty) {
    expandableDataManager.expandableDataUnexpandedKeys.value.clear();
    updateColumnDefs();
  }
};

const contractAllExpandableRows = () => {
  expandableDataManager.expandedExpandableDataKeys.value.clear();
  expandableDataManager.resetKeys();
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
    
    collapsibleDataManager.clearAll();
    expandableDataManager.clearAll();

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
        Object.keys(row.payload).forEach(key => collapsibleDataManager.collapsibleDataDetectedKeys.value.add(key));
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