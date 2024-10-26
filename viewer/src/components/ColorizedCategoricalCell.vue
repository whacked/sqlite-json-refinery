<template>
    <div style="background:{{ backgroundColor }};color:{{ textColor }}">
        {{ params.data.topic }}
    </div>
</template>

<script setup lang="ts">
const props = defineProps<{
    params: RenderParams;
}>();

(params: ColumnManager.RenderParams, _: any) => {
        const topic = params.data.topic;
        const backgroundColor = colorHash.hex(topic);
        const isDark = chroma(backgroundColor).luminance() < 0.6;
        const textColor = isDark ? 'white' : 'black';
        return h(ColorizedCategoricalCell, {
          params,
        });
      }
</script>