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

/* 夜间模式：覆盖设计变量，并给 <html class="dark"> 下的容器做全局兜底，
   使所有页面背景转黑、文字转白（头像/图片不受影响）。 */
html.dark {
  --driver-primary: #8b6dff;
  --driver-primary-dark: #6d4aff;
  --driver-accent: #ffbe2e;
  --driver-success: #22c55e;
  --driver-danger: #ef4444;
  --driver-ink: #f5f6fa;
  --driver-muted: #9aa3b2;
  --driver-line: #2a2e3a;
  --driver-bg: #0b0d12;
  --driver-card: #161a22;
}
html.dark body { background: var(--driver-bg); }
html.dark .driver-phone-shell { background: var(--driver-bg); color: var(--driver-ink); }
html.dark .driver-phone-shell * { color: var(--driver-ink); }
html.dark .driver-phone-shell section,
html.dark .driver-phone-shell .section-block,
html.dark .driver-phone-shell .group-panel,
html.dark .driver-phone-shell .mine-list,
html.dark .driver-phone-shell .income-card,
html.dark .driver-phone-shell [class*="card"],
html.dark .driver-phone-shell [class$="-panel"],
html.dark .driver-phone-shell [class$="-page"] {
  background: var(--driver-card) !important;
  color: var(--driver-ink) !important;
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
