import { ref, computed } from 'vue';

export function useWindowingService(pageSize: number) {
  const currentPage = ref(1);
  const totalItems = ref(0);

  const startIndex = computed(() => (currentPage.value - 1) * pageSize);
  const endIndex = computed(() => Math.min(startIndex.value + pageSize, totalItems.value));

  function setTotalItems(total: number) {
    totalItems.value = total;
  }

  function nextPage() {
    if (endIndex.value < totalItems.value) {
      currentPage.value++;
    }
  }

  function previousPage() {
    if (currentPage.value > 1) {
      currentPage.value--;
    }
  }

  return {
    currentPage,
    totalItems,
    startIndex,
    endIndex,
    setTotalItems,
    nextPage,
    previousPage,
  };
}
