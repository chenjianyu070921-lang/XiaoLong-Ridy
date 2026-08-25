<template>
  <div class="profile-page">
    <!-- 用户信息头部 -->
    <div class="profile-header">
      <div class="user-info">
        <div class="avatar" @click="editAvatar">
          <input ref="avatarInput" type="file" accept="image/png,image/jpeg,image/webp" class="avatar-file-input" @change="onAvatarSelected" />
          <img :src="userInfo.avatarUrl || '/default-avatar.png'" alt="" />
          <van-icon name="photograph" class="camera-icon" />
        </div>
        <div class="info">
          <h2 @click="editNickname">{{ userInfo.nickname || '乘客' }}<van-icon name="edit" size="14" class="edit-name-icon" /></h2>
          <p class="phone">{{ userInfo.phone || '138****5678' }}</p>
        </div>
        <button class="edit-btn" @click="goToSettings">
          <van-icon name="setting-o" size="20" />
        </button>
      </div>
    </div>

    <!-- 功能菜单 -->
    <div class="menu-card">
      <div 
        v-for="(item, index) in menuItems" 
        :key="index"
        class="menu-item"
        @click="handleMenuClick(item)"
      >
        <div class="menu-left">
          <component :is="'van-icon'" :name="item.icon" :size="20" :color="item.color" />
          <span>{{ item.label }}</span>
        </div>
        <div class="menu-right">
          <span class="extra" v-if="item.extra">{{ item.extra }}</span>
          <van-icon name="arrow" size="14" color="#9CA3AF" />
        </div>
      </div>
    </div>

    <!-- 其他服务 -->
    <div class="service-card">
      <h3>其他服务</h3>
      <div class="service-grid">
        <div 
          v-for="(item, index) in services" 
          :key="index"
          class="service-item"
          @click="handleServiceClick(item)"
        >
          <van-icon :name="item.icon" :size="24" :color="item.color" />
          <span>{{ item.label }}</span>
        </div>
      </div>
    </div>

    <!-- 资料编辑弹窗：使用明确的输入框和操作按钮，避免系统对话框内容拥挤。 -->
    <div v-if="dialogType" class="edit-dialog-mask" @click.self="closeEditDialog">
      <section class="edit-dialog" role="dialog" aria-modal="true">
        <div class="edit-dialog-head">
          <h3>{{ dialogType === 'nickname' ? '修改昵称' : '修改头像' }}</h3>
          <button type="button" class="edit-dialog-close" @click="closeEditDialog">×</button>
        </div>
        <p class="edit-dialog-hint">{{ dialogType === 'nickname' ? '设置一个容易记住的昵称' : '选择本地图片作为头像' }}</p>
        <input v-if="dialogType === 'nickname'" v-model="dialogValue" class="edit-dialog-input" maxlength="20" placeholder="请输入昵称" autofocus @keyup.enter="saveEditDialog" />
        <div v-else class="avatar-upload-preview">
          <img :src="dialogValue" alt="头像预览" />
          <button type="button" class="choose-avatar-btn" @click="chooseAvatar">重新选择图片</button>
        </div>
        <div class="edit-dialog-actions">
          <button type="button" class="edit-dialog-cancel" @click="closeEditDialog">取消</button>
          <button type="button" class="edit-dialog-save" :disabled="dialogLoading" @click="saveEditDialog">{{ dialogLoading ? '保存中...' : '保存' }}</button>
        </div>
      </section>
    </div>
    <!-- 退出登录按钮 -->
    <button class="logout-btn safe-area-bottom" @click="handleLogout">
      退出登录
    </button>

    <!-- 底部导航栏 -->
    <van-tabbar v-model="activeNav" active-color="#7C3AED" inactive-color="#6B7280">
      <van-tabbar-item icon="home-o" name="home" @click="goHome">首页</van-tabbar-item>
      <van-tabbar-item icon="orders-o" name="orders" @click="goOrders">订单</van-tabbar-item>
      <van-tabbar-item icon="user-o" name="profile">我的</van-tabbar-item>
    </van-tabbar>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { showToast, showDialog } from 'vant'
import { useUserStore } from '@/stores/user'
import { updateProfile } from '@/api/user'
import { getMyCoupons } from '@/api/order'

const router = useRouter()
const userStore = useUserStore()

// 状态
const activeNav = ref('profile')
const dialogType = ref('')
const dialogValue = ref('')
const dialogLoading = ref(false)
const avatarInput = ref(null)

// 用户信息
const userInfo = computed(() => userStore.userInfo)

// 统计数据
const couponCount = ref(0)

const stats = ref({
  totalTrips: 28,
  totalDistance: 156.8,
  totalSpent: 682.50
})

// 钱包信息
const wallet = ref({
  balance: '68.0',
  frozen: '0.00'
})

