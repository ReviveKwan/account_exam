const { API } = require('../../utils/request')
const { isLoggedIn } = require('../../utils/auth')
const { formatTime } = require('../../utils/exam-state')

const CN_NUMS = ['','一','二','三','四','五','六','七','八','九','十','十一','十二','十三','十四','十五','十六','十七','十八']
const BAR_COLORS = { 7: '#4F46E5', 8: '#059669', 13: '#E11D48' }

Page({
  data: {
    userInfo: null,
    chapters: [],
    dashboard: null,
    mode: 1, // 1=模考 2=背题
    loading: true
  },

  onShow() {
    if (!isLoggedIn()) {
      wx.redirectTo({ url: '/pages/login/login' })
      return
    }
    const app = getApp()
    this.setData({ userInfo: app.globalData.userInfo })
    this.loadData()
  },

  async loadData() {
    try {
      const data = await API.listChapters()
      const chapters = (data.chapters || []).map(ch => {
        // 预计算 badges（WXML 不支持 startsWith）
        const badges = (ch.badges || []).map(badge => {
          if (typeof badge === 'string' && badge.indexOf('pass') === 0) {
            return { type: 'pass', text: '最高 ' + ch.best_score + ' 分' }
          }
          if (badge === 'hot') return { type: 'hot', text: '热门高频' }
          if (badge === 'new') return { type: 'new', text: '2025新修' }
          return { type: '', text: badge }
        })
        return {
          ...ch,
          cnNo: CN_NUMS[ch.chapter_no] || ch.chapter_no,
          barColor: BAR_COLORS[ch.chapter_no] || '#4F46E5',
          badges,
          resumeRemainingText: ch.resume_remaining ? formatTime(ch.resume_remaining) : ''
        }
      })
      this.setData({
        chapters,
        dashboard: data.dashboard,
        loading: false
      })
    } catch (err) {
      console.error('加载失败:', err)
      this.setData({ loading: false })
      wx.showToast({ title: '加载失败', icon: 'none' })
    }
  },

  switchMode(e) {
    this.setData({ mode: parseInt(e.currentTarget.dataset.mode) })
  },

  async startExam(e) {
    const chapterId = e.currentTarget.dataset.chapterId
    const examId = e.currentTarget.dataset.examId
    wx.navigateTo({
      url: `/pages/exam-page/exam-page?examId=${examId}&mode=${this.data.mode}&chapterId=${chapterId}`
    })
  },

  resumeExam(e) {
    const examId = e.currentTarget.dataset.examId
    wx.navigateTo({
      url: `/pages/exam-page/exam-page?examId=${examId}&mode=1&resume=1`
    })
  }
})
