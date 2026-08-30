<template>
  <router-view />
</template>

<style>
/* 全局主题基线：默认浅色紫，主题类由 App 启动时根据本地偏好设置。 */
:root{font-family:Inter,"PingFang SC","Microsoft YaHei",sans-serif;color:#2e2c4e;background:#f2f3fa;font-synthesis:none;text-rendering:optimizeLegibility}
:root.dark-theme{color:#e6e3f5;background:#151325}
/* Element Plus 品牌主色：紫色系，深浅主题共用。 */
:root{--el-color-primary:#6c5ce7;--el-color-primary-light-3:#8b7eee;--el-color-primary-light-5:#a89ff2;--el-color-primary-light-7:#c6c0f7;--el-color-primary-light-8:#d9d5fa;--el-color-primary-light-9:#edebfd;--el-color-primary-dark-2:#584ac9}
*{box-sizing:border-box}body{margin:0;min-width:1100px;background:var(--page-bg,#f2f3fa);color:var(--text-color,#2e2c4e);transition:background-color .2s,color .2s}button,input{font:inherit}
/* 表格：紫色表头、紫色行悬停，卡片内无边框感。 */
.el-table{--el-table-bg-color:var(--panel-bg,#ffffff);--el-table-tr-bg-color:var(--panel-bg,#ffffff);--el-table-header-bg-color:var(--table-head-bg,#f6f5fd);--el-table-row-hover-bg-color:var(--hover-bg,#f4f2fd);--el-table-striped-bg-color:var(--stripe-bg,#faf9fe)!important;--el-table-border-color:var(--border-color,#e5e4f0);--el-table-text-color:var(--text-color,#2e2c4e);--el-table-header-text-color:var(--muted-color,#8b88a3)}
.el-table th.el-table__cell{font-weight:600}
/* 斑马纹兜底：直接命中条纹单元格，确保按需加载的表格样式无法覆盖。 */
.el-table--striped .el-table__body tr.el-table__row--striped td.el-table__cell{background:var(--stripe-bg,#faf9fe)!important}
.el-input__wrapper,.el-select__wrapper,.el-textarea__inner{background:var(--input-bg,#ffffff);box-shadow:0 0 0 1px var(--border-color,#e5e4f0) inset;color:var(--text-color,#2e2c4e);border-radius:8px}
.el-input__inner{color:var(--text-color,#2e2c4e)}
.el-pagination{--el-pagination-bg-color:var(--panel-bg,#ffffff);--el-pagination-text-color:var(--muted-color,#8b88a3);--el-pagination-button-color:var(--text-color,#2e2c4e)}
.el-dialog{background:var(--panel-bg,#ffffff);border:1px solid var(--border-color,#e5e4f0);border-radius:14px}
.el-dialog__title,.el-form-item__label{color:var(--text-color,#2e2c4e)}
.el-button{border-radius:8px}
.el-menu-item{height:44px;line-height:44px}
.el-message-box{background:var(--panel-bg,#ffffff);border:1px solid var(--border-color,#e5e4f0)}
.el-message-box__title,.el-message-box__content{color:var(--text-color,#2e2c4e)}
.el-tabs__item{color:var(--muted-color,#8b88a3)}.el-tabs__item.is-active{color:var(--el-color-primary)}
/* 交互状态：悬停、聚焦、选中在所有主题下统一表现。 */
::selection{background:rgba(108,92,231,.32);color:inherit}
:focus-visible{outline:2px solid var(--brand,#6c5ce7);outline-offset:1px;border-radius:4px}
.el-input__wrapper:hover,.el-select__wrapper:hover,.el-textarea__inner:hover{box-shadow:0 0 0 1px var(--brand,#6c5ce7) inset}
.el-input__wrapper.is-focus,.el-select__wrapper.is-focused,.el-textarea__inner:focus{box-shadow:0 0 0 1px var(--brand,#6c5ce7) inset}
.el-button--primary:not(.is-disabled):not(.is-link):hover{box-shadow:0 4px 14px rgba(108,92,231,.4)}
.el-table__body tr:hover>td.el-table__cell{transition:background-color .15s}
/* 浅色主题：原型图浅紫灰底 + 白色卡片。 */
:root.light-theme{--brand:#6c5ce7;--brand-strong:#584ac9;--page-bg:#f2f3fa;--panel-bg:#ffffff;--header-bg:#ffffff;--table-head-bg:#f6f5fd;--input-bg:#ffffff;--hover-bg:#f4f2fd;--active-bg:#ece9fb;--border-color:#e5e4f0;--text-color:#2e2c4e;--muted-color:#8b88a3;--card-shadow:0 1px 3px rgba(46,44,78,.06),0 6px 18px rgba(108,92,231,.06);--aside-grad:linear-gradient(180deg,#6a5ae2 0%,#5847c9 60%,#4f41bc 100%);--stripe-bg:#faf9fe}
/* 深色主题：低饱和紫黑底 + 分层卡片，主色提亮保证对比度。 */
:root.dark-theme{--brand:#9d8ff7;--brand-strong:#b7adf9;--page-bg:#100e1a;--panel-bg:#1b1829;--panel-bg-2:#242138;--header-bg:#181526;--table-head-bg:#252137;--input-bg:#141221;--hover-bg:#2a2544;--active-bg:#363057;--border-color:#2e2a46;--text-color:#e9e7f6;--muted-color:#a09bc4;--card-shadow:0 1px 2px rgba(0,0,0,.45),0 10px 28px rgba(0,0,0,.28);--aside-grad:linear-gradient(180deg,#282250 0%,#201c42 55%,#1a1734 100%);--stripe-bg:#282541}
/* 深色模式下 Element Plus 基础色板：主色提亮，弹出层、填充、边框、禁用态全部暗化。 */
:root.dark-theme{--el-color-primary:#8d7ef3;--el-color-primary-dark-2:#a396f6;--el-color-primary-light-3:#7a6ce0;--el-color-primary-light-5:#5e53c2;--el-color-primary-light-7:#4a418f;--el-color-primary-light-8:#3c356e;--el-color-primary-light-9:#2c2748;--el-bg-color-overlay:#242138;--el-fill-color-light:#2a2544;--el-fill-color:#221e33;--el-fill-color-blank:#1b1829;--el-fill-color-extra-light:#221e33;--el-border-color:#2e2a46;--el-border-color-light:#2e2a46;--el-border-color-lighter:#2e2a46;--el-border-color-extra-light:#28243c;--el-text-color-primary:#e9e7f6;--el-text-color-regular:#cfcbe4;--el-text-color-secondary:#a09bc4;--el-text-color-placeholder:#7d78a3;--el-disabled-bg-color:#221e33;--el-disabled-text-color:#6b6690;--el-disabled-border-color:#2e2a46;--el-mask-color:rgba(10,8,20,.72)}
/* 深色主按钮：提亮后保持白字可读，悬停泛光。 */
:root.dark-theme .el-button--primary:not(.is-disabled):not(.is-link){--el-button-text-color:#f3f1ff}
:root.dark-theme .el-tag{--el-tag-bg-color:#2c2748}
/* 深色指标卡图标：提亮渐变，保持圆形徽章的层次感。 */
:root.dark-theme .metric-icon{background:linear-gradient(135deg,#7a6ce0,#b7adf9);box-shadow:0 6px 16px rgba(141,126,243,.28)}
</style>

<script setup>
// 每次启动固定使用浅色紫主题，不读取历史偏好；会话内仍可通过侧边栏临时切换。
document.documentElement.classList.remove('dark-theme', 'light-theme')
document.documentElement.classList.add('light-theme')
</script>
