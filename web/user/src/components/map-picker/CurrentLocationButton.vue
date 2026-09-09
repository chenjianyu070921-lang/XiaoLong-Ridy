<template>
  <button
    type="button"
    class="locate-btn"
    :disabled="loading"
    aria-label="回到当前位置"
    @click="$emit('locate')"
  >
    <van-loading v-if="loading" size="18px" color="#7C3AED" />
    <van-icon v-else name="aim" size="20" color="#7C3AED" />
  </button>
</template>

<script setup>
defineProps({
  loading: { type: Boolean, default: false }
})

defineEmits(['locate'])
</script>

<style scoped>
.locate-btn {
  width: 46px;
  height: 46px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.96);
  box-shadow: 0 4px 14px rgba(15, 23, 42, 0.18);
  cursor: pointer;
}

.locate-btn:disabled { opacity: 0.6; cursor: not-allowed; }
</style>
