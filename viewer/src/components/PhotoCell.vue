<template>
    <div>
        <img v-for="photo in photoPaths" height="200" :src="photo" :alt="photo" class="photo-cell" />
    </div>
</template>

<style scoped>
img.photo-cell {
    /* height: 500px; */
}
</style>

<script setup lang="ts">
import { computed } from 'vue';

const props = defineProps<{
    params: {
        data: {
            photo: string | string[] | null;
        };
    };
}>();

// FIXME
// currently launched using
// static-web-server --root $STATICDATA/mmc/DCIM --port 7002
const BASE_PHOTO_PATH = 'http://localhost:7002/';

function makePhotoPath(photo: string) {
    const photoYear = photo.substring(0, 4);
    const photoMonth = photo.substring(4, 6);
    return `${BASE_PHOTO_PATH}/${photoYear}/${photoMonth}/${photo}`;
}

const photoPaths = computed(() => {
    const photo = props.params.data.photo;
    if (!photo) {
        return [];
    }
    return typeof photo === 'string' ? [makePhotoPath(photo)] : photo.map(makePhotoPath);
});
</script>