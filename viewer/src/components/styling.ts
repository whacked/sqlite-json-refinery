import chroma from "chroma-js";
import ColorHash from "color-hash";

export const colorHash = new ColorHash();


export namespace Colorizer {
    let getColorCache: Record<string, string> = {};
    let makeTextContainerStyleCache: Record<string, {
        backgroundColor: string;
        color: string;
    }> = {};

    export function getColor(value: any) {
        if (getColorCache[value]) return getColorCache[value];
        const color = colorHash.hex(value);
        getColorCache[value] = color;
        return color;
    }

    export function makeTextContainerStyle(text: string | null): {
        backgroundColor: string;
        color: string;
    } {
        if (text == null) {
            return {
                backgroundColor: '',
                color: '',
            }
        } else if (makeTextContainerStyleCache[text] == null) {
            const backgroundColor = getColor(text);
            const isDark = chroma(backgroundColor).luminance() < 0.6;
            const textColor = isDark ? 'white' : 'black';
            const style = {
                backgroundColor,
                color: textColor,
            }
            makeTextContainerStyleCache[text] = style;
        }
        return makeTextContainerStyleCache[text];
    }
}