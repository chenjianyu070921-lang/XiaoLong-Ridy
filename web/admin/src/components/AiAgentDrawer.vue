<script setup>
// AI 运营助手抽屉：三个受限场景的运营决策问答。
// 回答固定渲染为四分区（结论/数据证据/优先对象/建议动作），来源模式显式标识，
// 所有建议动作仅做页面跳转，绝不执行写操作。
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ChatDotRound, Delete, Promotion, Position } from '@element-plus/icons-vue'
import { aiApi } from '../api/modules'

const props = defineProps({ modelValue: { type: Boolean, default: false } })
const emit = defineEmits(['update:modelValue'])

const router = useRouter()
const suggestions = ref([])
const messages = ref([])
const question = ref('')
const conversationId = ref('')
const scene = ref('overview')
const loading = ref(false)

const sourceModeText = {
  realtime: '实时数据',
  demo_snapshot: '演示快照',
  template_fallback: '本地规则报告',
}
const sceneText = { overview: '运营总览', abnormal_order: '异常订单', risk_review: '风控审核' }

const sourceModeTag = (mode) => ({ realtime: 'success', demo_snapshot: 'warning', template_fallback: 'info' }[mode] || 'info')

// 近 30 天时间范围，供运营总览等需要时间维度的场景使用。
const timeRange = () => {
  const end = new Date()
  const start = new Date(end)
  start.setDate(end.getDate() - 30)
  const fmt = (d) => d.toISOString().slice(0, 19).replace('T', ' ')
  return { start_time: fmt(start), end_time: fmt(end) }
}

async function loadSuggestions() {
  try {
    const data = await aiApi.suggestions()
    suggestions.value = data?.items || []
  } catch {
    /* 快捷问题加载失败不阻塞手动提问 */
  }
}

async function ask(text) {
  const q = (text || question.value).trim()
  if (!q || loading.value) return
  question.value = ''
  loading.value = true
  try {
    const answer = await aiApi.ask({
      scene: scene.value,
      question: q,
      conversation_id: conversationId.value || undefined,
      ...timeRange(),
    })
    conversationId.value = answer.conversation_id
    messages.value.push({ question: q, answer })
  } catch {
    /* 错误已由 request 拦截器统一提示 */
  } finally {
    loading.value = false
  }
}

function quickAsk(s) {
  scene.value = s.scene
  ask(s.quick_prompt)
}

function navigate(route) {
  if (route) router.push(route)
}

async function feedback(item, helpful) {
  try {
    await aiApi.feedback({
      conversation_id: item.answer.conversation_id,
      trace_id: item.answer.trace_id,
      helpful,
    })
    ElMessage.success('感谢反馈')
  } catch {
    /* 忽略 */
  }
}

async function clearConversation() {
  if (!conversationId.value) return
  try {
    await aiApi.deleteConversation(conversationId.value)
    conversationId.value = ''
    messages.value = []
    ElMessage.success('会话已清空')
  } catch {
    /* 忽略 */
  }
}

function close() {
  emit('update:modelValue', false)
}

onMounted(loadSuggestions)
</script>

<template>
  <el-button
    v-if="!modelValue"
    class="ai-fab"
    type="primary"
    round
    :icon="ChatDotRound"
    @click="emit('update:modelValue', true)"
  >
    AI 运营助手
  </el-button>

  <el-drawer
    :model-value="modelValue"
    :with-header="false"
    size="460px"
    class="ai-drawer"
    @update:model-value="(v) => emit('update:modelValue', v)"
  >
    <div class="ai-wrap">
      <header class="ai-head">
        <div>
          <span class="ai-eyebrow">AI 运营助手</span>
          <h2>运营决策辅助</h2>
        </div>
        <el-button link :icon="Delete" @click="clearConversation" :disabled="!conversationId">清空会话</el-button>
      </header>

      <div class="ai-quick">
        <button v-for="s in suggestions" :key="s.scene" class="ai-quick-item" @click="quickAsk(s)">
          <span class="ai-quick-icon"><el-icon><Position /></el-icon></span>
          <span>{{ sceneText[s.scene] || s.scene }}</span>
          <small>{{ s.quick_prompt }}</small>
        </button>
      </div>

      <div class="ai-messages">
        <el-empty v-if="!messages.length" description="选择上方快捷问题，或输入运营问题开始分析" :image-size="72" />
        <div v-for="(item, idx) in messages" :key="idx" class="ai-msg">
          <div class="ai-ask">{{ item.question }}</div>
          <div class="ai-answer">
            <div class="ai-src"><el-tag :type="sourceModeTag(item.answer.source_mode)" size="small" effect="plain">{{ sourceModeText[item.answer.source_mode] || item.answer.source_mode }}</el-tag></div>

            <section class="ai-sec">
              <h4>结论</h4>
              <p>{{ item.answer.summary || '暂无结论' }}</p>
            </section>

            <section v-if="item.answer.evidence?.length" class="ai-sec">
              <h4>数据证据</h4>
              <div class="ai-evidence">
                <div v-for="ev in item.answer.evidence" :key="ev.label" class="ai-evidence-item">
                  <span>{{ ev.label }}</span>
                  <b>{{ ev.value }}</b>
                  <small v-if="ev.comparison">{{ ev.comparison }}</small>
                </div>
              </div>
            </section>

            <section v-if="item.answer.priorities?.length" class="ai-sec">
              <h4>优先对象</h4>
              <div v-for="p in item.answer.priorities" :key="p.id" class="ai-priority" @click="navigate(p.route)">
                <span class="ai-priority-level" :class="p.level">{{ p.level }}</span>
                <div>
                  <b class="mono">{{ p.id }}</b>
                  <div class="ai-reasons">{{ (p.reasons || []).join('、') }}</div>
                </div>
                <el-button link type="primary" size="small">查看</el-button>
              </div>
            </section>

            <section v-if="item.answer.actions?.length" class="ai-sec">
              <h4>建议动作</h4>
              <div class="ai-actions">
                <el-button v-for="act in item.answer.actions" :key="act.label" size="small" @click="navigate(act.route)">{{ act.label }}</el-button>
              </div>
            </section>

            <footer class="ai-foot">
              <span class="mono">trace: {{ item.answer.trace_id }}</span>
              <span class="ai-feedback">
                <el-button link size="small" @click="feedback(item, true)">有帮助</el-button>
                <el-button link size="small" @click="feedback(item, false)">无帮助</el-button>
              </span>
            </footer>
          </div>
        </div>
      </div>

      <footer class="ai-input">
        <el-select v-model="scene" size="small" class="ai-scene">
          <el-option label="运营总览" value="overview" />
          <el-option label="异常订单" value="abnormal_order" />
          <el-option label="风控审核" value="risk_review" />
        </el-select>
        <el-input
          v-model="question"
          placeholder="输入运营问题，如：分析今日取消率为何上升"
          @keyup.enter="ask()"
        />
        <el-button type="primary" :icon="Promotion" :loading="loading" @click="ask()">发送</el-button>
      </footer>
    </div>
  </el-drawer>
