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

// Use environment variable or build-time config for this
const IS_TEST_MODE = import.meta.env.MODE === 'development';
const TEST_PHOTO_URLS = [
    'https://picsum.photos/200/300',
    'https://picsum.photos/200/301',
    'https://picsum.photos/200/302'
];

function makePhotoPath(photo: string) {
    if (IS_TEST_MODE) {
        // In test mode, return a placeholder image
        const index = parseInt(photo.replace(/\D/g, '')) % TEST_PHOTO_URLS.length;
        return TEST_PHOTO_URLS[index];
    }

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