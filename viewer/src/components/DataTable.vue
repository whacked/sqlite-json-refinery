<template>
  <div class="main-grid-container">
    <div class="my-data-filters">
      <input v-model="quickFilterText" placeholder="Quick filter..." @input="onQuickFilterChanged" />
      <button @click="autoSizeColumns">Fit Columns</button>
    </div>
    <div class="ag-grid-container ag-theme-alpine">
      <div>
        {{ gridApi ? `${gridApi.getDisplayedRowCount()} / ${dataStore.totalRows} rows visible
        (${currentDisplayedRowsRange})` : '0 / 0 rows visible'
        }}; {{ totalExpandableRows }} expandable rows
      </div>


      <table class="data-keys-table">
        <tbody>
          <tr>
            <th>rows</th><td>{{ props.rowData.length }}</td>
          </tr>

          <tr>
            <th>column type tracker</th>
            <td>
              <ul>
                <li>{{ ColumnManager.UsableColumns.getTotalSpeciallyRenderedColumns(ColumnRendererType.CATEGORICAL) }} categorical</li>
                <li>{{ ColumnManager.UsableColumns.getTotalTransformedColumns(ColumnTransformation.NUMBER) }} coerce to number</li>
                <li>{{ ColumnManager.UsableColumns.getTotalSpeciallyRenderedColumns(ColumnRendererType.TIME) }} time</li>

                <li>{{ ColumnManager.UsableColumns.columnsSet.value.size }} total usable</li>
                <li>{{ ColumnManager.UsableColumns.getTotalVisibleSingleLevelColumns() }} single level</li>
                <li>{{ ColumnManager.UsableColumns.getTotalVisibleNestedDepthColumns() }} nested depth</li>
                <li>{{ ColumnManager.UsableColumns.getDerivedColumns().length }} derived</li>
              </ul>
              
            </td>
          </tr>
        </tbody>
      </table>

      <table>
      <tbody style="font-size: 20pt;">

            <tr>
              <th>usable columns ({{ ColumnManager.UsableColumns.columnsSet.value.size }})</th>
              <td>
                <ul>
                  <li
                    v-for="col in Array.from(ColumnManager.UsableColumns.columnsSet.value)"
                    :style="{ display: 'inline-block',
                      marginRight: '10px',
                      border: col.effectiveLookupPath.length == 1 ? '2px solid black' : '2px solid red',
                      background: col.shouldDisplay ? 'lightgreen' : 'lightgray',
                    }"
                    @click="toggleColumnShouldDisplay(col)"
                  >
                    {{ col.displayString }}
                  </li>
                </ul>
              </td>
            </tr>



            <tr class="collapsible-data-keys">
              <th>collapsible keys</th>
              <td>
                {{ ColumnManager.UsableColumns.getAllSingleLevelColumns().map(col => col.displayString).join(', ') }}
              </td>
            </tr>
            <tr class="collapsible-data-keys">
              <th>collapsed keys</th>
              <td>
                {{ ColumnManager.UsableColumns.getAllSingleLevelColumns().filter(col => !col.shouldDisplay).map(col => col.displayString).join(', ') }}
              </td>
            </tr>
            <tr class="collapsible-data-keys">
              <th>uncollapsed keys</th>
              <td>
                {{ ColumnManager.UsableColumns.getAllSingleLevelColumns().filter(col => col.shouldDisplay).map(col => col.displayString).join(', ') }}
              </td>
            </tr>


            <tr class="expandable-data-keys">
              <th>expandable keys</th>
              <td>
                {{ ColumnManager.UsableColumns.getAllNestedDepthColumns().map(col => col.displayString).join(', ') }}
              </td>
            </tr>
            <tr class="expandable-data-keys">
              <th>unexpanded keys keys</th>
              <td>
                {{ ColumnManager.UsableColumns.getAllNestedDepthColumns().filter(col => !col.shouldDisplay).map(col => col.displayString).join(', ') }}
              </td>
            </tr>
            <tr class="expandable-data-keys">
              <th>expanded keys</th>
              <td>
                {{ ColumnManager.UsableColumns.getAllNestedDepthColumns().filter(col => col.shouldDisplay).map(col => col.displayString).join(', ') }}
              </td>
            </tr>

        </tbody>
      </table>

      <details>
        <summary>
          column settings
        </summary>
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
                Array.from(ColumnManager.UsableColumns.columnsSet.value)
                .filter(col => col.derivedFromColumn == null)
            )"
            >
              <td
                :style="(
                  // ColumnManager.expandableDataManager.expandedExpandableDataKeys.value.has(column)
                  ColumnManager.UsableColumns.getAllNestedDepthColumns().filter(
                    col => (col.shouldDisplay && JSON.stringify(col.effectiveLookupPath) == JSON.stringify(column.effectiveLookupPath))
                  ).length > 0
                || ColumnManager.UsableColumns.getAllSingleLevelColumns().filter(col => 
                    (col.shouldDisplay && col.effectiveLookupPath[0] == column.apparentLookupPath)
                  ).length > 0
                ) ? Colorizer.makeTextContainerStyle(column.apparentLookupPath) : null"
              >
                <ColorizedNestedColumn :columnText="column.displayString" />
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
                <div class="emoji-toggle" @click="toggleColumnRendererType(column, ColumnRendererType.CATEGORICAL)">
                  <span :class="column.renderType == ColumnRendererType.CATEGORICAL ? 'checked' : 'unchecked'">
                    🛍️
                  </span>
                </div>
              </td>
              <td>
                <div class="emoji-toggle" @click="toggleColumnRendererType(column, ColumnRendererType.TIME)">
                  <span :class="column.renderType == ColumnRendererType.TIME ? 'checked' : 'unchecked'">
                    🕒
                  </span>
                </div>
              </td>
              <td>
                <div class="emoji-toggle" @click="toggleColumnTransformation(column, ColumnTransformation.NUMBER)">
                  <span :class="column.transformations?.includes(ColumnTransformation.NUMBER) ? 'checked' : 'unchecked'">
                    🔢
                  </span>
                </div>
              </td>
              <td>
                <button @click="extractUnits(column)">extract units</button>
              </td>
            </tr>

            <template
              v-for="(derivedColumn, index) in Array.from(ColumnManager.UsableColumns.columnsSet.value)
              .filter(col => col.derivedFromColumn != null)
              .sort((a, b) => a.displayString.localeCompare(b.displayString))
              "
            >
              <tr
                :style="Colorizer.makeTextContainerStyle(derivedColumn.derivedFromColumn!.effectiveLookupPath.join('.'))"
              >
                <td
                  :style="Colorizer.makeTextContainerStyle(derivedColumn.derivedFromColumn!.effectiveLookupPath.join('.'))"
                >
                  <ColorizedNestedColumn :columnText="derivedColumn.displayString" />
                </td>
                <td>
                  {{ derivedColumn.displayString }}
                </td>
                <td></td>
                <td></td>
                <td>
                  <button @click="underiveColumn(derivedColumn)">restore</button>
                </td>
              </tr>
            </template>

          </tbody>
        </table>
      </details>

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

        @keydown="onKeyDown"

        :cacheBlockSize="cacheBlockSize"
        :infiniteInitialRowCount="infiniteInitialRowCount"
        :maxBlocksInCache="maxBlocksInCache"
        :rowBuffer="rowBuffer"
        :rowHeight="ColumnManager.SERIALIZED_CUSTOMARY_COLUMN_KEYS.value.has('photo') || Array.from(ColumnManager.UsableColumns.columnsSet.value).filter(col => col.shouldDisplay && col.effectiveLookupPath.length == 1 && col.effectiveLookupPath[0] == 'photo').length > 0 ? 200 : null"
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
        :rowHeight="ColumnManager.SERIALIZED_CUSTOMARY_COLUMN_KEYS.value.has('photo') || Array.from(ColumnManager.UsableColumns.columnsSet.value).filter(col => col.shouldDisplay && col.effectiveLookupPath.length == 1 && col.effectiveLookupPath[0] == 'photo').length > 0 ? 200 : null"

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

