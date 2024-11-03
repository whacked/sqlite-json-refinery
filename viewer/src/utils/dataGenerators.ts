import { faker } from '@faker-js/faker';


export function generateData(count: number) {

  const fakeColumns: string[] = []

  const NUM_FAKE_COLUMNS = 3;

  for (let i = 0; i < NUM_FAKE_COLUMNS; ++i) {
    break;
    const fakeColumn = faker.lorem.word()
    fakeColumns.push(fakeColumn)
    if (i > 0) {
      const level2Column = `${fakeColumn}.${faker.lorem.word()}`
      fakeColumns.push(level2Column)
      if (i > 1) {
        const level3Column = `${level2Column}.${faker.lorem.word()}`
        fakeColumns.push(level3Column)
      }
    }
  }

  console.log("fakeColumns", fakeColumns);

  const randomCurrencies: string[] = [''];
  for (let i = 0; i < 5; ++i) {
    randomCurrencies.push(faker.finance.currencyCode());
  }
  function getRandomCurrency() {
    return randomCurrencies[Math.floor(Math.random() * randomCurrencies.length)];
  }

  return Array.from({ length: count }, () => {
    const sharedShapeData = {
      //*
      id: faker.string.uuid(),
      // .person.fullName(),
      // email: faker.internet.email(),
      // phone: faker.phone.number(),
      // company: faker.company.name(),
      country: faker.location.country(),
      createdAt: faker.date.past().toISOString(),
      // */
      price: `${(Math.random() * 100).toFixed(2)}${getRandomCurrency()}`,
      temperature: Math.random() > 0.5 ? `${Math.floor(Math.random() * 100)}.` : Math.random(),
      ...Object.fromEntries(fakeColumns.map(column => [column, faker.lorem.word()])),
    }

    const expandablePayloadStringData: { [key: string]: any } = {}
    expandablePayloadStringData["price"] = `${(Math.random() * 100).toFixed(2)}${getRandomCurrency()}`
    expandablePayloadStringData["quantity"] = Math.random() > 0.05 ? `${Math.floor(Math.random() * 100)}p` : null

    /* const numberOfExtraColumns = Math.floor(Math.random() * 30) + 1 */
    const numberOfExtraColumns = 4
    for (let i = 0; i < numberOfExtraColumns; i++) {
      break;
      if (Math.random() < 0.6) {
        continue;
      }
      const columnName = `e-${i}`;
      expandablePayloadStringData[columnName] = faker.lorem.word()
    }

    const extraNonPayloadData: { [key: string]: any } = {
      "foo": `${Math.random() * 100}x`,
    }
    for (let i = 0; i < 4; ++i) {
      break;
      if (Math.random() < 0.5) {
        continue;
      }
      const columnName = `x${i}`;
      extraNonPayloadData[columnName] = faker.animal.type()
    }

    return { ...sharedShapeData, /* payload: JSON.stringify(expandablePayloadStringData), */ /* ...extraNonPayloadData */ }
  })
}