// 菜单项
const menuItems = ref([
  { label: '我的钱包', icon: 'balance-o', color: '#7C3AED', action: 'wallet', extra: '' },
  { label: '优惠券', icon: 'coupon-o', color: '#F59E0B', action: 'coupons', extra: '加载中' },
  { label: '行程记录', icon: 'todo-list-o', color: '#10B981', action: 'trips', extra: '' },
  { label: '常用地址', icon: 'location-o', color: '#3B82F6', action: 'addresses', extra: '' },
  { label: '客服中心', icon: 'service-o', color: '#8B5CF6', action: 'service', extra: '' },
  { label: '设置', icon: 'setting-o', color: '#6B7280', action: 'settings', extra: '' }
])

// 服务项
const services = ref([
  { label: '邀请有礼', icon: 'gift-o', color: '#EF4444' },
  { label: '企业出行', icon: 'cluster-o', color: '#3B82F6' },
  { label: '安全中心', icon: 'shield-o', color: '#10B981' },
  { label: '帮助反馈', icon: 'question-o', color: '#F59E0B' }
])

// 方法

// 打开修改昵称弹窗，并预填当前昵称，避免使用无输入框的系统对话框造成布局错乱。
const editNickname = () => {
  dialogType.value = 'nickname'
  dialogValue.value = userInfo.value.nickname || ''
}

// 打开修改头像弹窗，头像字段只保存图片地址。
const editAvatar = () => {
  dialogType.value = 'avatar'
  dialogValue.value = userInfo.value.avatarUrl || ''
}

// 关闭编辑弹窗并清理输入内容。
const closeEditDialog = () => {
  if (!dialogLoading.value) {
    dialogType.value = ''
    dialogValue.value = ''
  }
}

// 保存昵称或头像，并同步 Pinia 与本地缓存，确保刷新后仍能显示最新资料。
const saveEditDialog = async () => {
  const value = dialogValue.value.trim()
  if (!value) {
    showToast(dialogType.value === 'nickname' ? '昵称不能为空' : '头像地址不能为空')
    return
  }
  if (dialogType.value === 'nickname' && [...value].length > 20) {
    showToast('昵称不能超过 20 个字')
    return
  }
  dialogLoading.value = true
  try {
    const payload = dialogType.value === 'nickname' ? { nickname: value } : { avatarUrl: value }
    await updateProfile(payload)
    userStore.userInfo = { ...userStore.userInfo, ...(dialogType.value === 'nickname' ? { nickname: value } : { avatarUrl: value }) }
    localStorage.setItem('userInfo', JSON.stringify(userStore.userInfo))
    showToast(dialogType.value === 'nickname' ? '昵称已更新' : '头像已更新')
    closeEditDialog()
  } catch (error) {
    showToast('更新失败，请重试')
  } finally {
    dialogLoading.value = false
  }
}
const goToSettings = () => router.push('/settings')

const goToWallet = () => router.push('/wallet')

const handleMenuClick = (item) => {
  const actions = {
    wallet: () => goToWallet(),
    coupons: () => router.push('/coupons'),
    trips: () => router.push('/orders'),
    addresses: () => showToast('地址管理功能开发中'),
    service: () => showToast('客服电话：400-123-4567'),
    settings: () => goToSettings()
  }
  
  if (actions[item.action]) {
    actions[item.action]()
  }
}

const handleServiceClick = (item) => {
  showToast(`${item.label}功能开发中`)
}

const goHome = () => router.replace('/home')
const goOrders = () => router.replace('/orders')

// 退出登录
const handleLogout = () => {
  showDialog({
    title: '提示',
    message: '确定要退出登录吗？',
    showCancelButton: true,
    confirmButtonText: '确定退出',
    cancelButtonText: '取消'
  }).then(() => {
    userStore.logout()
    showToast('已退出登录')
    router.replace('/splash')
  }).catch(() => {})
}

onMounted(async () => {
  // 从后端加载可用券数量，个人中心不再展示写死的优惠券数量。
  try {
    const result = await getMyCoupons(1)
    couponCount.value = Array.isArray(result?.list) ? result.list.length : 0
    const couponItem = menuItems.value.find(item => item.action === 'coupons')
    if (couponItem) couponItem.extra = `${couponCount.value}张可用`
  } catch (error) {
    const couponItem = menuItems.value.find(item => item.action === 'coupons')
    if (couponItem) couponItem.extra = '查看'
  }
})
</script>

<style scoped>
.profile-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 80px;
}

