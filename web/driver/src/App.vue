<template>
  <div class="driver-phone-shell">
    <div id="driver-home-popups"></div>
    <router-view v-slot="{ Component }">
      <transition name="fade" mode="out-in">
        <component :is="Component" />
      </transition>
    </router-view>
  </div>
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
  --driver-primary: #6d4aff;
  --driver-primary-dark: #4b2bc5;
  --driver-accent: #ffbe2e;
  --driver-success: #16a34a;
  --driver-danger: #dc2626;
  --driver-ink: #111827;
  --driver-muted: #667085;
  --driver-line: #e6eaf2;
  --driver-bg: #eef2f7;
  --driver-card: #ffffff;
}

* { box-sizing: border-box; }
html, body, #app { width: 100%; min-height: 100%; margin: 0; }
body {
  min-height: 100vh;
  display: flex;
  justify-content: center;
  background: #d9dee8;
  color: var(--driver-ink);
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}
button, input, textarea { font: inherit; }
button { -webkit-tap-highlight-color: transparent; }
a { color: inherit; text-decoration: none; }
#app { width: 100%; min-height: 100vh; display: flex; justify-content: center; }
.driver-phone-shell {
  position: relative;
  width: min(100vw, 390px);
  min-height: 100vh;
  margin: 0 auto;
  overflow-x: hidden;
  background: var(--driver-bg);
  box-shadow: 0 0 0 1px rgba(15, 23, 42, 0.08), 0 24px 60px rgba(15, 23, 42, 0.18);
}
.fade-enter-active, .fade-leave-active { transition: opacity .16s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
@media (max-width: 430px) {
  body { background: var(--driver-bg); }
  .driver-phone-shell { width: 100vw; box-shadow: none; }
}
</style>
