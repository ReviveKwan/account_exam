const { API } = require('../../utils/request')
const { isLoggedIn, logout } = require('../../utils/auth')

Page({
  data: {
    userInfo: null,
    records: []
  },

  onShow() {
    if (!isLoggedIn()) {
      wx.redirectTo({ url: '/pages/login/login' })
      return
    }
    const app = getApp()
    this.setData({ userInfo: app.globalData.userInfo })
    this.loadRecords()
  },

  async loadRecords() {
    try {
      const data = await API.listRecords(10)
      this.setData({ records: data.records || [] })
    } catch (err) {
      console.error('加载记录失败:', err)
    }
  },

  handleLogout() {
    wx.showModal({
      title: '确认退出',
      content: '退出后需要重新登录',
      success: res => {
        if (res.confirm) {
          logout()
          wx.redirectTo({ url: '/pages/login/login' })
        }
      }
    })
  }
})
