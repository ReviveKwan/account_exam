const { login, guestLogin } = require('../../utils/auth')

Page({
  data: {
    agreed: false,
    loading: false
  },

  toggleAgree() {
    this.setData({ agreed: !this.data.agreed })
  },

  async handleLogin() {
    if (!this.data.agreed) {
      wx.showToast({ title: '请先同意用户协议', icon: 'none' })
      return
    }
    if (this.data.loading) return
    this.setData({ loading: true })

    try {
      await login()
      wx.switchTab({ url: '/pages/exam-list/exam-list' })
    } catch (err) {
      console.error('登录失败:', err)
      wx.showToast({ title: '登录失败，请重试', icon: 'none' })
    } finally {
      this.setData({ loading: false })
    }
  },

  async handleGuest() {
    if (this.data.loading) return
    this.setData({ loading: true })
    try {
      await guestLogin()
      wx.switchTab({ url: '/pages/exam-list/exam-list' })
    } catch (err) {
      wx.showToast({ title: '进入失败，请重试', icon: 'none' })
    } finally {
      this.setData({ loading: false })
    }
  }
})
