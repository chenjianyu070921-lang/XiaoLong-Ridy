<template>
  <router-view />
</template>

<style>
/* 全局主题基线：默认深色，主题类由 App 启动时根据本地偏好设置。 */
:root{font-family:Inter,"PingFang SC","Microsoft YaHei",sans-serif;color:#dce7f3;background:#070e17;font-synthesis:none;text-rendering:optimizeLegibility}
:root.light-theme{color:#243447;background:#f4f7fb}
*{box-sizing:border-box}body{margin:0;min-width:1100px;background:var(--page-bg,#070e17);color:var(--text-color,#dce7f3);transition:background-color .2s,color .2s}button,input{font:inherit}
.el-table{--el-table-bg-color:var(--panel-bg,#101d2b);--el-table-tr-bg-color:var(--panel-bg,#101d2b);--el-table-header-bg-color:var(--header-bg,#0c1724);--el-table-row-hover-bg-color:var(--hover-bg,#162839);--el-table-border-color:var(--border-color,#1d3042);--el-table-text-color:var(--text-color,#c7d5e2);--el-table-header-text-color:var(--muted-color,#8295a8)}
.el-table th.el-table__cell{font-weight:600}.el-input__wrapper,.el-select__wrapper,.el-textarea__inner{background:var(--input-bg,#0d1a28);box-shadow:0 0 0 1px var(--border-color,#21364a) inset;color:var(--text-color,#dce7f3)}.el-input__inner{color:var(--text-color,#dce7f3)}.el-pagination{--el-pagination-bg-color:var(--panel-bg,#101d2b);--el-pagination-text-color:var(--muted-color,#93a7ba);--el-pagination-button-color:var(--text-color,#dce7f3)}.el-dialog{background:var(--panel-bg,#101d2b);border:1px solid var(--border-color,#294056)}.el-dialog__title,.el-form-item__label{color:var(--text-color,#dce7f3)}.el-menu-item{height:45px;line-height:45px}.el-menu-item.is-active{background:var(--active-bg,#1d2d3d)!important;color:#ff8538!important;border-left:3px solid #ff7625}.el-button--primary{--el-button-bg-color:#ff7625;--el-button-border-color:#ff7625;--el-button-hover-bg-color:#ff944d}
:root.dark-theme{--page-bg:#070e17;--panel-bg:#101d2b;--header-bg:#0c1724;--input-bg:#0d1a28;--hover-bg:#162839;--active-bg:#1d2d3d;--border-color:#1d3042;--text-color:#dce7f3;--muted-color:#8295a8}
:root.light-theme{--page-bg:#f4f7fb;--panel-bg:#ffffff;--header-bg:#ffffff;--input-bg:#ffffff;--hover-bg:#eef4fa;--active-bg:#e8f1fa;--border-color:#d8e1ea;--text-color:#243447;--muted-color:#718096}
</style>

<script setup>
// 启动时恢复主题偏好，确保登录页、注册页和后台页面统一显示。
const savedTheme = localStorage.getItem('admin-theme') || 'dark'
document.documentElement.classList.remove('dark-theme', 'light-theme')
document.documentElement.classList.add(`${savedTheme === 'light' ? 'light' : 'dark'}-theme`)
</script>
