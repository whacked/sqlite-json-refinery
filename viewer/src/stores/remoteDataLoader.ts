import { DefaultApi } from "@/openapi-interface.autogen";

const WS4SQL_SERVER_URL = 'http://localhost:12321';
const DATABASE_NAME = 'mqtt'


async function runQuery(query: string) {
    const jsonPayload = {
        transaction: [{
            query: query,
        }]
    }
    const response = await fetch(`${WS4SQL_SERVER_URL}/${DATABASE_NAME}`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(jsonPayload),
    });
    return response.json();
}

async function getTotalRowCount() {
    const response = await runQuery(`SELECT COUNT(*) FROM commonpayloaddata`);
    return response.results?.[0]?.resultSet?.[0]?.['COUNT(*)'];
}


export async function loadRemoteData(
    offset: number = 0,
    limit: number = 10,
) {
    const totalRowCount = await getTotalRowCount();
    const rowsResult = await runQuery(
        `SELECT * FROM commonpayloaddata ORDER BY id DESC LIMIT ${limit} OFFSET ${offset}`
    )
    const rows = rowsResult.results?.[0]?.resultSet;
    return Promise.resolve({
        rows,
        totalRowCount,
    });
}


export async function loadRemoteJsonl(
    offset: number = 0,
    limit: number = 10,
) {
    const api = new DefaultApi();
    const totalRowCount = await api.countRecords();
    const rowsResult = await api.listRecords({ offset, limit });
    return Promise.resolve({
        rows: rowsResult,
        totalRowCount,
    });
}