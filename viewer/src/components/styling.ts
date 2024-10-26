import ColorHash from "color-hash";

export const colorHash = new ColorHash();


export namespace Colorizer {
    let cache: Record<string, string> = {};

    export function getColor(value: any) {
        if (cache[value]) return cache[value];
        const color = colorHash.hex(value);
        cache[value] = color;
        return color;
    }
}