.profile-header {
  background: linear-gradient(135deg, #7C3AED 0%, #9333EA 100%);
  padding: 40px 20px 24px;
  color: white;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 24px;
}

.avatar {
  position: relative;
  width: 64px;
  height: 64px;
  cursor: pointer;
}

.avatar-file-input { position: absolute; width: 1px; height: 1px; opacity: 0; pointer-events: none; }
.avatar img {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  object-fit: cover;
  border: 3px solid rgba(255, 255, 255, 0.3);
  background: rgba(255, 255, 255, 0.2);
}

.camera-icon {
  position: absolute;
  bottom: 0;
  right: 0;
  background: white;
  color: #7C3AED !important;
  padding: 4px;
  border-radius: 50%;
  font-size: 12px;
}

.info h2 {
  font-size: 20px;
  font-weight: 600;
  margin-bottom: 4px;
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
}

.edit-name-icon {
  color: rgba(255, 255, 255, 0.8);
}

.phone {
  font-size: 13px;
  opacity: 0.9;
}

.edit-btn {
  margin-left: auto;
  width: 36px;
  height: 36px;
  background: rgba(255, 255, 255, 0.2);
  border: none;
  border-radius: 50%;
  color: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.stats {
  display: flex;
  align-items: center;
  justify-content: space-around;
  background: rgba(255, 255, 255, 0.15);
  border-radius: 12px;
  padding: 16px;
}

.stat-item {
  text-align: center;
}

.stat-item .value {
  display: block;
  font-size: 18px;
  font-weight: 700;
  margin-bottom: 4px;
}

.stat-item .label {
  font-size: 12px;
  opacity: 0.9;
}

.stat-divider {
  width: 1px;
  height: 32px;
  background: rgba(255, 255, 255, 0.2);
}

.wallet-card {
  margin: -20px 16px 12px;
  background: linear-gradient(135deg, #7C3AED 0%, #A78BFA 100%);
  border-radius: 12px;
  padding: 18px 16px;
  display: flex;
  align-items: center;
  gap: 14px;
  color: white;
  box-shadow: 0 4px 15px rgba(124, 58, 237, 0.3);
  cursor: pointer;
  transition: all 0.2s;
}

.wallet-card:active {
  transform: scale(0.98);
}

.wallet-icon {
  width: 48px;
  height: 48px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.wallet-info {
  flex: 1;
}

.wallet-info .label {
  font-size: 13px;
  opacity: 0.9;
  margin-bottom: 4px;
}

.balance {
  font-size: 22px;
  font-weight: 700;
}

.wallet-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.recharge-btn {
  padding: 6px 16px;
  background: white;
  color: #7C3AED;
  border: none;
  border-radius: 16px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
}

.menu-card,
.service-card {
  margin: 0 16px 12px;
  background: white;
  border-radius: 12px;
  overflow: hidden;
}

.service-card h3 {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
  padding: 16px 16px 12px;
}

.menu-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  border-bottom: 1px solid #F3F4F6;
  cursor: pointer;
  transition: background 0.2s;
}

.menu-item:last-child {
  border-bottom: none;
}

.menu-item:active {
  background: #F9FAFB;
}

.menu-left {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 15px;
  color: var(--text-primary);
}

.menu-right {
  display: flex;
  align-items: center;
  gap: 6px;
}

.extra {
  font-size: 13px;
  color: #9CA3AF;
}

.service-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  padding: 8px 0 16px;
}

.service-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 12px 0;
  cursor: pointer;
}

.service-item span {
  font-size: 12px;
  color: #6B7280;
}

.logout-btn {
  width: calc(100% - 32px);
  margin: 20px 16px 30px;
  height: 48px;
  background: white;
  border: 1px solid #D1D5DB;
  border-radius: 24px;
  font-size: 16px;
  color: #DC2626;
  cursor: pointer;
  transition: all 0.2s;
}

.logout-btn:active {
  background: #FEF2F2;
}

/* 资料编辑弹窗样式：使用半透明遮罩、紧凑输入区和清晰的主次按钮。 */
.edit-dialog-mask { position: fixed; inset: 0; z-index: 100; display: flex; align-items: center; justify-content: center; padding: 24px; background: rgba(15, 23, 42, 0.56); }
.edit-dialog { width: min(100%, 340px); padding: 22px; border-radius: 20px; background: #fff; box-shadow: 0 20px 60px rgba(15, 23, 42, 0.24); }
.edit-dialog-head { display: flex; align-items: center; justify-content: space-between; }
.edit-dialog-head h3 { color: #111827; font-size: 18px; font-weight: 700; }
.edit-dialog-close { width: 30px; height: 30px; border: 0; border-radius: 50%; background: #f3f4f6; color: #6b7280; font-size: 22px; line-height: 1; }
.edit-dialog-hint { margin: 10px 0 16px; color: #9ca3af; font-size: 13px; line-height: 1.5; }
.edit-dialog-input { width: 100%; height: 46px; padding: 0 14px; border: 1px solid #e5e7eb; border-radius: 12px; outline: none; color: #111827; background: #f9fafb; font-size: 15px; }
.avatar-upload-preview { display: flex; align-items: center; flex-direction: column; gap: 14px; }

.avatar-upload-preview img { width: 92px; height: 92px; border-radius: 50%; object-fit: cover; border: 3px solid #ede9fe; }

.choose-avatar-btn { border: 0; border-radius: 10px; padding: 9px 14px; color: #6d28d9; background: #f5f3ff; font-size: 13px; }


.edit-dialog-input:focus { border-color: #8b5cf6; background: #fff; box-shadow: 0 0 0 3px rgba(139, 92, 246, 0.12); }
.edit-dialog-actions { display: flex; gap: 10px; margin-top: 18px; }
.edit-dialog-actions button { flex: 1; height: 42px; border-radius: 12px; font-size: 14px; font-weight: 600; }
.edit-dialog-cancel { border: 1px solid #e5e7eb; background: #fff; color: #6b7280; }
.edit-dialog-save { border: 0; background: #7c3aed; color: #fff; }
.edit-dialog-save:disabled { opacity: .55; }
</style>


