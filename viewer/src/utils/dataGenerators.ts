import { faker } from '@faker-js/faker';


export function generateData(count: number) {

  return Array.from({ length: count }, () => {
    const sharedShapeData = {
      id: faker.string.uuid(),
      // .person.fullName(),
      // email: faker.internet.email(),
      // phone: faker.phone.number(),
      // company: faker.company.name(),
      country: faker.location.country(),
      createdAt: faker.date.past().toISOString(),
    }

    const extraData: { [key: string]: any } = {}
    const numberOfExtraColumns = Math.floor(Math.random() * 30) + 1
    for (let i = 0; i < numberOfExtraColumns; i++) {
      if (Math.random() < 0.6) {
        continue;
      }
      const columnName = `e-${i}`;
      extraData[columnName] = faker.lorem.words()
    }

    const extraNonPayloadData: { [key: string]: any } = {}
    for (let i = 0; i < 4; ++i) {
      if (Math.random() < 0.5) {
        continue;
      }
      const columnName = `x${i}`;
      extraNonPayloadData[columnName] = faker.animal.type()
    }

    return { ...sharedShapeData, payload: JSON.stringify(extraData), ...extraNonPayloadData }
  })
}