.data-keys-table {
  font-size: 18pt;
}

.expandable-data-keys {
  background: #d9eefb;
}

.collapsible-data-keys {
  background: #fce8dc;
}

</style>

<script setup lang="ts">
import { ref, onMounted, reactive, defineComponent, h, Ref, watch } from 'vue';
import { AgGridVue } from 'ag-grid-vue3';
import { BodyScrollEvent, ColDef, GridApi, GridReadyEvent, IDatasource, ValueFormatterParams, ValueGetterParams } from 'ag-grid-community';
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
import { ColumnTransformation, ColumnRendererType } from '@/utils/columnManager';


const props = defineProps<{
  rowData: any[];
}>();


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
        console.log(displayedRow);
      }
    }
  }
}


const currentDisplayedRowsRange = ref('');

const onModelUpdated = () => {
  console.log("%cupdated", "color: red; font-weight: bold; font-size: 2em;");
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

const addCoreColumn = () => {
  if (newColumnName.value && !ColumnManager.SERIALIZED_CUSTOMARY_COLUMN_KEYS.value.has(newColumnName.value)) {
    ColumnManager.SERIALIZED_CUSTOMARY_COLUMN_KEYS.value.add(newColumnName.value);
    newColumnName.value = ''; // Clear the input after adding
    updateColumnDefs();
  }
};

const removeColumnName = (column: string) => {
  if (ColumnManager.SERIALIZED_CUSTOMARY_COLUMN_KEYS.value.has(column)) {
    ColumnManager.SERIALIZED_CUSTOMARY_COLUMN_KEYS.value.delete(column);
    updateColumnDefs();
  }
};

const toggleColumnTracker = (column: string, tracker: Ref<Set<string>>) => {
  if (tracker.value.has(column)) {
    tracker.value.delete(column);
  } else {
    console.log("adding", column);
    tracker.value.add(column);
  }
  updateColumnDefs();
};

const toggleColumnShouldDisplay = (column: ColumnManager.ColumnKey) => {
  column.shouldDisplay = !column.shouldDisplay;
  updateColumnDefs();
}

const toggleColumnTransformation = (column: ColumnManager.ColumnKey, transformation: ColumnTransformation) => {
  if (column.transformations?.includes(transformation)) {
    column.transformations = column.transformations.filter(t => t !== transformation);
  } else {
    column.transformations = [...(column.transformations ?? []), transformation];
  }
  updateColumnDefs();
}

const toggleColumnRendererType = (column: ColumnManager.ColumnKey, rendererType: ColumnRendererType) => {
  if(column.renderType == rendererType) { 
    column.renderType = null;
  } else {
    column.renderType = rendererType;
  }
  updateColumnDefs();
}

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

const underiveColumn = (derivedColumn: ColumnManager.ColumnKey) => {
  if(!gridApi.value) return;

  const sourceColumn = derivedColumn.derivedFromColumn;
  const columnsToRemove = Array.from(ColumnManager.UsableColumns.columnsSet.value)
  .filter(col => {
    if(col.derivedFromColumn == sourceColumn) {
      ColumnManager.UsableColumns.removeColumn(col);
      return true;
    }
    return false;
  });

  const rowsToUpdate: any[] = [];
  for(let i = 0; i < gridApi.value.getDisplayedRowCount(); ++i) {
    const row = gridApi.value.getDisplayedRowAtIndex(i);
    if (!row || !row.data) continue;

    for(const column of columnsToRemove) {
      delete row.data[column.apparentLookupPath];
      rowsToUpdate.push(row.data);
    }
  }

  gridApi.value?.applyTransaction({ update: rowsToUpdate });

  updateColumnDefs();
}

const extractUnits = async (column: ColumnManager.ColumnKey) => {
  if(!gridApi.value) return;
  function makeUnitColumn(columnPrefix: string, unit: string) {
    return `${columnPrefix}.${unit == "." ? "" : unit}`;
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
    if(column.effectiveLookupPath[0] == ColumnManager.EXPANDABLE_DATA_COLUMN) {
      expandedDataValue = _getIn(row.data, column.effectiveLookupPath);
    } else {
      maybeOriginalDataValue = row.data[column.apparentLookupPath];
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
    let discoveredUnitColumn: string | undefined = undefined;
    if (parsedValue.unit != null) {
      discoveredUnitColumn = makeUnitColumn(column.apparentLookupPath, parsedValue.unit);
      discoveredUnitColumns.add(discoveredUnitColumn);
      rowsToUpdate.push(row.data);

      const derivedColumnLookupPath: string[] = [
        ...(column.effectiveLookupPath.slice(0, -1)),
        discoveredUnitColumn,
      ]

      const apparentLookupPath = derivedColumnLookupPath.join('.');
      row.data[apparentLookupPath] = parsedValue.value;
      ColumnManager.UsableColumns.addColumn({
        effectiveLookupPath: derivedColumnLookupPath,
        apparentLookupPath: derivedColumnLookupPath.join('.'),
        displayString: apparentLookupPath,
        shouldDisplay: true,
        serializedEffectiveLookupPath: JSON.stringify(derivedColumnLookupPath),
        derivedFromColumn: column,
      });
    }
  }

  if (trackIsFromExpandedData.size !== 1) {
    throw new Error(`Expected only 1 value for isFromExpandedData, but got ${trackIsFromExpandedData.size}`);
  }

  gridApi.value?.applyTransaction({ update: rowsToUpdate });
  updateColumnDefs();
}

const totalExpandableRows = ref<number>(0)

const processRows = async () => {  // process the incoming data, derive columns etc
  const sourceData = props.rowData?.length > 0 ? props.rowData : (await dataStore.fetchData(0, dataStore.totalRows));

  ColumnManager.UsableColumns.reset();

  rowData.value = sourceData.map(row => {
    let expandableData: object | null = null;

    Object.keys(row).forEach(key => {


      if (key != ColumnManager.EXPANDABLE_DATA_COLUMN) {

        let renderType: ColumnRendererType | null = null;

        if (ColumnManager.DEFAULT_TIME_COLUMN_KEYS.has(key)) {
          renderType = ColumnRendererType.TIME;
        } else if (ColumnManager.DEFAULT_CATEGORICAL_COLUMN_KEYS.has(key)) {
          renderType = ColumnRendererType.CATEGORICAL;
        }

        ColumnManager.UsableColumns.addColumn({
          effectiveLookupPath: [key],
          apparentLookupPath: key,
          displayString: key,
          shouldDisplay: ColumnManager.CUSTOMARY_COLUMN_KEYS.has(key),
          serializedEffectiveLookupPath: JSON.stringify([key]),
          renderType,
        });

      } else {

        const maybeExpandableData = JSON.parse(row[ColumnManager.EXPANDABLE_DATA_COLUMN]);
        if (maybeExpandableData) {
          totalExpandableRows.value++;
          expandableData = maybeExpandableData;
          Object.keys(maybeExpandableData).forEach(subKey => {
            const effectiveLookupPath = [ColumnManager.EXPANDABLE_DATA_COLUMN, subKey];
            ColumnManager.UsableColumns.addColumn({
              effectiveLookupPath,
              apparentLookupPath: subKey,
              displayString: ColumnManager.makeExpandableDataColumnKey(subKey),
              shouldDisplay: false,
              serializedEffectiveLookupPath: JSON.stringify(effectiveLookupPath),
            });
          });
        }
      }

      if (ColumnManager.SERIALIZED_CUSTOMARY_COLUMN_KEYS.value.has(key)) {
        ColumnManager.UsableColumns.addColumn({
          effectiveLookupPath: [key],
          apparentLookupPath: key,
          displayString: key,
          shouldDisplay: true,
          serializedEffectiveLookupPath: JSON.stringify([key]),
        });
        // ColumnManager.DEPRECATE_coreDetectedKeys.value.add({
        //   effectiveLookupPath: [key],
        //   apparentLookupPath: key,
        //   displayString: key,
        //   shouldDisplay: true,
        //   serializedEffectiveLookupPath: JSON.stringify([key]),
        // });
      } else if(key != ColumnManager.EXPANDABLE_DATA_COLUMN) {
        // collapsibleDataManager.detectedKeys.value.add(key);
      }
    });

    return {
      ...row,
      [ColumnManager.COLLAPSABLE_DATA_COLUMN_SHADOW]: JSON.stringify(ColumnManager.objectWithoutKeys(row, ColumnManager.SPECIAL_COLUMN_KEYS)),
      [ColumnManager.EXPANDABLE_DATA_COLUMN_SHADOW]: row[ColumnManager.EXPANDABLE_DATA_COLUMN],
      [ColumnManager.EXPANDABLE_DATA_COLUMN]: expandableData,
    }
  });
  updateColumnDefs();
  updateRowCount();
  console.log("processed rows");
};

function makeExpandedColumnHeader() {
  return defineComponent({
    props: ['displayName', 'onCustomAction'],
    setup(props) {
      //@ts-ignore  this is correct, but flagged by the linter
      const params = props.params;
      const sortOrder = ref(params.column.getSort());

      const collapseKey = () => {
        const columnKeyLookupPath = typeof params.key == 'string'
        ? [params.key]
        : params.key.effectiveLookupPath;
        console.log(`%ccollapsekey: %c${columnKeyLookupPath}`, 'color: gold; font-weight: bold;', 'color: blue; font-weight: bold; font-size: 18pt;');
        ColumnManager.UsableColumns.setDisplayOff(columnKeyLookupPath);
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

const expandableCollapsibleDataColumnHeader = makeExpandedColumnHeader();

const components = {
  expandedPayloadColumnHeader: expandableCollapsibleDataColumnHeader,
  expandedExtraDataColumnHeader: expandableCollapsibleDataColumnHeader,
  expandableCellRenderer: ExpandableCell,
  extractedDataCellRenderer: CollapsableCell,
  photoCellRenderer: PhotoCell,
  timeCellRenderer: TimeCell,
  colorizedCategoricalCellRenderer: ColorizedCategoricalCell,
};

const updateColumnDefs = () => {
  const cellRendererParams: ColumnManager.RenderParams = {
    coreDisplayParams: ColumnManager.SERIALIZED_CUSTOMARY_COLUMN_KEYS.value,
    
    toggleExpandCollapsibleKeys,
    toggleExpandExpandableKeys,
    toggleContractExpandableKeys,

    currentExpandedKeys: new Set<string>(ColumnManager.UsableColumns.getAllNestedDepthColumns().filter(col => col.shouldDisplay).map(col => col.effectiveLookupPath[1])),
    currentRestoredKeys: new Set<string>(ColumnManager.UsableColumns.getAllSingleLevelColumns().filter(col => col.shouldDisplay).map(col => col.effectiveLookupPath[0])),
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
    [ColumnManager.EXPANDABLE_DATA_COLUMN_SHADOW]: { 
      field: ColumnManager.EXPANDABLE_DATA_COLUMN_SHADOW,
      headerName: 'Expandable Data String (not shown; for filtering)', 
      hide: true 
    },
    [ColumnManager.COLLAPSABLE_DATA_COLUMN]: {
      // hide: true,
      field: ColumnManager.COLLAPSABLE_DATA_COLUMN, 
      headerName: ColumnManager.UsableColumns.getTotalCollapsedCountString(),
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
                ColumnManager.UsableColumns.getTotalHiddenSingleLevelColumns() == 0 ? "" : params.displayName
              ),
            ]),
            h('div', { class: 'button-container' }, [
              h('div', {}, [
                ColumnManager.UsableColumns.getTotalVisibleSingleLevelColumns() > 0 && h('button', {
                  class: 'ag-my-collapse-all',
                  title: 'Collapse all',
                  onClick: collapseAllCollapsibleRows
                }, '◀'),
                ColumnManager.UsableColumns.getTotalHiddenSingleLevelColumns() > 0 && h('button', {
                  class: 'ag-my-expand-all',
                  title: 'Expand all',
                  onClick: restoreAllCollapsedRows
                }, '▶'),
              ]),
            ]),
          ]);
        }
      }),
    },
  };

  function selectRenderer(
    columnKey: ColumnManager.ColumnKey,
  ) {
    const serializedPath = columnKey.serializedEffectiveLookupPath;
    const nativePath = columnKey.apparentLookupPath;
    const humanReadablePath = columnKey.displayString;
    
    // console.log("selectRenderer", serializedPath);
    if (columnKey.renderType == ColumnRendererType.TIME) {
      return {
        field: nativePath,
        headerName: humanReadablePath ?? serializedPath,
        cellRenderer: 'timeCellRenderer',
        valueFormatter: (params: ValueFormatterParams) => {
          if(columnKey.effectiveLookupPath[0] == ColumnManager.EXPANDABLE_DATA_COLUMN) {
            return params.value[columnKey.apparentLookupPath];
          } else {
            return params.value;
          }
        },
      }
    } else if (columnKey.renderType == ColumnRendererType.CATEGORICAL) {
      return {
        field: nativePath,
        headerName: humanReadablePath ?? serializedPath,
        cellRenderer: 'colorizedCategoricalCellRenderer',
      }
    } else {
      return coreDisplayParamSettings[nativePath] ?? {
        field: nativePath,
        headerName: humanReadablePath ?? serializedPath,
        cellRenderer: null,
      }
    }
  }

  const baseColumns: ColDef[] = (
    Array.from(ColumnManager.UsableColumns.getAllSingleLevelColumns())
    .filter(col => {
      return col.shouldDisplay && ColumnManager.SERIALIZED_CUSTOMARY_COLUMN_KEYS.value.has(col.serializedEffectiveLookupPath)
    })
    .map(col => {
      return selectRenderer(col);
    })
    .concat(
      [
        ColumnManager.COLLAPSABLE_DATA_COLUMN,
        // needed for filtering
        ColumnManager.EXPANDABLE_DATA_COLUMN_SHADOW,
        ColumnManager.COLLAPSABLE_DATA_COLUMN_SHADOW,
      ]
      .map(col => {
        const columnKey: ColumnManager.ColumnKey = {
          effectiveLookupPath: [col],
          apparentLookupPath: col,
          displayString: col,
          shouldDisplay: true,
          serializedEffectiveLookupPath: JSON.stringify([col]),
        };
        return selectRenderer(columnKey);
      }).filter(colDef => colDef != null)
    )
  );

  const collapsedDataExtractedColumns: ColDef[] = ColumnManager.UsableColumns.getAllSingleLevelColumns()
    .filter(col => {
      return col.shouldDisplay && !ColumnManager.SERIALIZED_CUSTOMARY_COLUMN_KEYS.value.has(col.serializedEffectiveLookupPath)
    })
    .map(col => ({
      field: col.effectiveLookupPath[0],
      headerName: col.displayString,
      headerClass: col.derivedFromColumn == null
      ? 'my-ag-table-collapsible-data-expanded-column'
      : 'my-ag-table-collapsible-data-expanded-column my-ag-table-is-derived-column',
      cellClass: 'my-ag-table-collapsible-data-expanded-cell',
      cellRenderer: selectRenderer(col).cellRenderer,
      headerComponent: expandableCollapsibleDataColumnHeader,
      headerComponentParams: {
        key: col.effectiveLookupPath[0],
      },
      valueFormatter: (params: ValueFormatterParams) => {
        if (col.derivedFromColumn == null) {
          if (col.transformations?.includes(ColumnTransformation.NUMBER)) {
            return coerceToNumber(params.value) ?? "";
          } else {
            return params.value ?? "";
          }
        } else {
          return params.data[col.apparentLookupPath] ?? "";
        }
      },
    }));

    const expandableDataExtractedColumns: ColDef[] = ColumnManager.UsableColumns.getAllNestedDepthColumns()
    .filter(col => col.shouldDisplay)
    .sort((a, b) => a.displayString.localeCompare(b.displayString))
    .map(colKey => {
      return {
        field: ColumnManager.EXPANDABLE_DATA_COLUMN,
        headerName: colKey.displayString,
        headerClass: colKey.derivedFromColumn == null
        ? 'my-ag-table-expandable-data-expanded-column'
        : 'my-ag-table-expandable-data-expanded-column my-ag-table-is-derived-column',
        cellClass: 'my-ag-table-expandable-data-expanded-cell',
        headerComponent: expandableCollapsibleDataColumnHeader,
        headerComponentParams: {
          key: colKey,
        },
        valueFormatter: (params: ValueFormatterParams) => {
          if (colKey.derivedFromColumn != null) {
            return params.data[colKey.apparentLookupPath] ?? "";
          } else {
            // WARN: assuming 1 level of nesting for now
            // WARN: the ?? part is _probably_ when the expanded data is visible
            const value = params.data?.[ColumnManager.EXPANDABLE_DATA_COLUMN][colKey.serializedEffectiveLookupPath] ?? params.value?.[colKey.apparentLookupPath];
            if (colKey.transformations?.includes(ColumnTransformation.NUMBER)) {
              return coerceToNumber(value) ?? "";
            } else if (colKey.renderType == ColumnRendererType.TIME) {
              return value;
            } else {
              return value ?? "";
            }
          }
        },
        cellRenderer: selectRenderer(colKey).cellRenderer,
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
                ColumnManager.UsableColumns.getTotalHiddenNestedDepthColumns() == 0 ? "" : params.displayName
              ),
            ]),
            h('div', { class: 'button-container' }, [
              h('div', {}, [
                ColumnManager.UsableColumns.getTotalVisibleNestedDepthColumns() > 0 && h('button', {
                  class: 'ag-my-collapse-all',
                  title: 'Collapse all',
                  onClick: contractAllExpandableRows
                }, '◀'),
                ColumnManager.UsableColumns.getTotalHiddenNestedDepthColumns() > 0 && h('button', {
                  class: 'ag-my-expand-all',
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
  if (ColumnManager.UsableColumns.visibleNestedDepthColumnDataRowIds.value.has(rowIndex)) {
    ColumnManager.UsableColumns.visibleNestedDepthColumnDataRowIds.value.delete(rowIndex);
  } else {
    ColumnManager.UsableColumns.visibleNestedDepthColumnDataRowIds.value.add(rowIndex);
    const row = rowData.value[rowIndex];
    if (row && row[ColumnManager.EXPANDABLE_DATA_COLUMN]) {
      Object.keys(row[ColumnManager.EXPANDABLE_DATA_COLUMN]).forEach(key => {
        ColumnManager.UsableColumns.setDisplayOn([ColumnManager.EXPANDABLE_DATA_COLUMN, key]);
      });
    }
  }
  updateColumnDefs();
};

const toggleContractExpandableKeys = (rowIndex: number) => {
  if (!ColumnManager.UsableColumns.visibleNestedDepthColumnDataRowIds.value.has(rowIndex)) {
    return;
  } else {
    ColumnManager.UsableColumns.visibleNestedDepthColumnDataRowIds.value.delete(rowIndex);
    const row = rowData.value[rowIndex];
    if (row && row[ColumnManager.EXPANDABLE_DATA_COLUMN]) {
      Object.keys(row[ColumnManager.EXPANDABLE_DATA_COLUMN]).forEach(key => {
        ColumnManager.UsableColumns.setDisplayOff([ColumnManager.EXPANDABLE_DATA_COLUMN, key]);
      });
    }
  }
  updateColumnDefs();
}

const toggleExpandCollapsibleKeys = (rowIndex: number) => {
  console.log("!!!toggleExpandExtraDataKeys", rowIndex);
  if (ColumnManager.collapsibleDataExpandedRows.value.has(rowIndex)) {
    ColumnManager.collapsibleDataExpandedRows.value.delete(rowIndex);
  } else {
    ColumnManager.collapsibleDataExpandedRows.value.add(rowIndex);
    const row = rowData.value[rowIndex];
    if (row) {
      Object.keys(row).filter(
        key => !ColumnManager.SERIALIZED_CUSTOMARY_COLUMN_KEYS.value.has(key) && !ColumnManager.SPECIAL_COLUMN_KEYS.has(key)
      ).forEach(key => {
        ColumnManager.UsableColumns.setDisplayOn([key]);
      });
    }
  }
  updateColumnDefs();
}

const restoreAllCollapsedRows = () => {
  ColumnManager.UsableColumns.getAllSingleLevelColumns().forEach(col => col.shouldDisplay = true);
  updateColumnDefs();
};

const collapseAllCollapsibleRows = () => {
  ColumnManager.UsableColumns.getAllSingleLevelColumns().forEach(
    col => {
      if(ColumnManager.SERIALIZED_CUSTOMARY_COLUMN_KEYS.value.has(col.serializedEffectiveLookupPath)) {
        return;
      }
      col.shouldDisplay = false
    }
  );
  updateColumnDefs();
};


const expandAllExpandableRows = () => {
  let isDirty = false;
  rowData.value.forEach(row => {
    if (row.payload) {
      const payloadKeys = Object.keys(row.payload);
      if (payloadKeys.length > 0) {
        ColumnManager.UsableColumns.visibleNestedDepthColumnDataRowIds.value.add(row.id);
        payloadKeys.forEach(key => {
          ColumnManager.UsableColumns.setDisplayOn([ColumnManager.EXPANDABLE_DATA_COLUMN, key]);
        });

        isDirty = true;
      }
    }
  });
  if (isDirty) {
    updateColumnDefs();
  }
};

const contractAllExpandableRows = () => {
  ColumnManager.UsableColumns.getAllNestedDepthColumns().forEach(col => col.shouldDisplay = false);
  ColumnManager.UsableColumns.visibleNestedDepthColumnDataRowIds.value.clear();
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
import { parse } from 'vue/compiler-sfc';
const plotContainer = ref(null);

interface PlotSettings {
  xColumn: ColumnKey | null;
  yColumn: ColumnKey | null;
}
const plotSettings = ref<PlotSettings>({
  xColumn: null,
  yColumn: null,
});

// MOVEME
function _getIn(obj: any, path: string[]) {
  return path.reduce((acc, key) => acc && acc[key], obj);
}

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
    console.log("row", row);
    // SUPER DUPER RISKY AND DIRTY
    X.push(ColumnManager.parseTimeValue(
      _getIn(row, plotSettings.value.xColumn.effectiveLookupPath)
      // row[plotSettings.value.xColumn] ?? row[ColumnManager.EXPANDABLE_DATA_COLUMN]?.[plotSettings.value.xColumn]
    ));
    Y.push(
      _getIn(row, plotSettings.value.yColumn.effectiveLookupPath)
      // row[plotSettings.value.yColumn] ?? row[ColumnManager.EXPANDABLE_DATA_COLUMN]?.[plotSettings.value.yColumn]
    );
  }

  console.log("X", plotSettings.value.xColumn, X);
  console.log("Y", plotSettings.value.yColumn, Y);

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
        // Object.keys(row.payload).forEach(key => ColumnManager.detectedKeys.value.add(key));
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