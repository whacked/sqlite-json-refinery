<template>
    <button @click="loadFile">Load JSONL File</button>
    <button @click="exampleLoadRemoteJsonl">Load Remote JSONL</button>
    <button @click="exampleLoadRemoteData">Load Remote Data</button>
    <button @click="exampleLoadFiniteFakeData">Load test Data</button>
    <button @click="loadInfiniteFakeData">Load Infinite Fake Data</button>
    <button @click="loadInfiniteJsonlData">Load Infinite JSONL Data</button>
    <code v-if="agGridDataProvider && agGridDataProvider.rows">
      {{ Array.from(agGridDataProvider.rows.values()).length }} items in cache
    </code>
    <DataTable :ref="dataTableRef" :dataProvider="agGridDataProvider" />
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
import * as ColumnManager from '@/utils/columnManager';
import { faker } from '@faker-js/faker';
import "ag-grid-community/styles/ag-grid.css"; // Mandatory CSS required by the Data Grid
import "ag-grid-community/styles/ag-theme-quartz.css"; // Optional Theme applied to the Data Grid
import { onMounted, ref, VNodeRef } from 'vue';
import DataTable from '../components/DataTable.vue';
import { loadRemoteData, loadRemoteJsonl } from '../stores/remoteDataLoader';
import { generateData } from '../utils/dataGenerators';

// TODO move this to autogen
interface CommonPayloadData {
    time: number;
    category: string;
    entry: string;
    payload: Record<string, string>;
}


// Fake data generator
const generateFakeRow = (index: number): CommonPayloadData => ({
    time: index,
    category: faker.location.city(),
    entry: faker.person.fullName(),
    payload: { email: faker.internet.email() },
});

// Fake async data fetcher
const fakeAsyncDataFetcher = (startRow: number, endRow: number): Promise<CommonPayloadData[]> => {
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

interface RowFetchWindow {
    rows: any[],
    startIndex: number,
    endIndex: number,
    totalRowCount: number,
}


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

async function loadFile() {
    const fileHandle = await getFileHandle();
    const file = 'getFile' in fileHandle ? await fileHandle.getFile() : fileHandle;
    const text = await file.text();
    const lines = text.split('\n');

    const parsedRows: any[] = [];
    for (const line of lines) {
        if (line.trim() !== '') {
            try {
                const parsedLine = JSON.parse(line);
                parsedRows.push(parsedLine);
            } catch (error) {
                console.error('Error parsing line:', line);
                console.error(error);
            }
        }
    }
    console.log(`fileCache length from ${file.name}:`, parsedRows.length);
    agGridDataProvider.value = {
        name: file.name,
        isInfinite: false,
        rows: parsedRows,
    };
}



const agGridDataProvider = ref<ColumnManager.AgGridDataProvider>();
const dataTableRef = ref<VNodeRef | null>(null);

const exampleLoadRemoteJsonl = async () => {
    const data = await loadRemoteJsonl(234, 567);
    agGridDataProvider.value = {
        name: "remoteJsonlData",
        isInfinite: false,
        rows: data.rows,
    };
}

const exampleLoadRemoteData = async () => {
    const data = await loadRemoteData(345, 678);
    agGridDataProvider.value = {
        name: "exampleRemoteData",
        isInfinite: false,
        rows: data.rows,
    };
}

const exampleLoadFiniteFakeData = async () => {
    agGridDataProvider.value = {
        name: "exampleFiniteFakeData",
        isInfinite: false,
        rows: generateData(19),
    };
}

const loadInfiniteJsonlData = async () => {
    // MOVE ME: this is the infinite datasource
    agGridDataProvider.value = {
        isInfinite: true,
        name: "infiniteJsonlData",
        rows: undefined,
        infiniteDataGetter: async (startRow: number, endRow: number) => {
            console.log('Fetching rows:', startRow, 'to', endRow);
            return loadRemoteJsonl(startRow, endRow).then(data => {
                return {
                    rows: data.rows,
                    totalRowCount: data.totalRowCount,
                };
            });
        }
    };
}

const loadInfiniteFakeData = async () => {
    // MOVE ME: this is the infinite datasource
    agGridDataProvider.value = {
        isInfinite: true,
        name: "infiniteFakeData",
        rows: undefined,
        infiniteDataGetter: async (startRow: number, endRow: number) => {
            console.log('Fetching rows:', startRow, 'to', endRow);
            const data = generateData(endRow - startRow);
            return {
                rows: data,
                totalRowCount: 123456,
            };
        }
    };
}

onMounted(async () => {
    exampleLoadFiniteFakeData();
});
</script>