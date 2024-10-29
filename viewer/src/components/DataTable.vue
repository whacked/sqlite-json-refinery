<template>
  <div class="main-grid-container">
    <div class="my-data-filters">
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
    <div class="ag-grid-container ag-theme-alpine">
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
            <th>expandable keys</th><td>{{ expandableDataManager.detectedKeys.value }}</td>
          </tr>
          <tr>
            <th>unexpanded keys</th><td>{{ expandableDataManager.expandableDataUnexpandedKeys.value }}</td>
          </tr>
          <tr>
            <th>expanded keys</th><td>{{ expandableDataManager.expandedExpandableDataKeys.value }}</td>
          </tr>

          <tr>
            <th>collapsible keys</th><td>{{ collapsibleDataManager.detectedKeys.value }}</td>
          </tr>
          <tr>
            <th>uncollapsed keys</th><td>{{ collapsibleDataManager.collapsibleDataExpandedKeys.value }}</td>
          </tr>
          <tr>
            <th>collapsed keys</th><td>{{ collapsibleDataManager.collapsibleDataCollapsedKeys.value }}</td>
          </tr>
        </tbody>
      </table>

      <div>
        <summary>colorize columns</summary>
        <table>
          <thead>
            <tr>
              <th>column</th>
              <th>set X</th>
              <th>set Y</th>
              <th>categorical</th>
              <th>time</th>
              <th>coerce to number</th>
              <th>extract units</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="column in (
                Array.from(ColumnManager.coreDetectedKeys.value)
                  .concat(Array.from(ColumnManager.expandableDataManager.expandedExpandableDataKeys.value))
                  .concat(Array.from(ColumnManager.collapsibleDataManager.collapsibleDataExpandedKeys.value))
            )"
            >
              <td
                :style="(ColumnManager.expandableDataManager.expandedExpandableDataKeys.value.has(column) || ColumnManager.collapsibleDataManager.collapsibleDataExpandedKeys.value.has(column)) ? Colorizer.makeTextContainerStyle(column) : null"
              >
                <ColorizedNestedColumn :column="column" />
              </td>
              <td>
                <label>
                  <input type="radio" :checked="plotSettings.xColumn == column"
                    @change="plotSettings.xColumn = column"
                  />
                </label>
              </td>
              <td>
                <label>
                  <input type="radio" :checked="plotSettings.yColumn == column"
                    @change="plotSettings.yColumn = column"
                  />
                </label>
              </td>
              <td>
                <label>
                  <input type="checkbox" :checked="ColumnManager.ColumnTypeTracker.categoricalColumns.value.has(column)"
                    @change="toggleColumnTracker(column, ColumnManager.ColumnTypeTracker.categoricalColumns)"
                  />
                </label>
              </td>
              <td>
                <label>
                  <input type="checkbox" :checked="ColumnManager.ColumnTypeTracker.timeColumns.value.has(column)"
                    @change="toggleColumnTracker(column, ColumnManager.ColumnTypeTracker.timeColumns)"
                  />
                </label>
              </td>
              <td>
                <label>
                  <input type="checkbox" :checked="ColumnManager.ColumnTypeTracker.coerceToNumberColumns.value.has(column)"
                    @change="toggleColumnTracker(column, ColumnManager.ColumnTypeTracker.coerceToNumberColumns)"
                  />
                </label>
              </td>
              <td>
                <button @click="extractUnits(column)">extract units</button>
              </td>
            </tr>
            <template
              v-for="columnGroup in ColumnManager.ColumnTypeTracker.derivedColumnGroups.value"
            >
              <tr
                v-for="(column, index) in columnGroup.derivedColumns"
                :style="Colorizer.makeTextContainerStyle(columnGroup.originalColumn)"
              >
                <td
                  :style="Colorizer.makeTextContainerStyle(column)"
                >
                  <ColorizedNestedColumn :column="column" />
                </td>
                <td></td>
                <td></td>
                <td></td>
                <td>
                  <button v-if="index == 0" @click="underiveColumn(columnGroup)">restore</button>
                </td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>

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

      <div>
        <label>
          <input type="checkbox" @change="togglePlot" />
          Show plot
        </label>
        <div ref="plotContainer"></div>
      </div>
      
    </div>
  </div>
</template>

<style scoped>
/* wtf, the ag styles are not getting applied from style.css */
.main-grid-container {
  width: 100%;
  height: 80em;
  background: beige;
}

.ag-grid-container {
  height: 60em;
}

