<template>
  <div class="driver-phone-shell">
    <div id="driver-home-popups"></div>
    <router-view v-slot="{ Component }">
      <transition name="fade" mode="out-in">
        <!-- 仅缓存 DriverHome：进入独立私信页 /chat/:orderId 时不卸载首页，
             其司机 WS / 心跳 / 定位 / 派单等实时连接随司机会话全程存活，聊天不随切页断开。 -->
        <keep-alive include="DriverHome">
          <component :is="Component" />
        </keep-alive>
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
  --driver-soft: #f3f5f9;
  --driver-faint: #98A2B3;
  --driver-on-primary: #ffffff;
  --driver-track: rgba(0, 0, 0, 0.08);
  --driver-st-ongoing-bg: #EFF6FF; --driver-st-ongoing-fg: #3B82F6;
  --driver-st-completed-bg: #ECFDF5; --driver-st-completed-fg: #059669;
  --driver-st-pending-bg: #FEF3C7; --driver-st-pending-fg: #D97706;
  --driver-st-cancelled-bg: #FEE2E2; --driver-st-cancelled-fg: #DC2626;
}

/* 夜间模式：覆盖设计变量，并给 <html class="dark"> 下的容器做全局兜底，
   使所有页面背景转黑、文字转白（头像/图片不受影响）。 */
html.dark {
  --driver-primary: #FFBC2C;
  --driver-primary-dark: #E6A800;
  --driver-accent: #FFBC2C;
  --driver-success: #22c55e;
  --driver-danger: #ef4444;
  --driver-ink: #FFFFFF;
  --driver-muted: #D0D5DD;
  --driver-faint: #98A2B3;
  --driver-line: #383E4C;
  --driver-bg: #1E222B;
  --driver-card: #2A2F3B;
  --driver-soft: rgba(255, 255, 255, 0.05);
  --driver-on-primary: #1E222B;
  --driver-track: rgba(255, 255, 255, 0.16);
  --driver-st-ongoing-bg: rgba(59, 130, 246, 0.2); --driver-st-ongoing-fg: #93c5fd;
  --driver-st-completed-bg: rgba(16, 185, 129, 0.18); --driver-st-completed-fg: #6ee7b7;
  --driver-st-pending-bg: rgba(245, 158, 11, 0.18); --driver-st-pending-fg: #fcd34d;
  --driver-st-cancelled-bg: rgba(239, 68, 68, 0.18); --driver-st-cancelled-fg: #fca5a5;
  /* index.css 通用变量同步深灰+暖黄 */
  --primary-color: #FFBC2C;
  --primary-light: #FFC94D;
  --primary-dark: #E6A800;
  --text-primary: #FFFFFF;
  --text-secondary: #D0D5DD;
  --text-light: #98A2B3;
  --bg-color: #1E222B;
  --bg-white: #2A2F3B;
  --border-color: #383E4C;
  /* vant 组件夜间覆盖 */
  --van-background: #1E222B;
  --van-background-2: #2A2F3B;
  --van-background-3: #383E4C;
  --van-text-color: #FFFFFF;
  --van-text-color-2: #D0D5DD;
  --van-text-color-3: #98A2B3;
  --van-border-color: #383E4C;
  --van-primary-color: #FFBC2C;
  --van-button-primary-background: #FFBC2C;
  --van-button-primary-border-color: #FFBC2C;
  --van-button-primary-color: #1E222B;
  --van-cell-background: #2A2F3B;
  --van-cell-background-active: #383E4C;
  --van-field-input-text-color: #FFFFFF;
  --van-field-placeholder-text-color: #98A2B3;
  --van-dialog-background: #2A2F3B;
  --van-popup-background: #2A2F3B;
  --van-dropdown-menu-background: #2A2F3B;
  --van-dropdown-menu-title-text-color: #D0D5DD;
  --van-nav-bar-background: #2A2F3B;
  --van-nav-bar-title-text-color: #FFFFFF;
  --van-toast-background: rgba(42, 47, 59, .92);
  --van-rate-icon-void-color: #383E4C;
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
html, body, #app { width: 100%; height: 100%; margin: 0; }
html, body { overflow: hidden; }
body {
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
  height: 100vh;
  margin: 0 auto;
  overflow-x: hidden;
  overflow-y: auto;
  background: var(--driver-bg);
  box-shadow: 0 0 0 1px rgba(15, 23, 42, 0.08), 0 24px 60px rgba(15, 23, 42, 0.18);
  scrollbar-width: none;            /* Firefox：默认隐藏滚动条 */
  -ms-overflow-style: none;         /* IE/旧 Edge：默认隐藏 */
}
.driver-phone-shell::-webkit-scrollbar {
  width: 0;
  height: 0;
  background: transparent;
}
.driver-phone-shell:hover::-webkit-scrollbar,
.driver-phone-shell:focus-within::-webkit-scrollbar {
  width: 6px;                       /* 悬浮/聚焦滚动区域时才显示 */
}
.driver-phone-shell::-webkit-scrollbar-thumb {
  border-radius: 3px;
  background: rgba(15, 23, 42, .22);
}
.driver-phone-shell::-webkit-scrollbar-track {
  background: transparent;
}
.fade-enter-active, .fade-leave-active { transition: opacity .16s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
@media (max-width: 430px) {
  body { background: var(--driver-bg); }
  .driver-phone-shell { width: 100vw; box-shadow: none; }
}
</style>
