<script setup>
// 业务列表页面壳：按路由域隔离页面入口，复用已验证的列表、筛选和表单工作区。
// 具体业务通过当前路由 path 选择 API 和字段，后续可在不改路由契约的情况下逐域细化。
import ConsoleWorkspace from '../console/index.vue'
</script>

<template>
  <ConsoleWorkspace />
</template>
