<template>
    <div class="test-container">
        <h3>PhotoCell Test Environment</h3>

        <!-- Test controls -->
        <div class="controls">
            <button @click="currentTest = 'single'">Test Single Photo</button>
            <button @click="currentTest = 'multiple'">Test Multiple Photos</button>
            <button @click="currentTest = 'null'">Test Null</button>
        </div>

        <!-- Test component -->
        <div class="test-component">
            <PhotoCell :params="testParams" />
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import PhotoCell from '../components/PhotoCell.vue';

// Test data
const TEST_PHOTOS = {
    single: { photo: '20240315_123456.jpg' },
    multiple: { photo: ['20240315_123456.jpg', '20240315_123457.jpg', '20240315_123458.jpg'] },
    null: { photo: null }
};

const currentTest = ref('single');

const testParams = computed(() => ({
    data: TEST_PHOTOS[currentTest.value as keyof typeof TEST_PHOTOS]
}));
</script>

<style scoped>
.test-container {
    padding: 20px;
}

.controls {
    margin-bottom: 20px;
}

.controls button {
    margin-right: 10px;
}

.test-component {
    border: 1px solid #ccc;
    padding: 20px;
}
</style>