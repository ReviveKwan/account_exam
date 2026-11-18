const { API } = require('../../utils/request')

Page({
  data: {
    recordId: null,
    record: null,
    answers: [],
    filterTab: 'all', // all / wrong / marked
    filteredDetails: [],
    loading: true
  },

  onLoad(options) {
    // recordId 是字符串（如 rec_xxx），不能 parseInt
    const recordId = decodeURIComponent(options.recordId || '')
    if (recordId) {
      this.setData({ recordId })
      this.loadResult(recordId)
    }
  },

  async loadResult(recordId) {
    try {
      const data = await API.getRecord(recordId)
      const record = data.record
      const answers = data.answers || []

      // 预计算时长（WXML 不支持 Math.floor）
      const sec = record.duration_seconds || 0
      record.durationText = Math.floor(sec / 60) + '分' + (sec % 60) + '秒'

      this.setData({
        record,
        answers,
        filteredDetails: answers,
        loading: false
      })
    } catch (err) {
      console.error('加载结果失败:', err)
      wx.showToast({ title: '加载失败', icon: 'none' })
      this.setData({ loading: false })
    }
  },

  switchFilter(e) {
    const tab = e.currentTarget.dataset.tab
    let filtered = this.data.answers
    if (tab === 'wrong') {
      filtered = this.data.answers.filter(a => a.is_correct === 0 && a.is_answered === 1)
    } else if (tab === 'marked') {
      filtered = this.data.answers.filter(a => a.is_marked === 1)
    }
    this.setData({ filterTab: tab, filteredDetails: filtered })
  },

  retake() {
    wx.navigateBack()
  },

  practiceWrong() {
    wx.switchTab({ url: '/pages/wrong-book/wrong-book' })
  }
})
