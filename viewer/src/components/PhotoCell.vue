<template>
    <div class="pc-photo-container">
        <!-- Main photo display -->
        <div class="pc-photo-wrapper" @click="showModal = true">
            <img :src="currentDisplayPhoto" :alt="currentDisplayPhoto" class="pc-photo-cell" />
            <div v-if="totalPhotos > 1" class="pc-photo-count">
                {{ totalPhotos }} photos
            </div>
            <div v-if="totalPhotos > 1">
                YOYO {{ totalPhotos }} photos
            </div>
        </div>

        <!-- Modal - using teleport to move it to body -->
        <Teleport to="body">
            <div v-if="showModal" class="pc-modal" @click.self="closeModal" @keydown="handleKeyPress">
                <div class="pc-modal-content">
                    <img :src="currentModalPhoto" :alt="currentModalPhoto" class="pc-modal-image" />

                    <!-- Navigation buttons for multiple photos -->
                    <template v-if="totalPhotos > 1">
                        <button class="pc-nav-button pc-prev" @click="prevPhoto">&lt;</button>
                        <button class="pc-nav-button pc-next" @click="nextPhoto">&gt;</button>
                        <div class="pc-photo-indicator">{{ currentIndex + 1 }} / {{ totalPhotos }}</div>
                    </template>
                </div>
            </div>
        </Teleport>
    </div>
</template>

<style>
/* Using 'pc-' prefix for PhotoCell component to avoid conflicts */
.pc-photo-cell {
    height: 200px;
    width: auto;
    object-fit: contain;
    border-radius: 8px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.pc-photo-wrapper {
    position: relative;
    display: inline-block;
    cursor: pointer;
}

.pc-photo-container {
    position: relative;
}

.pc-photo-wrapper {
    position: relative;
    display: inline-block;
    cursor: pointer;
}

.pc-photo-cell {
    height: 200px;
    width: auto;
    object-fit: contain;
    border-radius: 8px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.pc-photo-count {
    position: absolute;
    top: 10px;
    right: 10px;
    background-color: rgba(255, 0, 0, 0.8);
    color: white;
    padding: 4px 8px;
    border-radius: 12px;
    font-size: 0.9em;
}

.pc-modal {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    /* Use viewport units */
    height: 100vh;
    background-color: rgba(0, 0, 0, 0.8);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 9999;
    /* Higher z-index to ensure it's above everything */
}

.pc-modal-content {
    position: relative;
    max-width: calc(100vw - 40px);
    margin: 20px;
}

.pc-modal-image {
    max-height: calc(100vh - 40px);
    width: auto;
    max-width: calc(100vw - 40px);
    border-radius: 8px;
    border: 2px solid white;
    object-fit: contain;
}

.pc-nav-button {
    position: absolute;
    top: 50%;
    transform: translateY(-50%);
    background-color: rgba(255, 255, 255, 0.8);
    border: none;
    border-radius: 20%;
    width: 20px;
    height: 40px;
    margin: 0 -10px 0 -10px;
    cursor: pointer;
    font-size: 20px;
    display: flex;
    align-items: center;
    justify-content: center;
}

.pc-nav-button:hover {
    background-color: rgba(255, 255, 255, 0.9);
}

.pc-prev {
    left: -50px;
}

.pc-next {
    right: -50px;
}

.pc-photo-indicator {
    position: absolute;
    bottom: -30px;
    left: 50%;
    transform: translateX(-50%);
    color: white;
    background-color: rgba(0, 0, 0, 0.5);
    padding: 5px 10px;
    border-radius: 12px;
}
</style>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue';
// Add Teleport import if needed (it's built into Vue 3, so might not be necessary)
// import { Teleport } from 'vue';

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
// const IS_TEST_MODE = import.meta.env.MODE === 'development';
const IS_TEST_MODE = !true;
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

const showModal = ref(false);
const currentIndex = ref(0);

const totalPhotos = computed(() => photoPaths.value.length);
const currentDisplayPhoto = computed(() => photoPaths.value[0] || '');
const currentModalPhoto = computed(() => photoPaths.value[currentIndex.value] || '');

function closeModal() {
    showModal.value = false;
    currentIndex.value = 0;
}

function nextPhoto() {
    currentIndex.value = (currentIndex.value + 1) % totalPhotos.value;
}

function prevPhoto() {
    currentIndex.value = (currentIndex.value - 1 + totalPhotos.value) % totalPhotos.value;
}

function handleKeyPress(event: KeyboardEvent) {
    if (event.key === 'Escape') {
        closeModal();
    } else if (totalPhotos.value > 1) {
        if (event.key === 'ArrowLeft') {
            prevPhoto();
        } else if (event.key === 'ArrowRight') {
            nextPhoto();
        }
    }
}

// Event listeners for keyboard navigation
onMounted(() => {
    window.addEventListener('keydown', handleKeyPress);
});

onUnmounted(() => {
    window.removeEventListener('keydown', handleKeyPress);
});
</script>