</template>

<style scoped>
.ai-fab{position:fixed;right:28px;bottom:28px;z-index:2000;box-shadow:0 8px 22px rgba(108,92,231,.35)}
.ai-wrap{display:flex;flex-direction:column;height:100%}
.ai-head{display:flex;align-items:flex-start;justify-content:space-between;padding:6px 2px 16px;border-bottom:1px solid var(--border-color,#e5e4f0)}
.ai-eyebrow{color:var(--brand,#6c5ce7);font-size:12px;letter-spacing:.1em;font-weight:600}
.ai-head h2{margin:6px 0 0;font-size:18px;color:var(--text-color,#2e2c4e)}
.ai-quick{display:grid;grid-template-columns:repeat(3,1fr);gap:10px;margin:16px 0}
.ai-quick-item{display:flex;flex-direction:column;align-items:flex-start;gap:6px;padding:12px;border:1px solid var(--border-color,#e5e4f0);border-radius:12px;background:var(--panel-bg,#fff);cursor:pointer;text-align:left;transition:.2s}
.ai-quick-item:hover{border-color:var(--brand,#6c5ce7);box-shadow:0 4px 12px rgba(108,92,231,.12)}
.ai-quick-icon{width:26px;height:26px;border-radius:50%;background:linear-gradient(135deg,var(--brand,#6c5ce7),#9a8ff2);color:#fff;display:flex;align-items:center;justify-content:center;font-size:13px}
.ai-quick-item span{font-weight:600;color:var(--text-color,#2e2c4e);font-size:13px}
.ai-quick-item small{color:var(--muted-color,#8b88a3);font-size:11px;line-height:1.4}
.ai-messages{flex:1;overflow-y:auto;padding:4px 2px}
.ai-msg{margin-bottom:16px}
.ai-ask{background:var(--brand,#6c5ce7);color:#fff;padding:10px 14px;border-radius:12px 12px 12px 4px;font-size:14px;margin-bottom:10px}
.ai-answer{background:var(--panel-bg,#fff);border:1px solid var(--border-color,#e5e4f0);border-radius:12px;padding:14px}
.ai-src{margin-bottom:10px}
.ai-sec{margin-bottom:14px}
.ai-sec h4{margin:0 0 8px;font-size:13px;color:var(--muted-color,#8b88a3);font-weight:600}
.ai-sec p{margin:0;font-size:14px;color:var(--text-color,#2e2c4e);line-height:1.6}
.ai-evidence{display:flex;flex-wrap:wrap;gap:8px}
.ai-evidence-item{flex:1 1 45%;padding:10px;border-radius:10px;background:#f7f6fc}
.ai-evidence-item span{display:block;color:var(--muted-color,#8b88a3);font-size:12px}
.ai-evidence-item b{display:block;margin:4px 0;font-size:16px;color:var(--brand,#6c5ce7)}
.ai-evidence-item small{color:var(--muted-color,#8b88a3);font-size:11px}
.ai-priority{display:flex;align-items:center;gap:10px;padding:10px;border:1px solid var(--border-color,#e5e4f0);border-radius:10px;margin-bottom:8px;cursor:pointer}
.ai-priority:hover{border-color:var(--brand,#6c5ce7)}
.ai-priority-level{text-transform:uppercase;font-size:11px;font-weight:700;padding:3px 8px;border-radius:6px}
.ai-priority-level.high{background:#fde2e2;color:#f56c6c}
.ai-priority-level.medium{background:#fdf3d8;color:#e6a23c}
.ai-priority-level.low{background:#e8f4fd;color:#409eff}
.ai-priority div{flex:1;min-width:0}
.ai-reasons{color:var(--muted-color,#8b88a3);font-size:12px;margin-top:3px}
.ai-actions{display:flex;flex-wrap:wrap;gap:8px}
.ai-foot{display:flex;justify-content:space-between;align-items:center;margin-top:6px;padding-top:10px;border-top:1px dashed var(--border-color,#e5e4f0)}
.ai-foot .mono{font-size:11px;color:var(--muted-color,#8b88a3)}
.ai-input{display:flex;gap:8px;align-items:center;padding-top:12px;border-top:1px solid var(--border-color,#e5e4f0)}
.ai-scene{width:110px;flex:0 0 110px}
.mono{font-family:ui-monospace}
</style>
