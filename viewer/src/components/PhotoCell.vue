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
    'data:image/svg+xml,%3Csvg width="200" height="300" xmlns="http://www.w3.org/2000/svg"%3E%3Crect width="100%25" height="100%25" fill="%23ddd"/%3E%3Ctext x="50%25" y="50%25" dominant-baseline="middle" text-anchor="middle" font-family="sans-serif"%3ETest Image 1%3C/text%3E%3C/svg%3E',
    'data:image/svg+xml,%3Csvg width="200" height="300" xmlns="http://www.w3.org/2000/svg"%3E%3Crect width="100%25" height="100%25" fill="%23ddd"/%3E%3Ctext x="50%25" y="50%25" dominant-baseline="middle" text-anchor="middle" font-family="sans-serif"%3ETest Image 2%3C/text%3E%3C/svg%3E',
    'data:image/svg+xml,%3Csvg width="200" height="300" xmlns="http://www.w3.org/2000/svg"%3E%3Crect width="100%25" height="100%25" fill="%23ddd"/%3E%3Ctext x="50%25" y="50%25" dominant-baseline="middle" text-anchor="middle" font-family="sans-serif"%3ETest Image 3%3C/text%3E%3C/svg%3E'
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