.ag-theme-quartz {
  /* this actually works but overall it's a mess */
  --ag-header-background-color: #fff;
  --ag-header-foreground-color: #333;
  --ag-header-cell-hover-background-color: none;
  --ag-header-cell-moving-background-color: rgb(180, 140, 140);

  --ag-header-column-resize-handle-display: block;
  --ag-header-column-resize-handle-height: 80%;
  --ag-header-column-resize-handle-width: 2px;
  --ag-header-column-resize-handle-color: orange;
}

.main-grid {
  background: orange;
  height: 100%;
}

.add-column {
  display: inline-block;
}

.my-data-filters {
  margin-bottom: 10px;
}

</style>

<script setup lang="ts">
import { ref, onMounted, reactive, defineComponent, h, Ref, watch } from 'vue';
import { AgGridVue } from 'ag-grid-vue3';
import { BodyScrollEvent, ColDef, GridApi, GridReadyEvent, IDatasource, RowNode, ValueFormatterParams, ValueGetterParams } from 'ag-grid-community';
import { useDataStore } from '@/stores/dataStore';
import ExpandableCell from '@/components/ExpandableCell.vue';
import CollapsableCell from '@/components/CollapsableCell.vue';
import PhotoCell from '@/components/PhotoCell.vue';
import * as ColumnManager from '@/utils/columnManager';
import { loadRemoteData } from '@/stores/remoteDataLoader';
import ColorizedNestedColumn from '@/components/ColorizedNestedColumn.vue';
import TimeCell from '@/components/TimeCell.vue';
import ColorizedCategoricalCell from '@/components/ColorizedCategoricalCell.vue';
import { Colorizer } from './styling';
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
            return h('span', {
              class: 'ag-header-cell-text',
              style: Colorizer.makeTextContainerStyle(word),
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
  ColumnManager.availableColumns.value = Array.from(collapsibleDataManager.detectedKeys.value)
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

const toggleColumnTracker = (column: string, tracker: Ref<Set<string>>) => {
  if (tracker.value.has(column)) {
    tracker.value.delete(column);
  } else {
    tracker.value.add(column);
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

function coerceToNumber(value: string) {
  if(value == null || value == undefined || value == "") return null;
  return parseFloat(value);
}

const underiveColumn = (columnGroup: ColumnManager.DerivedColumnGroup) => {
  if(!gridApi.value) return;

  const rowsToUpdate: any[] = [];
  for(let i = 0; i < gridApi.value.getDisplayedRowCount(); ++i) {
    const row = gridApi.value.getDisplayedRowAtIndex(i);
    if (!row || !row.data) continue;

    for(const column of columnGroup.derivedColumns) {
      delete row.data[column];
      rowsToUpdate.push(row.data);
    }
  }

  gridApi.value?.applyTransaction({ update: rowsToUpdate });

  if( columnGroup.isFromExpandedData ) {
    expandableDataManager.unhideKey(columnGroup.originalColumn);
  } else {
    collapsibleDataManager.unhideKey(columnGroup.originalColumn);
  }

  ColumnManager.ColumnTypeTracker.derivedColumnGroups.value = ColumnManager.ColumnTypeTracker.derivedColumnGroups.value.filter(group => group.originalColumn !== columnGroup.originalColumn);
  updateColumnDefs();
}

const extractUnits = async (column: string) => {
  if(!gridApi.value) return;
  function makeUnitColumn(columnPrefix: string, unit: string) {
    return `${columnPrefix}.${unit}`;
  }
  const discoveredUnitColumns = new Set<string>();
  const trackIsFromExpandedData = new Set<boolean>();
  const rowsToUpdate: any[] = [];
  for(let i = 0; i < gridApi.value.getDisplayedRowCount(); ++i) {
    const row = gridApi.value.getDisplayedRowAtIndex(i);
    if (!row || !row.data) continue;

    let rowValue: any;
    // use a primitive method to check if the value is in the original data map
    // or if it's an expandable-data value
    let expandedDataValue: any | null;
    let maybeOriginalDataValue: any | null;
    if (column.startsWith(ColumnManager.EXPANDABLE_DATA_COLUMN + '.')) {
      expandedDataValue = row.data[ColumnManager.EXPANDABLE_DATA_COLUMN]?.[column.substring(ColumnManager.EXPANDABLE_DATA_COLUMN.length + 1)];
    } else {
      maybeOriginalDataValue = row.data[column];
    }

    if(expandedDataValue === undefined && maybeOriginalDataValue === undefined) {
      continue;
    }

    if (maybeOriginalDataValue !== undefined) {
      trackIsFromExpandedData.add(false);
      rowValue = maybeOriginalDataValue;
    } else {
      trackIsFromExpandedData.add(true);
      rowValue = expandedDataValue;
    }

    const parsedValue = ColumnManager.parseValueWithUnitSuffix(rowValue);
    if (parsedValue.unit != null) {
      const unitColumn = makeUnitColumn(column, parsedValue.unit);
      discoveredUnitColumns.add(unitColumn);
      row.data[unitColumn] = parsedValue.value;
      rowsToUpdate.push(row.data); 
    }
  }

  if (trackIsFromExpandedData.size !== 1) {
    throw new Error(`Expected only 1 value for isFromExpandedData, but got ${trackIsFromExpandedData.size}`);
  }

  gridApi.value?.applyTransaction({ update: rowsToUpdate });

  const isFromExpandedData = trackIsFromExpandedData.has(true);
  ColumnManager.ColumnTypeTracker.derivedColumnGroups.value.push({
    originalColumn: column,
    isFromExpandedData,
    derivedColumns: Array.from(discoveredUnitColumns),
  });

  // note this is super tricky because if the column is from the expanded data,
  // it gets placed into the row data, which gets picked up as collapsed data
  if(isFromExpandedData) {
    expandableDataManager.hideKey(column);
  } else {
    collapsibleDataManager.hideKey(column);
  }
  updateColumnDefs();
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
        collapsibleDataManager.detectedKeys.value.add(key);
      }
    });
    let expandableData: object | null = null;
    if (row[ColumnManager.EXPANDABLE_DATA_COLUMN]) {
      totalExpandableRows.value++;
      expandableData = JSON.parse(row[ColumnManager.EXPANDABLE_DATA_COLUMN]);
      if (expandableData) {
        Object.keys(expandableData).forEach(key => {
          const fullKey = ColumnManager.makeExpandableDataColumnKey(key);
          expandableDataManager.detectedKeys.value.add(fullKey);
          expandableDataManager.expandableDataUnexpandedKeys.value.add(fullKey);
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
    collapsibleDataHiddenKeys: collapsibleDataManager.hiddenKeys.value,
    expandableDataHiddenKeys: expandableDataManager.hiddenKeys.value,
    // for whatever reason, passing the manager itself has different behavior
    // or it's actually not the same object?
    expandableDataManager,
    collapsibleDataManager,
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
      headerName: `Collapsed (${collapsibleDataManager.collapsibleDataCollapsedKeys.value.size}/${collapsibleDataManager.detectedKeys.value.size})`,
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
    if (ColumnManager.ColumnTypeTracker.timeColumns.value.has(key)) {
      return {
        field: key,
        headerName: key,
        cellRenderer: 'timeCellRenderer',
      }
    } else if (ColumnManager.ColumnTypeTracker.categoricalColumns.value.has(key)) {
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

  const derivedColumns: ColDef[] = [];
  for(const group of ColumnManager.ColumnTypeTracker.derivedColumnGroups.value) {
    for(const col of group.derivedColumns) {
      // prevent the derived columns from being picked up as collapsed data
      collapsibleDataManager.hideKey(col)
      derivedColumns.push({
        field: col,
        headerName: col,
        headerClass: group.isFromExpandedData ? 'my-ag-table-expandable-data-header' : 'my-ag-table-derived-from-expanded-data-header',
        headerComponent: (
          defineComponent({
            props: ['displayName', 'onCustomAction'],
            setup(props) {
              //@ts-ignore  this is correct, but flagged by the linter
              const params = props.params;
              const sortOrder = ref(params.column.getSort());

              /* this does NOT work */
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
                  [h(ColorizedNestedColumn, {
                    column: params.displayName
                  })]
                ),
                /* sort button does not work */
                /* h('button', {
                  class: 'ag-my-sort-button',
                  onClick: onSortClicked
                }, (
                  sortOrder.value === "asc" ? '▲' : 
                  sortOrder.value === "desc" ? '▼' : '⇅'
                )), */
              ]);
            }
          })
        ),
        valueFormatter: (params: ValueFormatterParams) => {
          return params.data[col];
        },
      });
    }
  }

  const collapsedDataExtractedColumns: ColDef[] = Array.from(collapsibleDataManager.getVisibleKeys())
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

    const expandableDataExtractedColumns: ColDef[] = Array.from(expandableDataManager.getVisibleKeys())
    .sort((a, b) => a.localeCompare(b))
    .map(key => {
      const fullKey = `${ColumnManager.EXPANDABLE_DATA_COLUMN}.${key}`
        return {
        field: fullKey,
        headerName: key,
        headerClass: 'my-ag-table-expandable-data-expanded-column',
        cellClass: 'my-ag-table-expandable-data-expanded-cell',
        headerComponent: expandedPayloadColumnHeader,
        headerComponentParams: {
          key: key,
        },
        valueFormatter: (params: ValueFormatterParams) => {
          const subKey = key.substring(ColumnManager.EXPANDABLE_DATA_COLUMN.length + 1);
          const value = params.data[ColumnManager.EXPANDABLE_DATA_COLUMN][subKey];
          if (ColumnManager.ColumnTypeTracker.coerceToNumberColumns.value.has(fullKey)) {
            return coerceToNumber(value);
          }
          return value;
        },
        cellRenderer: selectRenderer(key).cellRenderer,
      }
    });

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
    ...derivedColumns,
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
        Object.keys(row[ColumnManager.EXPANDABLE_DATA_COLUMN]).forEach(fullKey => {
          expandableDataManager.expandedExpandableDataKeys.value.add(ColumnManager.makeExpandableDataColumnKey(fullKey));
          expandableDataManager.expandableDataUnexpandedKeys.value.delete(ColumnManager.makeExpandableDataColumnKey(fullKey));
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
        Object.keys(row[ColumnManager.EXPANDABLE_DATA_COLUMN]).forEach(fullKey => {
          expandableDataManager.expandedExpandableDataKeys.value.delete(ColumnManager.makeExpandableDataColumnKey(fullKey));
          expandableDataManager.expandableDataUnexpandedKeys.value.add(ColumnManager.makeExpandableDataColumnKey(fullKey));
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
        payloadKeys.forEach(key => expandableDataManager.expandedExpandableDataKeys.value.add(ColumnManager.makeExpandableDataColumnKey(key)));
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


import * as Plotly from 'plotly.js-dist';
const plotContainer = ref(null);

interface PlotSettings {
  xColumn: string | null;
  yColumn: string | null;
}
const plotSettings = ref<PlotSettings>({
  xColumn: null,
  yColumn: null,
});

const togglePlot = (event: Event) => {
  const isChecked = (event.target as HTMLInputElement).checked;
  if(!isChecked) {
    return;
  }

  if(plotSettings.value.xColumn == null || plotSettings.value.yColumn == null) {
    alert("No plot settings");
    return;
  }

  /*
  let timeColumnCandidate: string | null = null;
  for(const col of ColumnManager.ColumnTypeTracker.timeColumns.value) {
    if(rowData.value.every(row => row[col] != null)) {
      timeColumnCandidate = col;
      break;
    }
  }

  let valueColumnCandidate: string | null = null;
  for(const group of ColumnManager.ColumnTypeTracker.derivedColumnGroups.value) {
    for(const col of group.derivedColumns) {
      if(rowData.value.some(row => row[col] != null)) {
        valueColumnCandidate = col;
        break;
      }
    }
  }

  if(timeColumnCandidate == null) {
    console.warn("No time column candidate found");
    return;
  }
  if(valueColumnCandidate == null) {
    console.warn("No value column candidate found");
    return;
  }
  */

  const X: any[] = [];
  const Y: any[] = [];

  for(const row of rowData.value) {
    // SUPER DUPER RISKY AND DIRTY
    X.push(ColumnManager.parseTimeValue(
      row[plotSettings.value.xColumn] ?? row[ColumnManager.EXPANDABLE_DATA_COLUMN]?.[plotSettings.value.xColumn]
    ));
    Y.push(
      row[plotSettings.value.yColumn] ?? row[ColumnManager.EXPANDABLE_DATA_COLUMN]?.[plotSettings.value.yColumn]
    );
  }

  const plot = new Plotly.newPlot(plotContainer.value, [{
    x: X,
    y: Y,
    type: 'scatter',
    mode: 'markers',
    marker: {
      size: 10,
      color: 'rgba(17, 157, 255,0.5)',
      symbol: 'circle',
    },
  }], {
    xaxis: {
      title: plotSettings.value.xColumn,
      type: 'date'
    },
    yaxis: {
      title: plotSettings.value.yColumn,
    }
  });

  console.log(plot);
}

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
        Object.keys(row.payload).forEach(key => collapsibleDataManager.detectedKeys.value.add(key));
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