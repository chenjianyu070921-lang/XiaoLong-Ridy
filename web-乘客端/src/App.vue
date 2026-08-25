<template>
  <router-view v-slot="{ Component }">
    <transition name="fade" mode="out-in">
      <component :is="Component" />
    </transition>
  </router-view>
</template>

<script setup>
import { onMounted } from 'vue'

onMounted(() => {
  const meta = document.querySelector('meta[name="viewport"]') || document.createElement('meta')
  meta.name = 'viewport'
  meta.content = 'width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no, viewport-fit=cover'
  if (!meta.parentNode) document.head.appendChild(meta)
})
</script>

<style>
:root {
  --primary-color: #7c3aed;
  --primary-light: #8b5cf6;
  --primary-dark: #6d28d9;
  --secondary-color: #10b981;
  --warning-color: #f59e0b;
  --danger-color: #ef4444;
  --text-primary: #111827;
  --text-secondary: #4b5563;
  --text-light: #9ca3af;
  --bg-color: #f5f7fb;
  --bg-white: #ffffff;
  --border-color: #e5e7eb;
}

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html, body, #app {
  width: 100%;
  min-height: 100%;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  background: var(--bg-color);
  color: var(--text-primary);
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

a {
  color: inherit;
  text-decoration: none;
}

button, input, textarea {
  font: inherit;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.btn-primary {
  border: 0;
  border-radius: 16px;
  background: linear-gradient(135deg, #7c3aed 0%, #8b5cf6 100%);
  color: #fff;
  font-weight: 600;
  box-shadow: 0 10px 24px rgba(124, 58, 237, 0.24);
}
</style>
